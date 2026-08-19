package v2alpha

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/OpenSLO/go-sdk/internal"
	"github.com/OpenSLO/go-sdk/pkg/openslo"
)

var (
	_ = Object(Service{})
	_ = openslo.ObjectValidator[Service](Service{})
)

// NewService returns a service from metadata and spec.
func NewService(metadata Metadata, spec ServiceSpec) Service {
	return Service{
		APIVersion: APIVersion,
		Kind:       openslo.KindService,
		Metadata:   metadata,
		Spec:       spec,
	}
}

// Service identifies a high-level group for SLOs.
// Each [SLO] refers to a Service by metadata name through [SLOSpec.ServiceRef].
// Multiple SLOs can use the same Service name.
// The SDK does not verify that a referenced Service exists.
//
// This type is the SDK's v2alpha representation. The living, unstable
// [OpenSLO v2alpha proposal] does not define a standalone Service schema.
//
// [OpenSLO v2alpha proposal]: https://github.com/OpenSLO/OpenSLO/blob/e74b589cc98b98a5413611176d659a72318e7519/enhancements/v2alpha.md
type Service struct {
	APIVersion openslo.Version `json:"apiVersion"`
	Kind       openslo.Kind    `json:"kind"`
	Metadata   Metadata        `json:"metadata"`
	Spec       ServiceSpec     `json:"spec"`
}

// GetVersion returns [APIVersion].
func (s Service) GetVersion() openslo.Version {
	return APIVersion
}

// GetKind returns [openslo.KindService].
func (s Service) GetKind() openslo.Kind {
	return openslo.KindService
}

// GetName returns the service's metadata name.
func (s Service) GetName() string {
	return s.Metadata.Name
}

// Validate returns an error for an invalid service.
func (s Service) Validate() error {
	return serviceValidation.Validate(s)
}

// String returns the service's formatted version and kind.
// It also returns the metadata name when set.
func (s Service) String() string {
	return internal.GetObjectName(s)
}

// GetMetadata returns the service's metadata.
func (s Service) GetMetadata() Metadata {
	return s.Metadata
}

// GetValidator returns the validator configured for [Service].
func (s Service) GetValidator() govy.Validator[Service] {
	return serviceValidation
}

// ServiceSpec defines the descriptive attributes of a [Service].
type ServiceSpec struct {
	// Description optionally summarizes the service in at most 1,050 characters.
	Description string `json:"description,omitempty"`
}

var serviceValidation = govy.New(
	validationRulesAPIVersion(func(s Service) openslo.Version { return s.APIVersion }),
	validationRulesKind(func(s Service) openslo.Kind { return s.Kind }, openslo.KindService),
	validationRulesMetadata(func(s Service) Metadata { return s.Metadata }),
	govy.For(func(s Service) ServiceSpec { return s.Spec }).
		WithName("spec").
		Include(govy.New(
			govy.For(func(spec ServiceSpec) string { return spec.Description }).
				WithName("description").
				OmitEmpty().
				Rules(rules.StringMaxLength(1050)),
		)),
).WithNameFunc(internal.GetObjectName[Service])
