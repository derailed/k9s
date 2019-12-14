package dao

// Alias represents an alias resource.
type Alias struct {
	Generic
}

var _ Accessor = &Alias{}
