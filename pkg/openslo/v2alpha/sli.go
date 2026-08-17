package v2alpha

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/OpenSLO/go-sdk/internal"
	"github.com/OpenSLO/go-sdk/pkg/openslo"
)

var (
	_ = Object(SLI{})
	_ = openslo.ObjectValidator[SLI](SLI{})
)

// NewSLI returns an SLI from metadata and spec.
func NewSLI(metadata Metadata, spec SLISpec) SLI {
	return SLI{
		APIVersion: APIVersion,
		Kind:       openslo.KindSLI,
		Metadata:   metadata,
		Spec:       spec,
	}
}

// SLI describes how to read a metric from a data source, as defined by the
// [OpenSLO v2alpha SLI].
//
// [OpenSLO v2alpha SLI]: https://github.com/OpenSLO/OpenSLO/blob/e74b589cc98b98a5413611176d659a72318e7519/enhancements/v2alpha.md#sli
type SLI struct {
	APIVersion openslo.Version `json:"apiVersion"`
	Kind       openslo.Kind    `json:"kind"`
	Metadata   Metadata        `json:"metadata"`
	Spec       SLISpec         `json:"spec"`
}

// GetVersion returns [APIVersion].
func (s SLI) GetVersion() openslo.Version {
	return APIVersion
}

// GetKind returns [openslo.KindSLI].
func (s SLI) GetKind() openslo.Kind {
	return openslo.KindSLI
}

// GetName returns the SLI's metadata name.
func (s SLI) GetName() string {
	return s.Metadata.Name
}

// Validate returns an error for an invalid SLI.
func (s SLI) Validate() error {
	return sliValidation.Validate(s)
}

// String returns the SLI's formatted version, kind, and name.
func (s SLI) String() string {
	return internal.GetObjectName(s)
}

// GetMetadata returns the SLI's metadata.
func (s SLI) GetMetadata() Metadata {
	return s.Metadata
}

// GetValidator returns the validator configured for [SLI].
func (s SLI) GetValidator() govy.Validator[SLI] {
	return sliValidation
}

// SLISpec defines the metric used by an [SLI].
type SLISpec struct {
	// Description summarizes the indicator.
	Description string `json:"description,omitempty"`
	// ThresholdMetric supplies values for [SLOObjective.Operator] to compare with
	// [SLOObjective.Value].
	ThresholdMetric *SLIMetricSpec `json:"thresholdMetric,omitempty"`
	// RatioMetric supplies a success ratio for an SLO objective.
	RatioMetric *SLIRatioMetric `json:"ratioMetric,omitempty"`
}

// SLIRatioMetric derives a success ratio as good divided by total, total minus
// bad divided by total, or a precomputed raw ratio.
// For example, 990 good events out of 1,000 total events produce 0.99.
// 10 bad events with the same total produce the same success ratio.
type SLIRatioMetric struct {
	// Counter reports whether the good, bad, and total metrics are monotonically
	// increasing counters. It has no effect when Raw is used.
	Counter bool `json:"counter"`
	// Good is the success-count numerator used with Total.
	Good *SLIMetricSpec `json:"good,omitempty"`
	// Bad is the failure-count input used with Total to derive successes.
	Bad *SLIMetricSpec `json:"bad,omitempty"`
	// Total is the denominator paired with Good or Bad.
	Total *SLIMetricSpec `json:"total,omitempty"`
	// RawType identifies whether Raw contains a success or failure ratio when
	// Raw is used.
	RawType SLIRawMetricType `json:"rawType,omitempty"`
	// Raw supplies an already computed ratio.
	Raw *SLIMetricSpec `json:"raw,omitempty"`
}

// SLIRawMetricType identifies whether a raw ratio contains successes
// (good/total) or failures (bad/total).
type SLIRawMetricType string

const (
	SLIRawMetricTypeSuccess SLIRawMetricType = "success"
	SLIRawMetricTypeFailure SLIRawMetricType = "failure"
)

var validSLIRawMetricTypes = []SLIRawMetricType{
	SLIRawMetricTypeSuccess,
	SLIRawMetricTypeFailure,
}

// SLIMetricSpec supplies a provider-specific query in the v2alpha flattened
// layout.
type SLIMetricSpec struct {
	// DataSourceRef names an existing [DataSource].
	DataSourceRef string `json:"dataSourceRef,omitempty"`
	// DataSourceSpec embeds the data-source configuration.
	DataSourceSpec *DataSourceSpec `json:"dataSourceSpec,omitempty"`
	// Spec contains provider-specific query configuration at the same level as
	// the data-source selection.
	Spec map[string]any `json:"spec,omitempty"`
}

var sliValidation = govy.New(
	validationRulesAPIVersion(func(s SLI) openslo.Version { return s.APIVersion }),
	validationRulesKind(func(s SLI) openslo.Kind { return s.Kind }, openslo.KindSLI),
	validationRulesMetadata(func(s SLI) Metadata { return s.Metadata }),
	govy.For(func(s SLI) SLISpec { return s.Spec }).
		WithName("spec").
		Include(sliSpecValidation),
).WithNameFunc(internal.GetObjectName[SLI])

