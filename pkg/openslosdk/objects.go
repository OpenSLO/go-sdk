package openslosdk

import "github.com/OpenSLO/go-sdk/pkg/openslo"

// FilterByType filters [openslo.Object] slice and returns its subset matching the type constraint.
// You can use it to filter:
//   - specific version and kind, like [v1.SLO]
//   - specific version, like [v1.Object]
func FilterByType[T openslo.Object](objects []openslo.Object) []T {
	var filtered []T
	for i := range objects {
		if v, ok := objects[i].(T); ok {
			filtered = append(filtered, v)
		}
	}
	return filtered
}
