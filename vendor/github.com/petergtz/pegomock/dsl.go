// Copyright 2015 Peter Goetz
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pegomock

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/onsi/gomega/format"
	"github.com/petergtz/pegomock/internal/verify"
)

var GlobalFailHandler FailHandler

func RegisterMockFailHandler(handler FailHandler) {
	GlobalFailHandler = handler
}
func RegisterMockTestingT(t *testing.T) {
	RegisterMockFailHandler(BuildTestingTGomegaFailHandler(t))
}

var (
	lastInvocation      *invocation
	lastInvocationMutex sync.Mutex
)

var globalArgMatchers Matchers

func RegisterMatcher(matcher Matcher) {
	globalArgMatchers.append(matcher)
}

type invocation struct {
	genericMock *GenericMock
	MethodName  string
	Params      []Param
	ReturnTypes []reflect.Type
}

type GenericMock struct {
	sync.Mutex
	mockedMethods map[string]*mockedMethod
}

func (genericMock *GenericMock) Invoke(methodName string, params []Param, returnTypes []reflect.Type) ReturnValues {
	lastInvocationMutex.Lock()
	lastInvocation = &invocation{
		genericMock: genericMock,
		MethodName:  methodName,
		Params:      params,
		ReturnTypes: returnTypes,
	}
	lastInvocationMutex.Unlock()
	return genericMock.getOrCreateMockedMethod(methodName).Invoke(params)
}

func (genericMock *GenericMock) stub(methodName string, paramMatchers []Matcher, returnValues ReturnValues) {
	genericMock.stubWithCallback(methodName, paramMatchers, func([]Param) ReturnValues { return returnValues })
}

func (genericMock *GenericMock) stubWithCallback(methodName string, paramMatchers []Matcher, callback func([]Param) ReturnValues) {
	genericMock.getOrCreateMockedMethod(methodName).stub(paramMatchers, callback)
}

func (genericMock *GenericMock) getOrCreateMockedMethod(methodName string) *mockedMethod {
	genericMock.Lock()
	defer genericMock.Unlock()
	if _, ok := genericMock.mockedMethods[methodName]; !ok {
		genericMock.mockedMethods[methodName] = &mockedMethod{name: methodName}
	}
	return genericMock.mockedMethods[methodName]
}

func (genericMock *GenericMock) reset(methodName string, paramMatchers []Matcher) {
	genericMock.getOrCreateMockedMethod(methodName).reset(paramMatchers)
}