var sliSpecValidation = govy.New(
	govy.For(func(spec SLISpec) string { return spec.Description }).
		WithName("description").
		Rules(rules.StringMaxLength(1050)),
	govy.For(govy.GetSelf[SLISpec]()).
		Rules(rules.MutuallyExclusive(true, map[string]func(s SLISpec) any{
			"thresholdMetric": func(s SLISpec) any { return s.ThresholdMetric },
			"ratioMetric":     func(s SLISpec) any { return s.RatioMetric },
		}).
			WithDescription("exactly one of 'thresholdMetric' and 'ratioMetric' must be set")),
	govy.ForPointer(func(spec SLISpec) *SLIMetricSpec { return spec.ThresholdMetric }).
		WithName("thresholdMetric").
		Include(sliMetricSpecValidation),
	govy.ForPointer(func(spec SLISpec) *SLIRatioMetric { return spec.RatioMetric }).
		WithName("ratioMetric").
		Include(sliRatioMetricValidation),
)

var sliRatioMetricValidation = govy.New(
	govy.For(govy.GetSelf[SLIRatioMetric]()).
		Cascade(govy.CascadeModeStop).
		Rules(rules.MutuallyExclusive(true, map[string]func(m SLIRatioMetric) any{
			"total": func(m SLIRatioMetric) any { return m.Total },
			"raw":   func(m SLIRatioMetric) any { return m.Raw },
		}).
			WithDescription("exactly one of 'total' and 'raw' must be set")).
		Rules(rules.MutuallyExclusive(false, map[string]func(m SLIRatioMetric) any{
			"raw":  func(m SLIRatioMetric) any { return m.Raw },
			"good": func(m SLIRatioMetric) any { return m.Good },
			"bad":  func(m SLIRatioMetric) any { return m.Bad },
		})).
		Include(sliFractionMetricValidation).
		Include(sliRawMetricSpecValidation),
)

var sliFractionMetricValidation = govy.New(
	govy.For(govy.GetSelf[SLIRatioMetric]()).
		Rules(rules.OneOfProperties(map[string]func(m SLIRatioMetric) any{
			"good": func(m SLIRatioMetric) any { return m.Good },
			"bad":  func(m SLIRatioMetric) any { return m.Bad },
		})),
	govy.ForPointer(func(m SLIRatioMetric) *SLIMetricSpec { return m.Total }).
		WithName("total").
		Cascade(govy.CascadeModeContinue).
		Include(sliMetricSpecValidation),
	govy.ForPointer(func(m SLIRatioMetric) *SLIMetricSpec { return m.Good }).
		WithName("good").
		Cascade(govy.CascadeModeContinue).
		When(
			func(m SLIRatioMetric) bool { return m.Good != nil },
			govy.WhenDescription("'good' is set"),
		).
		Include(sliMetricSpecValidation),
	govy.ForPointer(func(m SLIRatioMetric) *SLIMetricSpec { return m.Bad }).
		WithName("bad").
		Cascade(govy.CascadeModeContinue).
		When(
			func(m SLIRatioMetric) bool { return m.Bad != nil },
			govy.WhenDescription("'bad' is set"),
		).
		Include(sliMetricSpecValidation),
).
	Cascade(govy.CascadeModeStop).
	When(
		func(m SLIRatioMetric) bool { return m.Total != nil },
		govy.WhenDescription("'total' is set"),
	)

var sliRawMetricSpecValidation = govy.New(
	govy.ForPointer(func(m SLIRatioMetric) *SLIMetricSpec { return m.Raw }).
		WithName("raw").
		Include(sliMetricSpecValidation),
	govy.For(func(m SLIRatioMetric) SLIRawMetricType { return m.RawType }).
		WithName("rawType").
		Required().
		Rules(rules.OneOf(validSLIRawMetricTypes...)),
).
	When(
		func(m SLIRatioMetric) bool { return m.Raw != nil },
		govy.WhenDescription("'raw' is set"),
	)

var sliMetricSpecValidation = govy.New(
	govy.For(govy.GetSelf[SLIMetricSpec]()).
		Rules(rules.MutuallyExclusive(true, map[string]func(s SLIMetricSpec) any{
			"dataSourceRef":  func(s SLIMetricSpec) any { return s.DataSourceRef },
			"dataSourceSpec": func(s SLIMetricSpec) any { return s.DataSourceSpec },
		}).
			WithDescription("exactly one of 'dataSourceRef' and 'dataSourceSpec' must be set")),
	govy.For(func(spec SLIMetricSpec) string { return spec.DataSourceRef }).
		WithName("dataSourceRef").
		OmitEmpty().
		Rules(rules.StringDNSLabel()),
	govy.ForPointer(func(spec SLIMetricSpec) *DataSourceSpec { return spec.DataSourceSpec }).
		WithName("dataSourceSpec").
		Include(dataSourceSpecValidation),
)
