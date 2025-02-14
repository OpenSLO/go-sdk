package openslosdk

import "fmt"

// ObjectFormat represents the serialization format of [openslo.Object].
type ObjectFormat int

const (
	FormatYAML ObjectFormat = iota + 1
	FormatJSON
)

// String implements the [fmt.Stringer] interface.
func (f ObjectFormat) String() string {
	switch f {
	case FormatYAML:
		return "yaml"
	case FormatJSON:
		return "json"
	default:
		return "unknown"
	}
}

// Validate checks if [ObjectFormat] is supported.
func (f ObjectFormat) Validate() error {
	switch f {
	case FormatYAML, FormatJSON:
		return nil
	default:
		return fmt.Errorf("unsupported %[1]T: %[1]s", f)
	}
}
