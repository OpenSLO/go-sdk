package v1

import (
	"encoding/json"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/OpenSLO/go-sdk/internal"
	"github.com/OpenSLO/go-sdk/pkg/openslo"
)

var (
	_ = Object(DataSource{})
	_ = openslo.ObjectValidator[DataSource](DataSource{})
)

// NewDataSource returns a data source from metadata and spec.
func NewDataSource(metadata Metadata, spec DataSourceSpec) DataSource {
	return DataSource{
		APIVersion: APIVersion,
		Kind:       openslo.KindDataSource,
		Metadata:   metadata,
		Spec:       spec,
	}
}

// DataSource stores reusable connection details for a metrics backend.
type DataSource struct {
	APIVersion openslo.Version `json:"apiVersion"`
	Kind       openslo.Kind    `json:"kind"`
	Metadata   Metadata        `json:"metadata"`
	Spec       DataSourceSpec  `json:"spec"`
}

// GetVersion returns [APIVersion].
func (d DataSource) GetVersion() openslo.Version {
	return APIVersion
}

// GetKind returns [openslo.KindDataSource].
func (d DataSource) GetKind() openslo.Kind {
	return openslo.KindDataSource
}

// GetName returns the name in the data source's [Metadata].
func (d DataSource) GetName() string {
	return d.Metadata.Name
}

// Validate returns an error for an invalid data source.
func (d DataSource) Validate() error {
	return dataSourceValidation.Validate(d)
}

// String returns the data source's formatted version, kind, and name.
func (d DataSource) String() string {
	return internal.GetObjectName(d)
}

// GetMetadata returns the data source's [Metadata].
func (d DataSource) GetMetadata() Metadata {
	return d.Metadata
}

// GetValidator returns the validator for DataSource objects.
func (d DataSource) GetValidator() govy.Validator[DataSource] {
	return dataSourceValidation
}

// DataSourceSpec defines reusable, source-specific connection configuration.
type DataSourceSpec struct {
	// Description summarizes the data source.
	Description string `json:"description,omitempty"`
	// Type names the implementation-defined metric source.
	Type string `json:"type"`
	// ConnectionDetails stores source-specific connection configuration as JSON.
	ConnectionDetails json.RawMessage `json:"connectionDetails"`
}

var dataSourceValidation = govy.New(
	validationRulesAPIVersion(func(d DataSource) openslo.Version { return d.APIVersion }),
	validationRulesKind(func(d DataSource) openslo.Kind { return d.Kind }, openslo.KindDataSource),
	validationRulesMetadata(func(d DataSource) Metadata { return d.Metadata }),
	govy.For(func(d DataSource) DataSourceSpec { return d.Spec }).
		WithName("spec").
		Include(govy.New(
			govy.For(func(spec DataSourceSpec) string { return spec.Description }).
				WithName("description").
				Rules(rules.StringMaxLength(1050)),
			govy.For(func(spec DataSourceSpec) string { return spec.Type }).
				WithName("type").
				Required(),
			govy.For(func(spec DataSourceSpec) json.RawMessage { return spec.ConnectionDetails }).
				WithName("connectionDetails").
				Required(),
		)),
).WithNameFunc(internal.GetObjectName[DataSource])
