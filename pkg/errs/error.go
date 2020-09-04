package errs

// ConstError is an error that can be declared as a constant:
// const ErrNope = errors.ConstError("Nope")
// This allows for easy comparison:
// err := Foo()
// if err == ErrNope { ... }
type ConstError string

// Error fulfils the builtin `error` interface
func (ce ConstError) Error() string {
	return string(ce)
}

const (
	// ErrNilPointerReceived indicates that a nil pointer was used to invoke a
	// func when a non-nil receiver was required:
	// var x *pkg.X
	// err := x.DoSomething()
	ErrNilPointerReceived = ConstError("Required non-nil pointer receiver for func is nil")
)
