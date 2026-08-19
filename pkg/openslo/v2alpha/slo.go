package v2alpha

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

// SLO defines a target for an SLI over a time window.
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

// String returns the SLO's formatted version and kind.
// It also returns the metadata name when set.
func (s SLO) String() string {
	return internal.GetObjectName(s)
}

// GetMetadata returns the SLO's metadata.
func (s SLO) GetMetadata() Metadata {
	return s.Metadata
}

// IsComposite reports whether at least one objective selects its own SLI.
func (s SLO) IsComposite() bool {
	return s.Spec.HasCompositeObjectives()
}

// GetValidator returns the validator configured for [SLO].
func (s SLO) GetValidator() govy.Validator[SLO] {
	return sloValidation
}

// SLOSpec defines an SLO's service, SLI, time window, budgeting method,
// objectives, and alert policies. A standard SLO applies one SLI to all
// objectives. A composite SLO can select a different SLI per objective.
type SLOSpec struct {
	// Description optionally summarizes the SLO in at most 1,050 characters.
	Description string `json:"description,omitempty"`
	// ServiceRef names the service associated with this SLO. The SDK serializes
	// the field as "serviceRef". The living v2alpha proposal calls it "service".
	ServiceRef string `json:"serviceRef"`
	// SLI embeds the service level indicator for a standard SLO.
	SLI *SLOSLIInline `json:"sli,omitempty"`
	// SLIRef names an existing [SLI] for a standard SLO.
	SLIRef *string `json:"sliRef,omitempty"`
	// BudgetingMethod applies the selected error-budget calculation to every
	// objective.
	BudgetingMethod SLOBudgetingMethod `json:"budgetingMethod"`
	// TimeWindow contains exactly one SLO evaluation window.
	TimeWindow []SLOTimeWindow `json:"timeWindow,omitempty"`
	// Objectives contains the SLO's budget targets and metric thresholds.
	// V2alpha permits multiple objectives for a threshold-metric SLO.
	// This SDK accepts an omitted Objectives field.
	Objectives []SLOObjective `json:"objectives"`
	// AlertPolicies contains policies associated with the SLO.
	// Each item must specify exactly one inline definition or metadata-name
	// reference.
	AlertPolicies []SLOAlertPolicy `json:"alertPolicies,omitempty"`
}

// HasCompositeObjectives reports whether at least one objective selects an SLI
// inline or by reference. It does not verify that every composite objective
// selects one.
func (s SLOSpec) HasCompositeObjectives() bool {
	for i := range s.Objectives {
		if s.Objectives[i].SLI != nil || s.Objectives[i].SLIRef != nil {
			return true
		}
	}
	return false
}

// SLOBudgetingMethod selects how an SLO consumes its error budget.
// Occurrences uses good events over total events, Timeslices uses good slices
// over total slices, and RatioTimeslices averages slice success ratios.
type SLOBudgetingMethod string

const (
	SLOBudgetingMethodOccurrences     SLOBudgetingMethod = "Occurrences"
	SLOBudgetingMethodTimeslices      SLOBudgetingMethod = "Timeslices"
	SLOBudgetingMethodRatioTimeslices SLOBudgetingMethod = "RatioTimeslices"
)

var validSLOBudgetingMethods = []SLOBudgetingMethod{
	SLOBudgetingMethodOccurrences,
	SLOBudgetingMethodTimeslices,
	SLOBudgetingMethodRatioTimeslices,
}

// SLOSLIInline embeds an SLI definition in an SLO or one of its objectives.
type SLOSLIInline struct {
	Metadata Metadata `json:"metadata"`
	Spec     SLISpec  `json:"spec"`
}