func (genericMock *GenericMock) Verify(
	inOrderContext *InOrderContext,
	invocationCountMatcher Matcher,
	methodName string,
	params []Param,
	options ...interface{},
) []MethodInvocation {
	var timeout time.Duration
	if len(options) == 1 {
		timeout = options[0].(time.Duration)
	}
	if GlobalFailHandler == nil {
		panic("No GlobalFailHandler set. Please use either RegisterMockFailHandler or RegisterMockTestingT to set a fail handler.")
	}
	defer func() { globalArgMatchers = nil }() // We don't want a panic somewhere during verification screw our global argMatchers

	if len(globalArgMatchers) != 0 {
		verifyArgMatcherUse(globalArgMatchers, params)
	}
	startTime := time.Now()
	// timeoutLoop:
	for {
		genericMock.Lock()
		methodInvocations := genericMock.methodInvocations(methodName, params, globalArgMatchers)
		genericMock.Unlock()
		if inOrderContext != nil {
			for _, methodInvocation := range methodInvocations {
				if methodInvocation.orderingInvocationNumber <= inOrderContext.invocationCounter {
					// TODO: should introduce the following, in case we decide support "inorder" and "eventually"
					// if time.Since(startTime) < timeout {
					// 	continue timeoutLoop
					// }
					GlobalFailHandler(fmt.Sprintf("Expected function call %v(%v) before function call %v(%v)",
						methodName, formatParams(params), inOrderContext.lastInvokedMethodName, formatParams(inOrderContext.lastInvokedMethodParams)))
				}
				inOrderContext.invocationCounter = methodInvocation.orderingInvocationNumber
				inOrderContext.lastInvokedMethodName = methodName
				inOrderContext.lastInvokedMethodParams = params
			}
		}
		if !invocationCountMatcher.Matches(len(methodInvocations)) {
			if time.Since(startTime) < timeout {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			var paramsOrMatchers interface{} = formatParams(params)
			if len(globalArgMatchers) != 0 {
				paramsOrMatchers = formatMatchers(globalArgMatchers)
			}
			timeoutInfo := ""
			if timeout > 0 {
				timeoutInfo = fmt.Sprintf(" after timeout of %v", timeout)
			}
			GlobalFailHandler(fmt.Sprintf(
				"Mock invocation count for %v(%v) does not match expectation%v.\n\n\t%v\n\n\t%v",
				methodName, paramsOrMatchers, timeoutInfo, invocationCountMatcher.FailureMessage(), formatInteractions(genericMock.allInteractions())))
		}
		return methodInvocations
	}
}

// TODO this doesn't need to be a method, can be a free function
func (genericMock *GenericMock) GetInvocationParams(methodInvocations []MethodInvocation) [][]Param {
	if len(methodInvocations) == 0 {
		return nil
	}
	result := make([][]Param, len(methodInvocations[len(methodInvocations)-1].params))
	for i, invocation := range methodInvocations {
		for u, param := range invocation.params {
			if result[u] == nil {
				result[u] = make([]Param, len(methodInvocations))
			}
			result[u][i] = param
		}
	}
	return result
}

func (genericMock *GenericMock) methodInvocations(methodName string, params []Param, matchers []Matcher) []MethodInvocation {
	var invocations []MethodInvocation
	if method, exists := genericMock.mockedMethods[methodName]; exists {
		method.Lock()
		for _, invocation := range method.invocations {
			if len(matchers) != 0 {
				if Matchers(matchers).Matches(invocation.params) {
					invocations = append(invocations, invocation)
				}
			} else {
				if reflect.DeepEqual(params, invocation.params) ||
					(len(params) == 0 && len(invocation.params) == 0) {
					invocations = append(invocations, invocation)
				}
			}
		}
		method.Unlock()
	}
	return invocations
}

func formatInteractions(interactions map[string][]MethodInvocation) string {
	if len(interactions) == 0 {
		return "There were no other interactions with this mock"
	}
	result := "But other interactions with this mock were:\n"
	for _, methodName := range sortedMethodNames(interactions) {
		result += formatInvocations(methodName, interactions[methodName])
	}
	return result
}

func formatInvocations(methodName string, invocations []MethodInvocation) (result string) {
	for _, invocation := range invocations {
		result += "\t" + methodName + "(" + formatParams(invocation.params) + ")\n"
	}
	return
}

func formatParams(params []Param) (result string) {
	for i, param := range params {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("%#v", param)
	}
	return
}

func formatMatchers(matchers []Matcher) (result string) {
	for i, matcher := range matchers {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("%v", matcher)
	}
	return
}

func sortedMethodNames(interactions map[string][]MethodInvocation) []string {
	methodNames := make([]string, len(interactions))
	i := 0
	for key := range interactions {
		methodNames[i] = key
		i++
	}
	sort.Strings(methodNames)
	return methodNames
}

func (genericMock *GenericMock) allInteractions() map[string][]MethodInvocation {
	interactions := make(map[string][]MethodInvocation)
	for methodName := range genericMock.mockedMethods {
		for _, invocation := range genericMock.mockedMethods[methodName].invocations {
			interactions[methodName] = append(interactions[methodName], invocation)
		}
	}
	return interactions
}

type mockedMethod struct {
	sync.Mutex
	name        string
	invocations []MethodInvocation
	stubbings   Stubbings
}

func (method *mockedMethod) Invoke(params []Param) ReturnValues {
	method.Lock()
	method.invocations = append(method.invocations, MethodInvocation{params, globalInvocationCounter.nextNumber()})
	method.Unlock()
	stubbing := method.stubbings.find(params)
	if stubbing == nil {
		return ReturnValues{}
	}
	return stubbing.Invoke(params)
}

func (method *mockedMethod) stub(paramMatchers Matchers, callback func([]Param) ReturnValues) {
	stubbing := method.stubbings.findByMatchers(paramMatchers)
	if stubbing == nil {
		stubbing = &Stubbing{paramMatchers: paramMatchers}
		method.stubbings = append(method.stubbings, stubbing)
	}
	stubbing.callbackSequence = append(stubbing.callbackSequence, callback)
}

func (method *mockedMethod) removeLastInvocation() {
	method.invocations = method.invocations[:len(method.invocations)-1]
}

func (method *mockedMethod) reset(paramMatchers Matchers) {
	method.stubbings.removeByMatchers(paramMatchers)
}

type Counter struct {
	count int
	sync.Mutex
}

func (counter *Counter) nextNumber() (nextNumber int) {
	counter.Lock()
	defer counter.Unlock()

	nextNumber = counter.count
	counter.count++
	return
}

var globalInvocationCounter = Counter{count: 1}

type MethodInvocation struct {
	params                   []Param
	orderingInvocationNumber int
}

type Stubbings []*Stubbing

func (stubbings Stubbings) find(params []Param) *Stubbing {
	for i := len(stubbings) - 1; i >= 0; i-- {
		if stubbings[i].paramMatchers.Matches(params) {
			return stubbings[i]
		}
	}
	return nil
}

func (stubbings Stubbings) findByMatchers(paramMatchers Matchers) *Stubbing {
	for _, stubbing := range stubbings {
		if matchersEqual(stubbing.paramMatchers, paramMatchers) {
			return stubbing
		}
	}
	return nil
}

func (stubbings *Stubbings) removeByMatchers(paramMatchers Matchers) {
	for i, stubbing := range *stubbings {
		if matchersEqual(stubbing.paramMatchers, paramMatchers) {
			*stubbings = append((*stubbings)[:i], (*stubbings)[i+1:]...)
		}
	}
}

func matchersEqual(a, b Matchers) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !reflect.DeepEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}

