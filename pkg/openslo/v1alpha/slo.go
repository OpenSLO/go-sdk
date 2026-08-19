package v1alpha

import (
	"errors"
	"time"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/OpenSLO/go-sdk/internal"
	"github.com/OpenSLO/go-sdk/pkg/openslo"
)

var (
	_ = Object(SLO{})
	_ = openslo.ObjectValidator[SLO](SLO{})
)

// NewSLO returns an SLO from metadata and spec.
func NewSLO(metadata Metadata, spec SLOSpec) SLO {
	return SLO{
		APIVersion: APIVersion,
		Kind:       openslo.KindSLO,
		Metadata:   metadata,
		Spec:       spec,
	}
}

// SLO is the legacy v1alpha SLO representation supported by this SDK. It
// defines reliability targets for a service level measured by an indicator.
type SLO struct {
	APIVersion openslo.Version `json:"apiVersion"`
	Kind       openslo.Kind    `json:"kind"`
	Metadata   Metadata        `json:"metadata"`
	Spec       SLOSpec         `json:"spec"`
}

// GetVersion returns [APIVersion].
func (s SLO) GetVersion() openslo.Version {
	return APIVersion
}

// GetKind returns [openslo.KindSLO].
func (s SLO) GetKind() openslo.Kind {
	return openslo.KindSLO
}

// GetName returns the SLO's metadata name.
func (s SLO) GetName() string {
	return s.Metadata.Name
}

// Validate returns an error for an invalid SLO.
func (s SLO) Validate() error {
	return sloValidation.Validate(s)
}

// String returns the SLO's formatted version and kind. It also returns
// [Metadata.Name] when set.
func (s SLO) String() string {
	return internal.GetObjectName(s)
}

// GetMetadata returns the SLO's metadata.
func (s SLO) GetMetadata() Metadata {
	return s.Metadata
}

// GetValidator returns the validator used by [SLO.Validate].
func (s SLO) GetValidator() govy.Validator[SLO] {
	return sloValidation
}

// SLOSpec defines the service, indicator, objectives, time window, and
// error-budget calculation for an [SLO].
type SLOSpec struct {
	// TimeWindows contains exactly one SLO evaluation window.
	TimeWindows []SLOTimeWindow `json:"timeWindows"`
	// BudgetingMethod applies the selected error-budget calculation to every
	// objective.
	BudgetingMethod SLOBudgetingMethod `json:"budgetingMethod"`
	// Description is an optional summary of the SLO.
	Description string `json:"description,omitempty"`
	// Indicator defines the threshold-metric form of the SLO. It must be nil for
	// the ratio form.
	Indicator *SLOIndicator `json:"indicator"`
	// Service is the metadata name of the [Service] whose reliability the SLO
	// measures.
	Service string `json:"service"`
	// Objectives contains reliability targets. For the ratio form, each
	// objective's [SLOObjective.RatioMetrics] defines the SLI metric queries.
	Objectives []SLOObjective `json:"objectives"`
}

// SLOBudgetingMethod identifies how an SLO calculates its error budget.
// Occurrences weights each event equally. Timeslices weights each time slice
// equally.
type SLOBudgetingMethod string

const (
	SLOBudgetingMethodOccurrences SLOBudgetingMethod = "Occurrences"
	SLOBudgetingMethodTimeslices  SLOBudgetingMethod = "Timeslices"
)

var validSLOBudgetingMethods = []SLOBudgetingMethod{
	SLOBudgetingMethodOccurrences,
	SLOBudgetingMethodTimeslices,
}

// SLOIndicator defines the threshold-metric form of a v1alpha service level
// indicator.
type SLOIndicator struct {
	// ThresholdMetric retrieves raw metric values. Each objective compares them
	// with its [Operator] and Value.
	ThresholdMetric SLOMetricSourceSpec `json:"thresholdMetric"`
}

