[![Build Status](https://travis-ci.org/petergtz/pegomock.svg?branch=master)](https://travis-ci.org/petergtz/pegomock)

PegoMock is a mocking framework for the [Go programming language](http://golang.org/). It integrates well with Go's built-in `testing` package, but can be used in other contexts too. It is based on [golang/mock](https://github.com/golang/mock), but uses a DSL closely related to [Mockito](http://site.mockito.org/mockito/docs/current/org/mockito/Mockito.html).

Getting Pegomock
================

Just `go get` it:

```
go get github.com/petergtz/pegomock/...
```

This will download the package and install an executable `pegomock` in your `$GOPATH/bin`.

Getting Started
===============

Using Pegomock with Golang’s XUnit-style Tests
----------------------------------------------

Use it like this:

```go

import (
	"github.com/petergtz/pegomock"
	"testing"
)

func TestUsingMocks(t *testing.T) {
	pegomock.RegisterMockTestingT(t)

	// use Pegomock here
}
```

There are two caveats:

-	You must register the `t *testing.T` passed to your test with Pegomock before you make any verifications associated with that test. So every `Test...` function in your suite should have the `RegisterTestingT(t)` line.
-	Pegomock uses a global (singleton) fail handler. This has the benefit that you don’t need to pass the fail handler down to each test, but does mean that you cannot run your XUnit style tests in parallel with Pegomock.

Using Pegomock with Ginkgo
--------------------------

When a Pegomock verification fails, it calls a `FailHandler`. This is a function that you must provide using `pegomock.RegisterMockFailHandler()`.

If you’re using [Ginkgo](http://onsi.github.io/ginkgo/), all you need to do is:

```go
pegomock.RegisterMockFailHandler(ginkgo.Fail)
```

before you start your test suite.

### Avoiding Ginkgo Naming Collision with `When` Function

Ginkgo introduced a new keyword in its DSL: `When`. This causes name collisions when dot-importing both Ginkgo and Pegomock. To avoid this, you can use a different dot-import for Pegomock which uses `Whenever` instead of `When`. Example:

```go
package some_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/petergtz/pegomock/ginkgo_compatible"
)

var _ = Describe("Some function", func() {
	When("certain condition", func() {
		It("succeeds", func() {
			mock := NewMockPhoneBook()
			Whenever(mock.GetPhoneNumber(EqString("Tom"))).ThenReturn("123-456-789")
		})
	})
})
```

Generating Your First Mock and Using It
---------------------------------------

Let's assume you have:

```go
type Display interface {
	Show(text string)
}
```

The simplest way is to call `pegomock` from within your go package specifying the interface by its name:

```
cd path/to/package
pegomock generate Display
```

This will generate a `mock_display_test.go` file which you can now use in your tests:

```go
// creating mock
display := NewMockDisplay()

// using the mock
display.Show("Hello World!")

// verifying
display.VerifyWasCalledOnce().Show("Hello World!")
```

Why yet Another Mocking Framework for Go?
=========================================

I've looked at some of the other frameworks, but found none of them satisfying:
- [GoMock](https://github.com/golang/mock) seemed overly complicated when setting up mocks and verifying them. The command line interface is also not quite intuitive. That said, Pegomock is based on the GoMock, reusing mostly the mockgen code.
- [Counterfeiter](https://github.com/maxbrunsfeld/counterfeiter) uses a DSL that I didn't find expressive enough. It often seems to need more lines of code too. In one of its samples, it uses e.g.:

	```go
	fake.DoThings("stuff", 5)
	Expect(fake.DoThingsCallCount()).To(Equal(1))

	str, num := fake.DoThingsArgsForCall(0)
	Expect(str).To(Equal("stuff"))
	Expect(num).To(Equal(uint64(5)))
	```

	In Pegomock, this can be written as simple as:

	```go
	fake.DoThings("stuff", 5)
	fake.VerifyWasCalledOnce().DoThings("stuff", 5)
	```
- [Hel](https://github.com/nelsam/hel) uses a new and interesting approach to setting up and verifying mocks. However, I wonder how flexible it actually is. E.g. how about providing a callback function when stubbing? Can this be modeled with its current approach using channels?

In addition, Pegomock provides a "watch" command similar to [Ginkgo](http://onsi.github.io/ginkgo/), which constantly watches over changes in an interface and updates its mocks. It gives the framework a much more dynamic feel, similar to mocking frameworks in Ruby or Java.

Using Mocks In Your Tests
=========================

Verifying Behavior
------------------

Interface:

```go
type Display interface {
	Show(text string)
}
```

Test:

```go
// creating mock:
display := NewMockDisplay()

// using the mock:
display.Show("Hello World!")

// verifying:
display.VerifyWasCalledOnce().Show("Hello World!")
```

Stubbing
--------

Interface:

```go
type PhoneBook interface {
	GetPhoneNumber(name string) string
}
```

Test:

```go
// creating the mock
phoneBook := NewMockPhoneBook()

// stubbing:
When(phoneBook.GetPhoneNumber("Tom")).ThenReturn("345-123-789")
When(phoneBook.GetPhoneNumber("Invalid")).ThenPanic("Invalid Name")

// prints "345-123-789":
fmt.Println(phoneBook.GetPhoneNumber("Tom"))

// panics:
fmt.Println(phoneBook.GetPhoneNumber("Invalid"))

// prints "", because GetPhoneNumber("Dan") was not stubbed
fmt.Println(phoneBook.GetPhoneNumber("Dan"))

// Although it is possible to verify a stubbed invocation, usually it's redundant
// If your code cares what GetPhoneNumber("Tom") returns, then something else breaks (often even before a verification gets executed).
// If your code doesn't care what GetPhoneNumber("Tom") returns, then it should not be stubbed.

// Not convinced? See http://monkeyisland.pl/2008/04/26/asking-and-telling.
phoneBook.VerifyWasCalledOnce().GetPhoneNumber("Tom")
```

-	By default, for all methods that return a value, a mock will return zero values.
-	Once stubbed, the method will always return a stubbed value, regardless of how many times it is called.

Stubbing Functions That Have no Return Value
--------------------------------------------

Stubbing functions that have no return value requires a slightly different approach, because such functions cannot be passed directly to another function. However, we can wrap them in an anonymous function:

```go
// creating mock:
display := NewMockDisplay()

// stubbing
When(func() { display.Show("Hello World!") }).ThenPanic("Panicking")

// panics:
display.Show("Hello World!")
```

Argument Matchers
-----------------

Pegomock provides matchers for stubbing and verification.

Verification:

```go
display := NewMockDisplay()

// Calling mock
display.Show("Hello again!")

// Verification:
display.VerifyWasCalledOnce().Show(AnyString())
```

Stubbing:

```go
phoneBook := NewMockPhoneBook()

// Stubbing:
When(phoneBook.GetPhoneNumber(AnyString())).ThenReturn("123-456-789")

// Prints "123-456-789":
fmt.Println(phoneBook.GetPhoneNumber("Dan"))
// Also prints "123-456-789":
fmt.Println(phoneBook.GetPhoneNumber("Tom"))
```

**Important**: When you use argument matchers, you must always use them for all arguments:

```go
// Incorrect, panics:
When(contactList.getContactByFullName("Dan", AnyString())).thenReturn(Contact{...})
// Correct:
When(contactList.getContactByFullName(EqString("Dan"), AnyString())).thenReturn(Contact{...})
```

### Writing Your Own Argument Matchers

**Important:** `Eq...` and `Any...` matchers for types used in mock methods, can now be _auto-generated_ while generating the mock. So writing your own argument matchers is not necessary for most use cases. See section [The Pegomock CLI](#generating-mocks) for more information.

You can also write your own matchers for non-basic types. E.g. if you have a `struct MyType`, you can write an _Equals_ and _Any_ matcher like this:
```go
func EqMyType(value MyType) MyType {
	RegisterMatcher(&EqMatcher{Value: value})
	return MyType{}
}

func AnyMyType() MyType {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf(MyType{})))
	return MyType{}
}
```


Verifying the Number of Invocations
-----------------------------------

```go
display := NewMockDisplay()

// Calling mock
display.Show("Hello")
display.Show("Hello, again")
display.Show("And again")

// Verification:
display.VerifyWasCalled(Times(3)).Show(AnyString())
// or:
display.VerifyWasCalled(AtLeast(3)).Show(AnyString())
// or:
display.VerifyWasCalled(Never()).Show("This one was never called")
```

Verifying in Order
------------------

```go
display1 := NewMockDisplay()
display2 := NewMockDisplay()

// Calling mocks
display1.Show("One")
display1.Show("Two")
display2.Show("Another two")
display1.Show("Three")

// Verification:
inOrderContext := new(InOrderContext)
display1.VerifyWasCalledInOrder(Once(), inOrderContext).Show("One")
display2.VerifyWasCalledInOrder(Once(), inOrderContext).Show("Another two")
display1.VerifyWasCalledInOrder(Once(), inOrderContext).Show("Three")
```

Note that it's not necessary to verify the call for `display.Show("Two")` if that one is not of any interested. An `InOrderContext` only verifies that the verifications that are done, are in order.

Stubbing with Callbacks
------------------------

```go
phoneBook := NewMockPhoneBook()

// Stubbing:
When(phoneBook.GetPhoneNumber(AnyString())).Then(func(params []Param) ReturnValues {
	return []ReturnValue{fmt.Sprintf("1-800-CALL-%v", strings.ToUpper(params[0]))}
},


// Prints "1-800-CALL-DAN":
fmt.Println(phoneBook.GetPhoneNumber("Dan"))
// Prints "1-800-CALL-TOM":
fmt.Println(phoneBook.GetPhoneNumber("Tom"))
```


Verifying with Argument Capture
--------------------------------

In some cases it can be useful to capture the arguments from mock invocations and assert on them separately. This method is only recommended if the techniques using matchers are not sufficient.

```go
display := NewMockDisplay()

// Calling mock
display.Show("Hello")
display.Show("Hello, again")
display.Show("And again")

// Verification and getting captured arguments
text := display.VerifyWasCalled(AtLeast(1)).Show(AnyString()).GetCapturedArguments()

// Captured arguments are from last invocation
Expect(text).To(Equal("And again"))
```

You can also get all captured arguments:

```go
// Verification and getting all captured arguments
texts := display.VerifyWasCalled(AtLeast(1)).Show(AnyString()).GetAllCapturedArguments()

// Captured arguments are a slice
Expect(texts).To(ConsistOf("Hello", "Hello, again", "And again"))
```



The Pegomock CLI
================

Installation
------------

Install it via:

```
go install github.com/petergtz/pegomock/pegomock
```

Generating Mocks
----------------

Pegomock can generate mocks in two different ways:

1.	by parsing source code Go files

	```
	pegomock generate [<flags>] <gofile>
	```

2.	by building a Go package and using reflection

	```
	pegomock generate [<flags>] [<packagepath>] <interfacename>
	```

Flags can be any of the following:

-	`--output,-o`: Output file; defaults to mock_<interface>_test.go.

-	`--package`: Package of the generated code; defaults to the package from which pegomock was executed suffixed with _test

- `--generate-matchers,-m`: This will auto-generate argument matchers and place them in a `matchers` directory alongside the mock source code itself.

For more flags, run:

```
pegomock --help
```

Generating Mocks with `--use-experimental-model-gen`
----------------------------------------------------

There are a number of shortcomings in the current reflection-based implementation.
To overcome these, there is now an option to use a new, experimental implementation that is based on [golang.org/x/tools/go/loader](https://godoc.org/golang.org/x/tools/go/loader).
To use it when generating your mocks, invoke `pegomock` like this:

```
pegomock generate --use-experimental-model-gen [<flags>] [<packagepath>] <interfacename>
```

What are the benefits?
- The current default uses the [reflect](https://golang.org/pkg/reflect/) package to introspect the interface for which a mock should be generated. But reflection cannot determine method parameter names, only types. This forces the generator to generate them based on a pattern. In a code editor with code assistence, those pattern-based names (such as `_param0`, `_param1`) are non-descriptive and provide less help while writing code. The new implementation properly parses the source (including *all* dependent packages) and subsequently uses the same names as used in the interface definition.
- With the current default you cannot generate an interface that lives in the `main` package. It's due to the way this implementation works: it imports the interface's package into temporarily generated code that gets compiled on the fly. This compilation fails, because there are now two `main` functions.
- The new implementation is simpler and will probably become the default in the future, because it will be easier to maintain.

What are the drawbacks?
- There is only one drawback: maturity. The new implementation is not complete yet, and also might have some bugs that still need to be fixed.

Users of Pegomock are encouraged to use this new option and report any problems by [opening an issue](https://github.com/petergtz/pegomock/issues/new). Help to stabilize it is greatly appreciated.

Generating mocks with `go generate`
----------------------------------

`pegomock` can be used with `go generate`. Simply add the directive to your source file.

Here's an example for a Display interface used by a calculator program:

```go
// package/path/to/display/display.go

package display

type Display interface {
	Show(text string)
}
```

```go
// package/path/to/calculator/calculator_test.go

package calculator_test

//go:generate pegomock generate package/path/to/display Display

// Use generated mock
mockDisplay := NewMockDisplay()
...
```

Generating it:
```sh
cd package/path/to/calculator
go generate
```

**Note:** While you could add the directive adjacent to the interface definition, the author's opinion is that this violates clean dependency management and would pollute the package of the interface.
It's better to generate the mock in the same package, where it is used (if this coincides with the interface package, that's fine). That way, not only stays the interface's package clean, the tests also don't need to prefix the mock with a package, or use a dot-import.

Continuously Generating Mocks
-----------------------------

The `watch` command lets Pegomock generate mocks continuously on every change to an interface:

```
pegomock watch
```

For this, Pegomock expects an `interfaces_to_mock` file in the package directory where the mocks should be generated. In fact, `pegomock watch` will create it for you if it doesn't exist yet. The contents of the file are similar to the ones of the `generate` command:

```
# Any line starting with a # is treated as comment.

# interface name without package specifies an Interface in the current package:
PhoneBook

 # generates a mock for SomeInterfacetaken from mypackage:
path/to/my/mypackage SomeInterface

# you can also specify a Go file:
display.go

# and use most of the flags from the "generate" command
--output my_special_output.go MyInterface
```

Flags can be:

- `--recursive,-r`: Recursively watch sub-directories as well.

Removing Generated Mocks
-----------------------------

Sometimes it can be useful to systematically remove all mocks and matcher files generated by Pegomock. For this purpose, there is the `remove` command. By simply calling it from the current directory
```
pegomock remove
```
it will remove all Pegomock-generated files in the current directory. It supports additional flags, such as `--recursive` to recursively remove all Pegomock-generated files in sub-directories as well. To see all possible options, run:
```
pegomock remove --help
```
