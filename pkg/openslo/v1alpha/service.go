package v1alpha

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

// Service groups related SLOs.
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

// String returns the qualified object name.
func (s Service) String() string {
	return internal.GetObjectName(s)
}

// GetMetadata returns the service's metadata.
func (s Service) GetMetadata() Metadata {
	return s.Metadata
}

// GetValidator returns the validator used by [Service.Validate].
func (s Service) GetValidator() govy.Validator[Service] {
	return serviceValidation
}

// ServiceSpec contains the descriptive properties of a [Service].
type ServiceSpec struct {
	// Description summarizes the service.
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
				Rules(rules.StringMaxLength(1050)),
		)),
).WithNameFunc(internal.GetObjectName[Service])