// SLOMetricSourceSpec describes a provider-specific metric query.
type SLOMetricSourceSpec struct {
	// Source identifies the metric data source.
	Source string `json:"source"`
	// QueryType identifies the query language or query form.
	QueryType string `json:"queryType"`
	// Query is the provider-specific expression that retrieves the metric.
	Query string `json:"query"`
}

// SLOObjective defines a reliability target and, for the ratio form, its metric queries.
type SLOObjective struct {
	// DisplayName is a human-readable objective name.
	DisplayName string `json:"displayName"`
	// Value is the metric threshold used by [Operator].
	Value *float64 `json:"value,omitempty"`
	// RatioMetrics supplies a good-events-to-total-events indicator.
	RatioMetrics *SLORatioMetrics `json:"ratioMetrics"`
	// BudgetTarget is the desired fraction of good events or time slices.
	BudgetTarget *float64 `json:"target"`
	// TimeSliceTarget is the minimum success ratio that makes a time slice good.
	// It is used by the Timeslices budgeting method.
	TimeSliceTarget *float64 `json:"timeSliceTarget,omitempty"`
	// Operator compares values returned by the threshold metric with Value.
	Operator Operator `json:"op,omitempty"`
}

// SLORatioMetrics defines an indicator as the ratio of good events to total
// events. For example, 99 successful requests out of 100 total requests produce
// a ratio of 0.99.
type SLORatioMetrics struct {
	// Good retrieves the numerator: events considered successful.
	Good SLOMetricSourceSpec `json:"good"`
	// Total retrieves the denominator: all considered events.
	Total SLOMetricSourceSpec `json:"total"`
	// Incremental reports whether the queried metrics are monotonically
	// increasing counters rather than values that can rise or fall.
	Incremental bool `json:"incremental"`
}

// SLOTimeWindow defines the period over which an SLO is evaluated. For example,
// a Unit of Week and a Count of 4 define a four-week window. A rolling window
// requires IsRolling to be true and Calendar to be nil. A calendar-aligned
// window requires IsRolling to be false and Calendar to be non-nil.
type SLOTimeWindow struct {
	// Unit combines with Count to set the window length.
	Unit SLOTimeWindowUnit `json:"unit"`
	// Count sets how many Units form the window.
	Count int `json:"count"`
	// IsRolling selects a continuously advancing window when true and a
	// calendar-aligned window when false.
	IsRolling bool `json:"isRolling"`
	// Calendar defines the alignment when IsRolling is false. It must be nil
	// when IsRolling is true.
	Calendar *SLOCalendar `json:"calendar,omitempty"`
}

// SLOTimeWindowUnit identifies the unit used to express an [SLOTimeWindow].
type SLOTimeWindowUnit string

const (
	SLOTimeWindowUnitSecond  SLOTimeWindowUnit = "Second"
	SLOTimeWindowUnitDay     SLOTimeWindowUnit = "Day"
	SLOTimeWindowUnitWeek    SLOTimeWindowUnit = "Week"
	SLOTimeWindowUnitMonth   SLOTimeWindowUnit = "Month"
	SLOTimeWindowUnitQuarter SLOTimeWindowUnit = "Quarter"
)

var validSLOTimeWindowUnits = []SLOTimeWindowUnit{
	SLOTimeWindowUnitSecond,
	SLOTimeWindowUnitDay,
	SLOTimeWindowUnitWeek,
	SLOTimeWindowUnitMonth,
	SLOTimeWindowUnitQuarter,
}

// SLOCalendar anchors a calendar-aligned [SLOTimeWindow].
type SLOCalendar struct {
	// StartTime is the local anchor in YYYY-MM-DD HH:MM:SS format, for example
	// 2026-07-25 14:30:00.
	StartTime string `json:"startTime"`
	// TimeZone is the IANA time-zone name used to interpret StartTime.
	TimeZone string `json:"timeZone"`
}

// Operator selects the comparison between a threshold metric and an objective value.
type Operator string

