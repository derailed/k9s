package verify

import "fmt"

func Argument(arg bool, message string, a ...interface{}) {
	if !arg {
		panic(fmt.Sprintf(message, a...))
	}
}
