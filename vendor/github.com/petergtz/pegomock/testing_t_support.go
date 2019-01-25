// Copyright (c) 2013-2014 Onsi Fakhouri
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package pegomock

import (
	"regexp"
	"runtime/debug"
	"strings"
)

type testingT interface {
	Errorf(format string, args ...interface{})
}

func BuildTestingTGomegaFailHandler(t testingT) FailHandler {
	return func(message string, callerSkip ...int) {
		skip := 1
		if len(callerSkip) > 0 {
			skip = callerSkip[0]
		}
		stackTrace := pruneStack(string(debug.Stack()), skip)
		t.Errorf("\n%s\n%s", stackTrace, message)
	}
}

func pruneStack(fullStackTrace string, skip int) string {
	stack := strings.Split(fullStackTrace, "\n")
	if len(stack) > 2*(skip+1) {
		stack = stack[2*(skip+1):]
	}
	prunedStack := []string{}
	re := regexp.MustCompile(`\/ginkgo\/|\/pkg\/testing\/|\/pkg\/runtime\/`)
	for i := 0; i < len(stack)/2; i++ {
		if !re.Match([]byte(stack[i*2])) {
			prunedStack = append(prunedStack, stack[i*2])
			prunedStack = append(prunedStack, stack[i*2+1])
		}
	}
	return strings.Join(prunedStack, "\n")
}