type Stubbing struct {
	paramMatchers    Matchers
	callbackSequence []func([]Param) ReturnValues
	sequencePointer  int
}

func (stubbing *Stubbing) Invoke(params []Param) ReturnValues {
	defer func() {
		if stubbing.sequencePointer < len(stubbing.callbackSequence)-1 {
			stubbing.sequencePointer++
		}
	}()
	return stubbing.callbackSequence[stubbing.sequencePointer](params)
}

type Matchers []Matcher

func (matchers Matchers) Matches(params []Param) bool {
	if len(matchers) != len(params) { // Technically, this is not an error. Variadic arguments can cause this
		return false
	}

	for i := range params {
		if !matchers[i].Matches(params[i]) {
			return false
		}
	}
	return true
}

func (matchers *Matchers) append(matcher Matcher) {
	*matchers = append(*matchers, matcher)
}

type ongoingStubbing struct {
	genericMock   *GenericMock
	MethodName    string
	ParamMatchers []Matcher
	returnTypes   []reflect.Type
}

func When(invocation ...interface{}) *ongoingStubbing {
	callIfIsFunc(invocation)
	verify.Argument(lastInvocation != nil,
		"When() requires an argument which has to be 'a method call on a mock'.")
	defer func() {
		lastInvocationMutex.Lock()
		lastInvocation = nil
		lastInvocationMutex.Unlock()

		globalArgMatchers = nil
	}()
	lastInvocation.genericMock.mockedMethods[lastInvocation.MethodName].removeLastInvocation()

	paramMatchers := paramMatchersFromArgMatchersOrParams(globalArgMatchers, lastInvocation.Params)
	lastInvocation.genericMock.reset(lastInvocation.MethodName, paramMatchers)
	return &ongoingStubbing{
		genericMock:   lastInvocation.genericMock,
		MethodName:    lastInvocation.MethodName,
		ParamMatchers: paramMatchers,
		returnTypes:   lastInvocation.ReturnTypes,
	}
}

func callIfIsFunc(invocation []interface{}) {
	if len(invocation) == 1 {
		actualType := actualTypeOf(invocation[0])
		if actualType != nil && actualType.Kind() == reflect.Func && !reflect.ValueOf(invocation[0]).IsNil() {
			if !(actualType.NumIn() == 0 && actualType.NumOut() == 0) {
				panic("When using 'When' with function that does not return a value, " +
					"it expects a function with no arguments and no return value.")
			}
			reflect.ValueOf(invocation[0]).Call([]reflect.Value{})
		}
	}
}

// Deals with nils without panicking
func actualTypeOf(iface interface{}) reflect.Type {
	defer func() { recover() }()
	return reflect.TypeOf(iface)
}

func paramMatchersFromArgMatchersOrParams(argMatchers []Matcher, params []Param) []Matcher {
	if len(argMatchers) != 0 {
		verifyArgMatcherUse(argMatchers, params)
		return argMatchers
	}
	return transformParamsIntoEqMatchers(params)
}

