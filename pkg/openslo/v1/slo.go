package v1

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

// SLO represents a target value or range for a service level measured by an
// [SLI].
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

// GetName returns the name in the SLO's [Metadata].
func (s SLO) GetName() string {
	return s.Metadata.Name
}

// Validate returns an error for an invalid SLO.
func (s SLO) Validate() error {
	return sloValidation.Validate(s)
}

// String returns the SLO's formatted version, kind, and name.
func (s SLO) String() string {
	return internal.GetObjectName(s)
}

// IsComposite reports whether the SLO has objective-level indicators.
func (s SLO) IsComposite() bool {
	return s.Spec.HasCompositeObjectives()
}

// GetMetadata returns the SLO's [Metadata].
func (s SLO) GetMetadata() Metadata {
	return s.Metadata
}

// GetValidator returns the validator for SLO objects.
func (s SLO) GetValidator() govy.Validator[SLO] {
	return sloValidation
}

// SLOSpec defines the service association, indicator placement, budgeting
// method, evaluation window, objectives, and alert policies of an [SLO].
type SLOSpec struct {
	// Description summarizes the SLO.
	Description string `json:"description,omitempty"`
	// Service names the associated service. Consumers define how to resolve the
	// name to a [Service].
	Service string `json:"service"`
	// Indicator defines a standard SLO's SLI inline. Composite SLOs place
	// indicators on individual Objectives.
	Indicator *SLOIndicatorInline `json:"indicator,omitempty"`
	// IndicatorRef names an existing [SLI] for a standard SLO. Composite SLOs
	// place indicator references on individual Objectives.
	IndicatorRef *string `json:"indicatorRef,omitempty"`
	// BudgetingMethod applies the selected error-budget calculation to every
	// objective.
	BudgetingMethod SLOBudgetingMethod `json:"budgetingMethod"`
	// TimeWindow contains the SLO's evaluation window.
	TimeWindow []SLOTimeWindow `json:"timeWindow,omitempty"`
	// Objectives contains the SLO's target definitions.
	Objectives []SLOObjective `json:"objectives"`
	// AlertPolicies contains inline alert policies or references to existing
	// [AlertPolicy] objects.
	AlertPolicies []SLOAlertPolicy `json:"alertPolicies,omitempty"`
}

// HasCompositeObjectives reports whether any objective has an indicator.
// It does not verify that every objective in a composite SLO has one.
func (s SLOSpec) HasCompositeObjectives() bool {
	for i := range s.Objectives {
		if s.Objectives[i].Indicator != nil || s.Objectives[i].IndicatorRef != nil {
			return true
		}
	}
	return false
}

// SLOBudgetingMethod identifies how an [SLO] calculates objective success and
// error-budget use.
// Occurrences uses the ratio of good events to total events. Timeslices counts
// slices that meet [SLOObjective.TimeSliceTarget]. RatioTimeslices averages
// success ratios across slices. Composite SLOs apply objective weights to each
// calculation.
type SLOBudgetingMethod string

const (
	// SLOBudgetingMethodOccurrences uses the ratio of good events to total events,
	// so traffic volume determines each period's influence.
	// For a composite SLO, each objective's weight scales its burn rate.
	SLOBudgetingMethodOccurrences SLOBudgetingMethod = "Occurrences"
	// SLOBudgetingMethodTimeslices uses the ratio of slices meeting
	// [SLOObjective.TimeSliceTarget] to all slices, giving each slice equal
	// influence. Any bad objective makes a composite slice bad.
	SLOBudgetingMethodTimeslices SLOBudgetingMethod = "Timeslices"
	// SLOBudgetingMethodRatioTimeslices averages success ratios across slices
	// without classifying them against [SLOObjective.TimeSliceTarget].
	// For a composite SLO, this method combines weighted deficits from 100 percent.
	SLOBudgetingMethodRatioTimeslices SLOBudgetingMethod = "RatioTimeslices"
)

var validSLOBudgetingMethods = []SLOBudgetingMethod{
	SLOBudgetingMethodOccurrences,
	SLOBudgetingMethodTimeslices,
	SLOBudgetingMethodRatioTimeslices,
}

// SLOIndicatorInline embeds an [SLI] in an [SLOSpec] or [SLOObjective].
type SLOIndicatorInline struct {
	Metadata Metadata `json:"metadata"`
	Spec     SLISpec  `json:"spec"`
}

