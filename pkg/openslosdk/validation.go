package openslosdk

import (
	"fmt"

	"github.com/nobl9/govy/pkg/govy"

	"github.com/OpenSLO/go-sdk/pkg/openslo"
)

var objectsValidator = govy.New(
	govy.For(govy.GetSelf[openslo.Object]()).
		Rules(
			govy.NewRule(func(o openslo.Object) error {
				err := o.Validate()
				if vErr, ok := err.(*govy.ValidatorError); ok {
					return vErr.Errors
				}
				return err
			}),
		),
).
	WithNameFunc(func(o openslo.Object) string {
		return fmt.Sprintf("%s %s", o.GetVersion(), o.GetKind())
	})

func Validate(objects ...openslo.Object) error {
	return objectsValidator.ValidateSlice(objects)
}
