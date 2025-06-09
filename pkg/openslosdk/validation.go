package openslosdk

import (
	"github.com/nobl9/govy/pkg/govy"

	"github.com/OpenSLO/go-sdk/internal"
	"github.com/OpenSLO/go-sdk/pkg/openslo"
)

func Validate(objects ...openslo.Object) error {
	errs := make(govy.ValidatorErrors, 0)
	for i, object := range objects {
		err := object.Validate()
		if err == nil {
			continue
		}
		vErr, ok := err.(*govy.ValidatorError)
		if !ok {
			return err
		}
		vErr.SliceIndex = &i
		vErr.Name = internal.GetObjectName(object)
		errs = append(errs, vErr)
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}
