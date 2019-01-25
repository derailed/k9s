package pegomock

type FailHandler func(message string, callerSkip ...int)

type Mock interface{}
type Param interface{}
type ReturnValue interface{}
type ReturnValues []ReturnValue