// SLOObjective defines one error-budget target and, for a threshold SLI, its
// metric comparison. The living v2alpha proposal also defines objective labels,
// which this SDK does not model.
//
// For a standard SLO with an inline threshold SLI, validation requires
// [SLOObjective.Operator] and [SLOObjective.Value]. For a standard SLO with an
// inline ratio SLI, validation forbids them. The SDK does not apply these
// metric-type rules to referenced SLIs or to SLIs embedded in composite
// objectives.
type SLOObjective struct {
	// DisplayName is a human-readable name for this objective.
	// It is not part of the enclosing object's [Metadata].
	DisplayName string `json:"displayName,omitempty"`
	// Operator compares a threshold metric with Value.
	Operator Operator `json:"op,omitempty"`
	// Value is the comparison threshold for a threshold metric.
	Value *float64 `json:"value,omitempty"`
	// Target is the desired success proportion. For example, 0.995 means
	// 99.5 percent.
	Target *float64 `json:"target,omitempty"`
	// TargetPercent is the desired success percentage.
	TargetPercent *float64 `json:"targetPercent,omitempty"`
	// TimeSliceTarget sets the per-slice success threshold for Timeslices.
	TimeSliceTarget *float64 `json:"timeSliceTarget,omitempty"`
	// TimeSliceWindow sets the size of each slice for Timeslices and
	// RatioTimeslices. OpenSLO also accepts a number interpreted as minutes. This
	// SDK represents only duration shorthand.
	TimeSliceWindow *DurationShorthand `json:"timeSliceWindow,omitempty"`
	// SLI embeds this objective's service level indicator for a composite SLO.
	SLI *SLOSLIInline `json:"sli,omitempty"`
	// SLIRef names this objective's existing [SLI] for a composite SLO.
	SLIRef *string `json:"sliRef,omitempty"`
	// CompositeWeight scales this objective's contribution to a composite SLO.
	// The living v2alpha proposal permits it only with multiple objectives and
	// defaults it to 1. This SDK does not enforce the objective-count restriction
	// and preserves an omitted value as nil.
	CompositeWeight *float64 `json:"compositeWeight,omitempty"`
}

// SLOTimeWindow describes one rolling or calendar-aligned evaluation window.
// If [SLOTimeWindow.IsRolling] is true, [SLOTimeWindow.Calendar] must be nil.
// If IsRolling is false, Calendar must be non-nil.
type SLOTimeWindow struct {
	// Duration is the length of the evaluation window.
	Duration DurationShorthand `json:"duration"`
	// IsRolling selects a rolling window when true and a calendar-aligned window
	// when false.
	IsRolling bool `json:"isRolling"`
	// Calendar defines the alignment of a calendar window.
	Calendar *SLOCalendar `json:"calendar,omitempty"`
}

// SLOCalendar defines the starting wall-clock time and time zone for a
// calendar-aligned [SLOTimeWindow].
type SLOCalendar struct {
	// StartTime is the local date and time when calendar alignment starts.
	StartTime string `json:"startTime"`
	// TimeZone determines how StartTime maps to an instant.
	TimeZone string `json:"timeZone"`
}

// SLOAlertPolicy associates exactly one inline or referenced alert policy with
// an [SLO].
type SLOAlertPolicy struct {
	*SLOAlertPolicyInline
	*SLOAlertPolicyRef
}

// SLOAlertPolicyInline is an alert-policy definition embedded in an SLO.
// The inline form contains kind, metadata, and spec, but no API version.
type SLOAlertPolicyInline struct {
	Kind     openslo.Kind    `json:"kind"`
	Metadata Metadata        `json:"metadata"`
	Spec     AlertPolicySpec `json:"spec"`
}

