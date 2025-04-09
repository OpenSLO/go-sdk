package internal

import (
	"fmt"
	"strings"

	"github.com/OpenSLO/go-sdk/pkg/openslo"
)

// GetObjectName returns a pretty-formatted name of the [openslo.Object].
func GetObjectName[T openslo.Object](o T) string {
	version := o.GetVersion().String()
	i := strings.Index(version, "/")
	if i == -1 {
		return ""
	}
	version = version[i+1:]
	if name := o.GetName(); name != "" {
		return fmt.Sprintf("%s.%s '%s'", version, o.GetKind(), o.GetName())
	}
	return fmt.Sprintf("%s.%s", version, o.GetKind())
}
