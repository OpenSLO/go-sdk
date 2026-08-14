package v2alpha

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

// DataSource makes metric-source connection details reusable across SLIs.
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

// GetName returns the data source's metadata name.
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

// GetMetadata returns the data source's metadata.
func (d DataSource) GetMetadata() Metadata {
	return d.Metadata
}

// GetValidator returns the validator configured for [DataSource].
func (d DataSource) GetValidator() govy.Validator[DataSource] {
	return dataSourceValidation
}

// DataSourceSpec defines a data-source type and its provider-specific connection
// configuration.
type DataSourceSpec struct {
	// Description summarizes the data source.
	Description string `json:"description,omitempty"`
	// Type identifies the data-source implementation understood by the consumer.
	Type string `json:"type"`
	// ConnectionDetails stores provider-specific connection configuration as JSON.
	ConnectionDetails json.RawMessage `json:"connectionDetails"`
}

var dataSourceValidation = govy.New(
	validationRulesAPIVersion(func(d DataSource) openslo.Version { return d.APIVersion }),
	validationRulesKind(func(d DataSource) openslo.Kind { return d.Kind }, openslo.KindDataSource),
	validationRulesMetadata(func(d DataSource) Metadata { return d.Metadata }),
	govy.For(func(d DataSource) DataSourceSpec { return d.Spec }).
		WithName("spec").
		Required().
		Include(dataSourceSpecValidation),
).WithNameFunc(internal.GetObjectName[DataSource])

var dataSourceSpecValidation = govy.New(
	govy.For(func(spec DataSourceSpec) string { return spec.Description }).
		WithName("description").
		Rules(rules.StringMaxLength(1050)),
	govy.For(func(spec DataSourceSpec) string { return spec.Type }).
		WithName("type").
		Required(),
	govy.For(func(spec DataSourceSpec) json.RawMessage { return spec.ConnectionDetails }).
		WithName("connectionDetails").
		Required(),
)