// SLOAlertPolicyRef identifies a separately defined [AlertPolicy].
type SLOAlertPolicyRef struct {
	// AlertPolicyRef is the metadata name of the alert policy to use.
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
		Rules(validationRuleForSLOSLI()).
		Include(
			getSLOSLIValidation(
				func(s SLOSpec) *SLOSLIInline { return s.SLI },
				func(s SLOSpec) *string { return s.SLIRef },
			),
			sloTimeSlicesObjectiveValidation,
			sloRatioTimeSlicesObjectiveValidation,
		),
	govy.For(func(spec SLOSpec) string { return spec.Description }).
		WithName("description").
		OmitEmpty().
		Rules(rules.StringMaxLength(1050)),
	govy.For(func(spec SLOSpec) string { return spec.ServiceRef }).
		WithName("serviceRef").
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
		IncludeForEach(sloCompositeObjectiveValidation).
		When(
			func(s SLOSpec) bool { return s.HasCompositeObjectives() },
			govy.WhenDescription("is composite SLO"),
		),
	sloObjectivesProperty.
		IncludeForEach(sloRatioObjectiveValidationWhenInlinedSLI).
		When(
			func(s SLOSpec) bool { return s.SLI != nil && s.SLI.Spec.RatioMetric != nil },
			govy.WhenDescription("'sli.spec.ratioMetric' is set"),
		),
	sloObjectivesProperty.
		IncludeForEach(sloThresholdObjectiveValidationWhenInlinedSLI).
		When(
			func(s SLOSpec) bool { return s.SLI != nil && s.SLI.Spec.ThresholdMetric != nil },
			govy.WhenDescription("'sli.spec.thresholdMetric' is set"),
		),
)

func getSLOSLIValidation[T any](
	sliGetter func(T) *SLOSLIInline,
	sliRefGetter func(T) *string,
) govy.Validator[T] {
	return govy.New(
		govy.For(govy.GetSelf[T]()).
			Rules(rules.MutuallyExclusive(true, map[string]func(t T) any{
				"sli":    func(t T) any { return sliGetter(t) },
				"sliRef": func(t T) any { return sliRefGetter(t) },
			}).WithDescription("exactly one of 'sli' and 'sliRef' must be set")),
		govy.ForPointer(sliGetter).
			WithName("sli").
			Cascade(govy.CascadeModeContinue).
			Include(govy.New(
				validationRulesMetadata(func(s SLOSLIInline) Metadata { return s.Metadata }),
				govy.For(func(s SLOSLIInline) SLISpec { return s.Spec }).
					WithName("spec").
					Include(sliSpecValidation),
			)),
		govy.ForPointer(sliRefGetter).
			WithName("sliRef").
			Rules(rules.StringDNSLabel()),
	).
		// Another validation rule on 'spec' level already checks a scenario
		// in which neither 'sli' nor 'sliRef' are provided.
		When(
			func(t T) bool { return sliGetter(t) != nil || sliRefGetter(t) != nil },
			govy.WhenDescription("'sli' or 'sliRef' is set"),
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
		}).WithDescription(
			"'calendar' must be set when 'isRolling' is false and cannot be set when 'isRolling' is true",
		)),
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

var sloCompositeObjectiveValidation = govy.New(
	govy.For(govy.GetSelf[SLOObjective]()).
		Include(
			getSLOSLIValidation(
				func(s SLOObjective) *SLOSLIInline { return s.SLI },
				func(s SLOObjective) *string { return s.SLIRef },
			),
		),
	govy.ForPointer(func(s SLOObjective) *float64 { return s.CompositeWeight }).
		WithName("compositeWeight").
		Rules(rules.GT(0.0)),
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

func validationRuleForSLOSLI() govy.Rule[SLOSpec] {
	msg := "'sli' or 'sliRef' fields must either be defined on the 'spec' level (standard SLOs)" +
		" or on the 'spec.objectives[*]' level (composite SLOs)"
	return govy.NewRule(func(s SLOSpec) error {
		hasComposites := s.HasCompositeObjectives()
		hasSLI := s.SLI != nil || s.SLIRef != nil
		if !hasComposites && !hasSLI {
			return errors.New(msg + ", but none were provided")
		}
		if hasComposites && hasSLI {
			return errors.New(msg + ", but not both")
		}
		return nil
	}).
		WithErrorCode(rules.ErrorCodeMutuallyExclusive).
		WithDescription(msg)
}
