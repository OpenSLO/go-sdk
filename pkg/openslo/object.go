package openslo

import (
	"fmt"

	"github.com/nobl9/govy/pkg/govy"
)

// Object represents a generic OpenSLO object definition.
// All OpenSLO objects implement this interface.
type Object interface {
	fmt.Stringer
	// GetVersion returns the API [Version] of the [Object].
	GetVersion() Version
	// GetKind returns the [Kind] of the [Object].
	GetKind() Kind
	// GetName returns the name of the [Object].
	GetName() string
	// Validate performs static validation of the [Object].
	Validate() error
}

// ObjectValidator is an interface implemented by every [Object].
// It is separated from the [Object] interface in order to keep the latter non-generic.
type ObjectValidator[T Object] interface {
	// GetValidator returns a fully initialized [govy.Validator] instance for this specific [Object].
	GetValidator() govy.Validator[T]
}