const (
	OperatorGT  Operator = "gt"
	OperatorLT  Operator = "lt"
	OperatorGTE Operator = "gte"
	OperatorLTE Operator = "lte"
)

var validOperators = []Operator{
	OperatorGT,
	OperatorLT,
	OperatorGTE,
	OperatorLTE,
}

var sloValidation = govy.New(
	validationRulesAPIVersion(func(s SLO) openslo.Version { return s.APIVersion }),
	validationRulesKind(func(s SLO) openslo.Kind { return s.Kind }, openslo.KindSLO),
	validationRulesMetadata(func(s SLO) Metadata { return s.Metadata }),
	govy.For(func(s SLO) SLOSpec { return s.Spec }).
		WithName("spec").
		Required().
		Cascade(govy.CascadeModeStop).
		Rules(
			govy.NewRule(func(s SLOSpec) error {
				hasRatioMetrics := false
				for i := range s.Objectives {
					if s.Objectives[i].RatioMetrics != nil {
						hasRatioMetrics = true
						break
					}
				}
				hasIndicator := s.Indicator != nil
				if hasRatioMetrics && hasIndicator {
					return errors.New("only one of 'indicator' and 'objectives[*].ratioMetrics' can be set")
				}
				if !hasRatioMetrics && !hasIndicator {
					return errors.New("one of 'indicator' or 'objectives[*].ratioMetrics' must be set")
				}
				return nil
			}).
				WithDescription("exactly one of 'indicator' and 'objectives[*].ratioMetrics' must be set").
				WithErrorCode(rules.ErrorCodeMutuallyExclusive),
		).
		Include(sloSpecValidation),
).WithNameFunc(internal.GetObjectName[SLO])

var sloSpecValidation = govy.New(
	govy.For(govy.GetSelf[SLOSpec]()).
		Include(sloTimeSlicesObjectiveValidation),
	govy.For(func(spec SLOSpec) string { return spec.Description }).
		WithName("description").
		OmitEmpty().
		Rules(rules.StringMaxLength(1050)),
	govy.For(func(spec SLOSpec) string { return spec.Service }).
		WithName("service").
		Required(),
	govy.ForPointer(func(spec SLOSpec) *SLOIndicator { return spec.Indicator }).
		WithName("indicator").
		Include(sloIndicatorValidation),
	govy.For(func(spec SLOSpec) SLOBudgetingMethod { return spec.BudgetingMethod }).
		WithName("budgetingMethod").
		Required().
		Rules(rules.OneOf(validSLOBudgetingMethods...)),
	govy.ForSlice(func(spec SLOSpec) []SLOTimeWindow { return spec.TimeWindows }).
		WithName("timeWindows").
		Rules(rules.SliceLength[[]SLOTimeWindow](1, 1)).
		IncludeForEach(sloTimeWindowValidation),
	govy.ForSlice(func(spec SLOSpec) []SLOObjective { return spec.Objectives }).
		WithName("objectives").
		IncludeForEach(sloObjectiveValidation),
).Cascade(govy.CascadeModeContinue)

var sloIndicatorValidation = govy.New(
	govy.For(func(i SLOIndicator) SLOMetricSourceSpec { return i.ThresholdMetric }).
		WithName("thresholdMetric").
		Required().
		Include(sloMetricSourceSpecValidation),
)

