package openslosdk

import (
	"github.com/nobl9/govy/pkg/govy"

	"github.com/OpenSLO/go-sdk/internal"
	"github.com/OpenSLO/go-sdk/pkg/openslo"
)

func Validate(objects ...openslo.Object) error {
	return objectsValidator.ValidateSlice(objects)
}

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
	WithNameFunc(internal.GetObjectName)
