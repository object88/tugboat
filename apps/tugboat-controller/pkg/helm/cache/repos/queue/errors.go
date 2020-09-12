package queue

import "github.com/object88/tugboat/pkg/errs"

const (
	// ErrEmptyKey results from attempting to use an empty string for a key.  All
	// keys must be a non-empty string value.
	ErrEmptyKey = errs.ConstError("Key may not be an empty string")

	// ErrNilRespondent results from attempting to pass a nil channel to
	// `Line.Enqueue`.
	ErrNilRespondent = errs.ConstError("Respondent channel may not be nil")
)