var sloTimeWindowValidation = govy.New(
	govy.For(govy.GetSelf[SLOTimeWindow]()).
		Rules(govy.NewRule(func(s SLOTimeWindow) error {
			if s.IsRolling && s.Calendar != nil {
				return govy.NewRuleError("'calendar' cannot be set when 'isRolling' is true")
			}
			if !s.IsRolling && s.Calendar == nil {
				return govy.NewRuleError("'calendar' must be set when 'isRolling' is false")
			}
			return nil
		}).WithDescription(
			"'calendar' must be set when 'isRolling' is false and cannot be set when 'isRolling' is true",
		)),
	govy.For(func(t SLOTimeWindow) SLOTimeWindowUnit { return t.Unit }).
		WithName("unit").
		Required().
		Rules(rules.OneOf(validSLOTimeWindowUnits...)),
	govy.For(func(t SLOTimeWindow) int { return t.Count }).
		WithName("count").
		Rules(rules.GT(0)),
	govy.ForPointer(func(t SLOTimeWindow) *SLOCalendar { return t.Calendar }).
		WithName("calendar").
		Include(govy.New(
			govy.For(func(c SLOCalendar) string { return c.StartTime }).
				WithName("startTime").
				Rules(rules.StringDateTime(time.DateTime)),
			govy.For(func(c SLOCalendar) string { return c.TimeZone }).
				WithName("timeZone").
				Rules(rules.StringTimeZone()),
		)),
)

var sloObjectiveValidation = govy.New(
	govy.For(func(s SLOObjective) string { return s.DisplayName }).
		WithName("displayName").
		Rules(rules.StringMaxLength(1050)),
	govy.ForPointer(func(s SLOObjective) *SLORatioMetrics { return s.RatioMetrics }).
		WithName("ratioMetrics").
		Include(sloRatioMetricsValidation),
	govy.ForPointer(func(s SLOObjective) *float64 { return s.Value }).
		WithName("value").
		When(
			func(s SLOObjective) bool { return s.RatioMetrics == nil },
			govy.WhenDescription("'ratioMetrics' is not set"),
		).
		Required(),
	govy.ForPointer(func(s SLOObjective) *float64 { return s.BudgetTarget }).
		WithName("target").
		Required().
		Rules(rules.GTE(0.0), rules.LT(1.0)),
	govy.For(func(s SLOObjective) Operator { return s.Operator }).
		WithName("op").
		When(
			func(s SLOObjective) bool { return s.RatioMetrics == nil },
			govy.WhenDescription("'ratioMetrics' is not set"),
		).
		Required().
		Rules(rules.OneOf(validOperators...)),
	govy.For(func(s SLOObjective) Operator { return s.Operator }).
		WithName("op").
		When(
			func(s SLOObjective) bool { return s.RatioMetrics != nil },
			govy.WhenDescription("'ratioMetrics' is set"),
		).
		Rules(rules.Forbidden[Operator]()),
)

var sloTimeSlicesObjectiveValidation = govy.New(
	govy.ForSlice(func(spec SLOSpec) []SLOObjective { return spec.Objectives }).
		WithName("objectives").
		IncludeForEach(govy.New(
			govy.ForPointer(func(s SLOObjective) *float64 { return s.TimeSliceTarget }).
				WithName("timeSliceTarget").
				Required().
				Rules(rules.GTE(0.0), rules.LTE(1.0)),
		)),
).
	When(
		func(s SLOSpec) bool { return s.BudgetingMethod == SLOBudgetingMethodTimeslices },
		govy.WhenDescription("'budgetingMethod' is 'Timeslices'"),
	)

var sloRatioMetricsValidation = govy.New(
	govy.For(func(s SLORatioMetrics) SLOMetricSourceSpec { return s.Good }).
		WithName("good").
		Required().
		Include(sloMetricSourceSpecValidation),
	govy.For(func(s SLORatioMetrics) SLOMetricSourceSpec { return s.Total }).
		WithName("total").
		Required().
		Include(sloMetricSourceSpecValidation),
)

var sloMetricSourceSpecValidation = govy.New(
	govy.For(func(s SLOMetricSourceSpec) string { return s.Source }).
		WithName("source").
		Required().
		Rules(rules.StringAlpha()),
	govy.For(func(s SLOMetricSourceSpec) string { return s.QueryType }).
		WithName("queryType").
		Required().
		Rules(rules.StringAlpha()),
	govy.For(func(s SLOMetricSourceSpec) string { return s.Query }).
		WithName("query").
		Required().
		Rules(rules.StringNotEmpty()),
)
