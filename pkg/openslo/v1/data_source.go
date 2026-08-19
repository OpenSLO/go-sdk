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

// DataSource represents reusable connection details for a metric source.
// [SLIMetricSource.MetricSourceRef] selects it by [Metadata.Name]. A referenced
// metric source can omit [SLIMetricSource.Type] and keep connection details,
// such as authentication settings, outside the [SLI]. An SLI can instead define
// an inline metric source with [SLIMetricSource.Type] and [SLIMetricSource.Spec].
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

// String returns the data source's formatted version and kind. It also returns
// [Metadata.Name] when set.
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
	// Type identifies the implementation-defined metric source type, such as
	// Prometheus or Datadog.
	Type string `json:"type"`
	// ConnectionDetails contains implementation-defined connection data encoded
	// as JSON. The metric-source implementation defines its fields, which can
	// include endpoints or authentication settings.
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
				OmitEmpty().
				Rules(rules.StringMaxLength(1050)),
			govy.For(func(spec DataSourceSpec) string { return spec.Type }).
				WithName("type").
				Required(),
			govy.For(func(spec DataSourceSpec) json.RawMessage { return spec.ConnectionDetails }).
				WithName("connectionDetails").
				Required(),
		)),
).WithNameFunc(internal.GetObjectName[DataSource])
