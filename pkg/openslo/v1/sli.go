package v1

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

// SLI defines a derived reliability indicator and the queries used to calculate
// it for an [SLO].
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

// GetName returns the name in the SLI's [Metadata].
func (s SLI) GetName() string {
	return s.Metadata.Name
}

// Validate returns an error for an invalid SLI.
func (s SLI) Validate() error {
	return sliValidation.Validate(s)
}

// String returns the SLI's formatted version and kind. It also returns
// [Metadata.Name] when set.
func (s SLI) String() string {
	return internal.GetObjectName(s)
}

// GetMetadata returns the SLI's [Metadata].
func (s SLI) GetMetadata() Metadata {
	return s.Metadata
}

// GetValidator returns the validator for SLI objects.
func (s SLI) GetValidator() govy.Validator[SLI] {
	return sliValidation
}

// SLISpec defines the query or queries used to calculate an [SLI].
type SLISpec struct {
	// Description summarizes the SLI.
	Description string `json:"description,omitempty"`
	// ThresholdMetric defines a query that returns raw values.
	// [SLOObjective.Operator] compares each value with [SLOObjective.Value].
	ThresholdMetric *SLIMetricSpec  `json:"thresholdMetric,omitempty"`
	RatioMetric     *SLIRatioMetric `json:"ratioMetric,omitempty"`
}

// SLIRatioMetric defines an indicator from good divided by total or (total minus
// bad) divided by total. It can instead use a precomputed success or failure
// ratio identified by [SLIRatioMetric.RawType]. For example, 99 good events out
// of 100 produce a ratio of 0.99. One bad event out of 100 produces the same
// ratio.
type SLIRatioMetric struct {
	// Counter reports whether the queried good, bad, and total metrics are
	// monotonically increasing. It has no effect when Raw is used.
	Counter bool `json:"counter"`
	// Good supplies the numerator for a good-over-total ratio.
	Good *SLIMetricSpec `json:"good,omitempty"`
	// Bad supplies the number subtracted from Total for a failure-based ratio.
	Bad *SLIMetricSpec `json:"bad,omitempty"`
	// Total supplies the denominator for a Good- or Bad-based ratio.
	Total *SLIMetricSpec `json:"total,omitempty"`
	// RawType selects whether Raw is interpreted as a success or failure ratio
	// when Raw is used.
	RawType SLIRawMetricType `json:"rawType,omitempty"`
	// Raw defines a query for a precomputed success or failure ratio.
	Raw *SLIMetricSpec `json:"raw,omitempty"`
}

// SLIMetricSpec defines one query used to read metric data for an [SLI].
type SLIMetricSpec struct {
	MetricSource SLIMetricSource `json:"metricSource"`
}

// SLIMetricSource identifies a metrics backend and supplies the configuration
// needed to retrieve a metric.
type SLIMetricSource struct {
	// MetricSourceRef names an existing [DataSource].
	MetricSourceRef string `json:"metricSourceRef,omitempty"`
	// Type identifies the implementation-defined metric-source type.
	// OpenSLO requires Type when [SLIMetricSource.MetricSourceRef] is omitted.
	// Otherwise, OpenSLO infers Type from the referenced [DataSource]. This SDK
	// does not validate the requirement.
	Type string `json:"type,omitempty"`
	// Spec contains source-specific query or metric-retrieval configuration.
	Spec map[string]any `json:"spec"`
}

// SLIRawMetricType identifies how a precomputed raw ratio is interpreted.
type SLIRawMetricType string

const (
	// SLIRawMetricTypeSuccess interprets Raw as good divided by total.
	SLIRawMetricTypeSuccess SLIRawMetricType = "success"
	// SLIRawMetricTypeFailure interprets Raw as bad divided by total.
	SLIRawMetricTypeFailure SLIRawMetricType = "failure"
)

var validSLIRawMetricTypes = []SLIRawMetricType{
	SLIRawMetricTypeSuccess,
	SLIRawMetricTypeFailure,
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
		OmitEmpty().
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
	govy.For(func(spec SLIMetricSpec) SLIMetricSource { return spec.MetricSource }).
		WithName("metricSource").
		Required().
		Include(govy.New(
			govy.For(func(source SLIMetricSource) string { return source.MetricSourceRef }).
				WithName("metricSourceRef").
				OmitEmpty().
				Rules(rules.StringDNSLabel()),
			govy.For(func(source SLIMetricSource) map[string]any { return source.Spec }).
				WithName("spec").
				Required().
				Rules(rules.MapMinLength[map[string]any](1)),
		)),
)