func verifyArgMatcherUse(argMatchers []Matcher, params []Param) {
	verify.Argument(len(argMatchers) == len(params),
		"Invalid use of matchers!\n\n %v matchers expected, %v recorded.\n\n"+
			"This error may occur if matchers are combined with raw values:\n"+
			"    //incorrect:\n"+
			"    someFunc(AnyInt(), \"raw String\")\n"+
			"When using matchers, all arguments have to be provided by matchers.\n"+
			"For example:\n"+
			"    //correct:\n"+
			"    someFunc(AnyInt(), EqString(\"String by matcher\"))",
		len(params), len(argMatchers),
	)
}

func transformParamsIntoEqMatchers(params []Param) []Matcher {
	paramMatchers := make([]Matcher, len(params))
	for i, param := range params {
		paramMatchers[i] = &EqMatcher{Value: param}
	}
	return paramMatchers
}

var (
	genericMocksMutex sync.Mutex
	genericMocks      = make(map[Mock]*GenericMock)
)

func GetGenericMockFrom(mock Mock) *GenericMock {
	genericMocksMutex.Lock()
	defer genericMocksMutex.Unlock()
	if genericMocks[mock] == nil {
		genericMocks[mock] = &GenericMock{mockedMethods: make(map[string]*mockedMethod)}
	}
	return genericMocks[mock]
}

func (stubbing *ongoingStubbing) ThenReturn(values ...ReturnValue) *ongoingStubbing {
	checkAssignabilityOf(values, stubbing.returnTypes)
	stubbing.genericMock.stub(stubbing.MethodName, stubbing.ParamMatchers, values)
	return stubbing
}

func checkAssignabilityOf(stubbedReturnValues []ReturnValue, expectedReturnTypes []reflect.Type) {
	verify.Argument(len(stubbedReturnValues) == len(expectedReturnTypes),
		"Different number of return values")
	for i := range stubbedReturnValues {
		if stubbedReturnValues[i] == nil {
			switch expectedReturnTypes[i].Kind() {
			case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint,
				reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32,
				reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.Array, reflect.String,
				reflect.Struct:
				panic("Return value 'nil' not assignable to return type " + expectedReturnTypes[i].Kind().String())
			}
		} else {
			verify.Argument(reflect.TypeOf(stubbedReturnValues[i]).AssignableTo(expectedReturnTypes[i]),
				"Return value of type %T not assignable to return type %v", stubbedReturnValues[i], expectedReturnTypes[i])
		}
	}
}

func (stubbing *ongoingStubbing) ThenPanic(v interface{}) *ongoingStubbing {
	stubbing.genericMock.stubWithCallback(
		stubbing.MethodName,
		stubbing.ParamMatchers,
		func([]Param) ReturnValues { panic(v) })
	return stubbing
}

func (stubbing *ongoingStubbing) Then(callback func([]Param) ReturnValues) *ongoingStubbing {
	stubbing.genericMock.stubWithCallback(
		stubbing.MethodName,
		stubbing.ParamMatchers,
		callback)
	return stubbing
}

type InOrderContext struct {
	invocationCounter       int
	lastInvokedMethodName   string
	lastInvokedMethodParams []Param
}

// Matcher ... it is guaranteed that FailureMessage will always be called after Matches
// so an implementation can save state
type Matcher interface {
	Matches(param Param) bool
	FailureMessage() string
	fmt.Stringer
}

func DumpInvocationsFor(mock Mock) {
	fmt.Print(SDumpInvocationsFor(mock))
}

func SDumpInvocationsFor(mock Mock) string {
	result := &bytes.Buffer{}
	for _, mockedMethod := range GetGenericMockFrom(mock).mockedMethods {
		for _, invocation := range mockedMethod.invocations {
			fmt.Fprintf(result, "Method invocation: %v (\n", mockedMethod.name)
			for _, param := range invocation.params {
				fmt.Fprint(result, format.Object(param, 1), ",\n")
			}
			fmt.Fprintln(result, ")")
		}
	}
	return result.String()
}

// InterceptMockFailures runs a given callback and returns an array of
// failure messages generated by any Pegomock verifications within the callback.
//
// This is accomplished by temporarily replacing the *global* fail handler
// with a fail handler that simply annotates failures.  The original fail handler
// is reset when InterceptMockFailures returns.
func InterceptMockFailures(f func()) []string {
	originalHandler := GlobalFailHandler
	failures := []string{}
	RegisterMockFailHandler(func(message string, callerSkip ...int) {
		failures = append(failures, message)
	})
	f()
	RegisterMockFailHandler(originalHandler)
	return failures
}
