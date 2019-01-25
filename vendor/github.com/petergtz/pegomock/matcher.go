package pegomock

import (
	"fmt"
	"reflect"

	"github.com/petergtz/pegomock/internal/verify"
	"sync"
)

type EqMatcher struct {
	Value  Param
	actual Param
	sync.Mutex
}

func (matcher *EqMatcher) Matches(param Param) bool {
	matcher.Lock()
	defer matcher.Unlock()

	matcher.actual = param
	return reflect.DeepEqual(matcher.Value, param)
}

func (matcher *EqMatcher) FailureMessage() string {
	return fmt.Sprintf("Expected: %v; but got: %v", matcher.Value, matcher.actual)
}

func (matcher *EqMatcher) String() string {
	return fmt.Sprintf("Eq(%v)", matcher.Value)
}

type AnyMatcher struct {
	Type   reflect.Type
	actual reflect.Type
	sync.Mutex
}

func NewAnyMatcher(typ reflect.Type) *AnyMatcher {
	verify.Argument(typ != nil, "Must provide a non-nil type")
	return &AnyMatcher{Type: typ}
}

func (matcher *AnyMatcher) Matches(param Param) bool {
	matcher.Lock()
	defer matcher.Unlock()

	matcher.actual = reflect.TypeOf(param)
	if matcher.actual == nil {
		switch matcher.Type.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
			reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
			return true
		default:
			return false
		}
	}
	return matcher.actual.AssignableTo(matcher.Type)
}

func (matcher *AnyMatcher) FailureMessage() string {
	return fmt.Sprintf("Expected: %v; but got: %v", matcher.Type, matcher.actual)
}

func (matcher *AnyMatcher) String() string {
	return fmt.Sprintf("Any(%v)", matcher.Type)
}

type AtLeastIntMatcher struct {
	Value  int
	actual int
}

func (matcher *AtLeastIntMatcher) Matches(param Param) bool {
	matcher.actual = param.(int)
	return param.(int) >= matcher.Value
}

func (matcher *AtLeastIntMatcher) FailureMessage() string {
	return fmt.Sprintf("Expected: at least %v; but got: %v", matcher.Value, matcher.actual)
}

func (matcher *AtLeastIntMatcher) String() string {
	return fmt.Sprintf("AtLeast(%v)", matcher.Value)
}

type AtMostIntMatcher struct {
	Value  int
	actual int
}

func (matcher *AtMostIntMatcher) Matches(param Param) bool {
	matcher.actual = param.(int)
	return param.(int) <= matcher.Value
}

func (matcher *AtMostIntMatcher) FailureMessage() string {
	return fmt.Sprintf("Expected: at most %v; but got: %v", matcher.Value, matcher.actual)
}

func (matcher *AtMostIntMatcher) String() string {
	return fmt.Sprintf("AtMost(%v)", matcher.Value)
}