// SLOObjective defines a success target and, when applicable, a threshold
// comparison or composite-specific indicator. For example, Target 0.995 and
// TargetPercent 99.5 both express a 99.5 percent target.
type SLOObjective struct {
	// DisplayName is the objective's human-readable name.
	DisplayName string `json:"displayName,omitempty"`
	// Operator compares threshold-metric samples with Value.
	Operator Operator `json:"op,omitempty"`
	// Value sets the threshold for metric sample comparisons.
	// It is distinct from the success target expressed by Target or TargetPercent.
	Value *float64 `json:"value,omitempty"`
	// Target expresses the success target as a fraction.
	Target *float64 `json:"target,omitempty"`
	// TargetPercent expresses the success target as a percentage.
	TargetPercent *float64 `json:"targetPercent,omitempty"`
	// TimeSliceTarget classifies a slice as good when BudgetingMethod is
	// [SLOBudgetingMethodTimeslices].
	TimeSliceTarget *float64 `json:"timeSliceTarget,omitempty"`
	// TimeSliceWindow sets the slice size and query-evaluation interval for
	// [SLOBudgetingMethodTimeslices] and [SLOBudgetingMethodRatioTimeslices].
	TimeSliceWindow *DurationShorthand `json:"timeSliceWindow,omitempty"`
	// Indicator defines this objective's SLI inline for a composite SLO.
	Indicator *SLOIndicatorInline `json:"indicator,omitempty"`
	// IndicatorRef names this objective's [SLI] for a composite SLO.
	IndicatorRef *string `json:"indicatorRef,omitempty"`
	// CompositeWeight scales this objective's contribution to a multi-objective
	// composite SLO. OpenSLO uses a weight of 1 when omitted. This SDK leaves the
	// field unset.
	CompositeWeight *float64 `json:"compositeWeight,omitempty"`
}

// SLOTimeWindow defines either a rolling or calendar-aligned evaluation window.
type SLOTimeWindow struct {
	// Duration is the length of the evaluation window.
	Duration DurationShorthand `json:"duration"`
	// IsRolling selects a rolling window when true and a calendar-aligned window
	// when false.
	IsRolling bool `json:"isRolling"`
	// Calendar defines the alignment of a calendar window.
	Calendar *SLOCalendar `json:"calendar,omitempty"`
}

// SLOCalendar anchors a calendar-aligned [SLOTimeWindow] in a time zone.
type SLOCalendar struct {
	// StartTime anchors the first calendar window.
	StartTime string `json:"startTime"`
	// TimeZone controls the interpretation of StartTime and later boundaries.
	TimeZone string `json:"timeZone"`
}

// SLOAlertPolicy associates an alert policy with an [SLO].
type SLOAlertPolicy struct {
	*SLOAlertPolicyInline
	*SLOAlertPolicyRef
}

// SLOAlertPolicyInline is an [AlertPolicy] embedded in an [SLO].
type SLOAlertPolicyInline struct {
	Kind     openslo.Kind    `json:"kind"`
	Metadata Metadata        `json:"metadata"`
	Spec     AlertPolicySpec `json:"spec"`
}

// SLOAlertPolicyRef refers to an existing [AlertPolicy].
type SLOAlertPolicyRef struct {
	// AlertPolicyRef names an existing alert policy.
	AlertPolicyRef string `json:"alertPolicyRef"`
}

var sloValidation = govy.New(
	validationRulesAPIVersion(func(s SLO) openslo.Version { return s.APIVersion }),
	validationRulesKind(func(s SLO) openslo.Kind { return s.Kind }, openslo.KindSLO),
	validationRulesMetadata(func(s SLO) Metadata { return s.Metadata }),
	govy.For(func(s SLO) SLOSpec { return s.Spec }).
		WithName("spec").
		Required().
		Include(sloSpecValidation),
).WithNameFunc(internal.GetObjectName[SLO])

