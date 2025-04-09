package openslo

import "fmt"

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
