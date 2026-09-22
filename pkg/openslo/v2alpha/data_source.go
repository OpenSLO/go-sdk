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

// NewDataSource returns a [DataSource] from metadata and spec.
func NewDataSource(metadata Metadata, spec DataSourceSpec) DataSource {
	return DataSource{
		APIVersion: APIVersion,
		Kind:       openslo.KindDataSource,
		Metadata:   metadata,
		Spec:       spec,
	}
}

// DataSource represents reusable connection details for a metric source.
// [SLIMetricSpec.DataSourceRef] selects it by [Metadata.Name].
// A metric query can instead embed [SLIMetricSpec.DataSourceSpec].
// [SLIMetricSpec.Spec] contains implementation-defined query configuration.
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

// GetName returns the [DataSource]'s [Metadata.Name].
func (d DataSource) GetName() string {
	return d.Metadata.Name
}

// Validate returns an error for an invalid [DataSource].
func (d DataSource) Validate() error {
	return dataSourceValidation.Validate(d)
}

// String returns the [DataSource]'s formatted version and kind.
// It also returns [Metadata.Name] when set.
func (d DataSource) String() string {
	return internal.GetObjectName(d)
}

// GetMetadata returns the [DataSource]'s [Metadata].
func (d DataSource) GetMetadata() Metadata {
	return d.Metadata
}

// GetValidator returns the validator configured for [DataSource].
func (d DataSource) GetValidator() govy.Validator[DataSource] {
	return dataSourceValidation
}

// DataSourceSpec defines a metric-source type and its implementation-defined connection data.
type DataSourceSpec struct {
	// Description summarizes the [DataSource].
	Description string `json:"description,omitempty"`
	// Type identifies the metric-source type, such as Prometheus or Datadog.
	// The consuming implementation defines the accepted values.
	Type string `json:"type"`
	// ConnectionDetails contains implementation-defined connection data encoded as JSON,
	// such as endpoints or authentication settings.
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
		OmitEmpty().
		Rules(rules.StringMaxLength(1050)),
	govy.For(func(spec DataSourceSpec) string { return spec.Type }).
		WithName("type").
		Required(),
	govy.For(func(spec DataSourceSpec) json.RawMessage { return spec.ConnectionDetails }).
		WithName("connectionDetails").
		Required(),
)