var sloSpecValidation = govy.New(
	govy.For(govy.GetSelf[SLOSpec]()).
		Rules(validationRuleForIndicator()).
		Include(
			getSLOIndicatorValidation(
				func(s SLOSpec) *SLOIndicatorInline { return s.Indicator },
				func(s SLOSpec) *string { return s.IndicatorRef },
			),
			sloTimeSlicesObjectiveValidation,
			sloRatioTimeSlicesObjectiveValidation,
		),
	govy.For(func(spec SLOSpec) string { return spec.Description }).
		WithName("description").
		Rules(rules.StringMaxLength(1050)),
	govy.For(func(spec SLOSpec) string { return spec.Service }).
		WithName("service").
		Required(),
	govy.For(func(spec SLOSpec) SLOBudgetingMethod { return spec.BudgetingMethod }).
		WithName("budgetingMethod").
		Required().
		Rules(rules.OneOf(validSLOBudgetingMethods...)),
	govy.ForSlice(func(spec SLOSpec) []SLOTimeWindow { return spec.TimeWindow }).
		WithName("timeWindow").
		Rules(rules.SliceLength[[]SLOTimeWindow](1, 1)).
		IncludeForEach(sloTimeWindowValidation),
	govy.ForSlice(func(spec SLOSpec) []SLOAlertPolicy { return spec.AlertPolicies }).
		WithName("alertPolicies").
		IncludeForEach(sloAlertPolicyValidation),
	sloObjectivesProperty.
		IncludeForEach(sloObjectiveValidation),
	sloObjectivesProperty.
		IncludeForEach(sloRatioObjectiveValidationWhenInlinedSLI).
		When(
			func(s SLOSpec) bool { return s.Indicator != nil && s.Indicator.Spec.RatioMetric != nil },
			govy.WhenDescription("'indicator.spec.ratioMetric' is set"),
		),
	sloObjectivesProperty.
		IncludeForEach(sloThresholdObjectiveValidationWhenInlinedSLI).
		When(
			func(s SLOSpec) bool { return s.Indicator != nil && s.Indicator.Spec.ThresholdMetric != nil },
			govy.WhenDescription("'indicator.spec.thresholdMetric' is set"),
		),
	govy.ForSlice(func(spec SLOSpec) []SLOObjective { return spec.Objectives }).
		WithName("objectives").
		When(
			func(s SLOSpec) bool { return s.HasCompositeObjectives() },
			govy.WhenDescription("is composite SLO"),
		).
		IncludeForEach(sloCompositeObjectiveValidation),
)

func getSLOIndicatorValidation[T any](
	indicatorGetter func(T) *SLOIndicatorInline,
	indicatorRefGetter func(T) *string,
) govy.Validator[T] {
	return govy.New(
		govy.For(govy.GetSelf[T]()).
			Rules(rules.MutuallyExclusive(true, map[string]func(t T) any{
				"indicator":    func(t T) any { return indicatorGetter(t) },
				"indicatorRef": func(t T) any { return indicatorRefGetter(t) },
			}).WithDescription("exactly one of 'indicator' and 'indicatorRef' must be set")),
		govy.ForPointer(indicatorGetter).
			WithName("indicator").
			Cascade(govy.CascadeModeContinue).
			Include(govy.New(
				validationRulesMetadata(func(s SLOIndicatorInline) Metadata { return s.Metadata }),
				govy.For(func(s SLOIndicatorInline) SLISpec { return s.Spec }).
					WithName("spec").
					Include(sliSpecValidation),
			)),
		govy.ForPointer(indicatorRefGetter).
			WithName("indicatorRef").
			Rules(rules.StringDNSLabel()),
	).
		// Another validation rule on 'spec' level already checks a scenario
		// in which neither 'indicator' nor 'indicatorRef' are provided.
		When(
			func(t T) bool { return indicatorGetter(t) != nil || indicatorRefGetter(t) != nil },
			govy.WhenDescription("'indicator' or 'indicatorRef' is set"),
		).
		Cascade(govy.CascadeModeStop)
}

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
		})),
	govy.For(func(t SLOTimeWindow) DurationShorthand { return t.Duration }).
		WithName("duration").
		Required().
		Include(durationShortHandValidation),
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

var sloAlertPolicyValidation = govy.New(
	govy.For(govy.GetSelf[SLOAlertPolicy]()).
		Rules(rules.MutuallyExclusive(true, map[string]func(a SLOAlertPolicy) any{
			"alertPolicyRef": func(a SLOAlertPolicy) any { return a.SLOAlertPolicyRef },
			// It's impossible to list all fields that constitute the inlined version in the error message,
			// therefore 'spec' must suffice.
			"spec": func(a SLOAlertPolicy) any { return a.SLOAlertPolicyInline },
		}).WithDescription("exactly one of 'alertPolicyRef' and 'spec' must be set")),
	govy.ForPointer(func(a SLOAlertPolicy) *SLOAlertPolicyRef {
		return a.SLOAlertPolicyRef
	}).
		Include(govy.New(
			govy.For(func(ref SLOAlertPolicyRef) string { return ref.AlertPolicyRef }).
				WithName("alertPolicyRef").
				Required().
				Rules(rules.StringDNSLabel()),
		)).Cascade(govy.CascadeModeContinue),
	govy.ForPointer(func(a SLOAlertPolicy) *SLOAlertPolicyInline {
		return a.SLOAlertPolicyInline
	}).
		Include(govy.New(
			govy.For(func(inline SLOAlertPolicyInline) openslo.Kind { return inline.Kind }).
				WithName("kind").
				Required().
				Rules(rules.EQ(openslo.KindAlertPolicy)),
			validationRulesMetadata(func(a SLOAlertPolicyInline) Metadata { return a.Metadata }),
			govy.For(func(inline SLOAlertPolicyInline) AlertPolicySpec { return inline.Spec }).
				WithName("spec").
				Required().
				Include(alertPolicySpecValidation),
		)).Cascade(govy.CascadeModeContinue),
).Cascade(govy.CascadeModeStop)

