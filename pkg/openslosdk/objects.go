package openslosdk

import "github.com/OpenSLO/go-sdk/pkg/openslo"

// FilterByKind filters [openslo.Object] slice and returns its subset matching the type constraint.
func FilterByKind[T openslo.Object](objects []openslo.Object) []T {
	var filtered []T
	for i := range objects {
		if v, ok := objects[i].(T); ok {
			filtered = append(filtered, v)
		}
	}
	return filtered
}