var sloObjectivesProperty = govy.ForSlice(func(spec SLOSpec) []SLOObjective { return spec.Objectives }).
	WithName("objectives")

var sloObjectiveValidation = govy.New(
	govy.For(govy.GetSelf[SLOObjective]()).
		Rules(rules.MutuallyExclusive(true, map[string]func(o SLOObjective) any{
			"target":        func(o SLOObjective) any { return o.Target },
			"targetPercent": func(o SLOObjective) any { return o.TargetPercent },
		}).WithDescription("exactly one of 'target' and 'targetPercent' must be set")),
	govy.ForPointer(func(s SLOObjective) *float64 { return s.Target }).
		WithName("target").
		Rules(rules.GTE(0.0), rules.LT(1.0)),
	govy.ForPointer(func(s SLOObjective) *float64 { return s.TargetPercent }).
		WithName("targetPercent").
		Rules(rules.GTE(0.0), rules.LT(100.0)),
)

// Since operator and value are only required when using threshold metric SLI
// we have no way of checking it if the SLI is only referenced and not inlined.
var sloThresholdObjectiveValidationWhenInlinedSLI = govy.New(
	govy.ForPointer(func(s SLOObjective) *float64 { return s.Value }).
		WithName("value").
		Required(),
	govy.For(func(s SLOObjective) Operator { return s.Operator }).
		WithName("op").
		Required().
		Include(operatorValidation),
)

var sloRatioObjectiveValidationWhenInlinedSLI = govy.New(
	govy.For(func(s SLOObjective) *float64 { return s.Value }).
		WithName("value").
		Rules(rules.Forbidden[*float64]()),
	govy.For(func(s SLOObjective) Operator { return s.Operator }).
		WithName("op").
		Rules(rules.Forbidden[Operator]()),
)

var sloCompositeObjectiveValidation = govy.New(
	govy.For(govy.GetSelf[SLOObjective]()).
		Include(
			getSLOIndicatorValidation(
				func(s SLOObjective) *SLOIndicatorInline { return s.Indicator },
				func(s SLOObjective) *string { return s.IndicatorRef },
			),
		),
	govy.ForPointer(func(s SLOObjective) *float64 { return s.CompositeWeight }).
		WithName("compositeWeight").
		Rules(rules.GT(0.0)),
)

var sloTimeSlicesObjectiveValidation = govy.New(
	govy.ForSlice(func(spec SLOSpec) []SLOObjective { return spec.Objectives }).
		WithName("objectives").
		IncludeForEach(govy.New(
			govy.ForPointer(func(s SLOObjective) *float64 { return s.TimeSliceTarget }).
				WithName("timeSliceTarget").
				Required().
				Rules(rules.GT(0.0), rules.LTE(1.0)),
			validationRulesForTimeSliceWindow(),
		)),
).
	When(
		func(s SLOSpec) bool { return s.BudgetingMethod == SLOBudgetingMethodTimeslices },
		govy.WhenDescription("'budgetingMethod' is 'Timeslices'"),
	)

var sloRatioTimeSlicesObjectiveValidation = govy.New(
	govy.ForSlice(func(spec SLOSpec) []SLOObjective { return spec.Objectives }).
		WithName("objectives").
		IncludeForEach(govy.New(
			validationRulesForTimeSliceWindow(),
		)),
).
	When(
		func(s SLOSpec) bool { return s.BudgetingMethod == SLOBudgetingMethodRatioTimeslices },
		govy.WhenDescription("'budgetingMethod' is 'RatioTimeslices'"),
	)

func validationRulesForTimeSliceWindow() govy.PropertyRules[DurationShorthand, SLOObjective] {
	return govy.ForPointer(func(s SLOObjective) *DurationShorthand { return s.TimeSliceWindow }).
		WithName("timeSliceWindow").
		Required().
		Include(durationShortHandValidation)
}

func validationRuleForIndicator() govy.Rule[SLOSpec] {
	msg := "'indicator' or 'indicatorRef' fields must either be defined on the 'spec' level (standard SLOs)" +
		" or on the 'spec.objectives[*]' level (composite SLOs)"
	return govy.NewRule(func(s SLOSpec) error {
		hasComposites := s.HasCompositeObjectives()
		hasIndicator := s.Indicator != nil || s.IndicatorRef != nil
		if !hasComposites && !hasIndicator {
			return errors.New(msg + ", but none were provided")
		}
		if hasComposites && hasIndicator {
			return errors.New(msg + ", but not both")
		}
		return nil
	}).
		WithErrorCode(rules.ErrorCodeMutuallyExclusive).
		WithDescription(msg)
}
