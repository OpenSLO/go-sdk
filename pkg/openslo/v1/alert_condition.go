package v1

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/OpenSLO/go-sdk/internal"
	"github.com/OpenSLO/go-sdk/pkg/openslo"
)

var (
	_ = Object(AlertCondition{})
	_ = openslo.ObjectValidator[AlertCondition](AlertCondition{})
)

// NewAlertCondition returns an AlertCondition from metadata and spec.
func NewAlertCondition(metadata Metadata, spec AlertConditionSpec) AlertCondition {
	return AlertCondition{
		APIVersion: APIVersion,
		Kind:       openslo.KindAlertCondition,
		Metadata:   metadata,
		Spec:       spec,
	}
}

// AlertCondition defines the condition under which an SLO alert is considered
// breaching.
type AlertCondition struct {
	APIVersion openslo.Version    `json:"apiVersion"`
	Kind       openslo.Kind       `json:"kind"`
	Metadata   Metadata           `json:"metadata"`
	Spec       AlertConditionSpec `json:"spec"`
}

// GetVersion returns [APIVersion].
func (a AlertCondition) GetVersion() openslo.Version {
	return APIVersion
}

// GetKind returns [openslo.KindAlertCondition].
func (a AlertCondition) GetKind() openslo.Kind {
	return openslo.KindAlertCondition
}

// GetName returns the name in the AlertCondition's [Metadata].
func (a AlertCondition) GetName() string {
	return a.Metadata.Name
}

// Validate returns an error for an invalid alert condition.
func (a AlertCondition) Validate() error {
	return alertConditionValidation.Validate(a)
}

// String returns the AlertCondition's formatted version, kind, and name.
func (a AlertCondition) String() string {
	return internal.GetObjectName(a)
}

// GetMetadata returns the AlertCondition's [Metadata].
func (a AlertCondition) GetMetadata() Metadata {
	return a.Metadata
}

// GetValidator returns the validator for AlertCondition objects.
func (a AlertCondition) GetValidator() govy.Validator[AlertCondition] {
	return alertConditionValidation
}

// AlertConditionSpec defines an alert's severity and burn-rate condition.
type AlertConditionSpec struct {
	// Severity is an implementation-defined classification such as "sev1" or "page".
	Severity  string             `json:"severity"`
	Condition AlertConditionType `json:"condition"`
	// Description summarizes the alert condition.
	Description string `json:"description,omitempty"`
}

// AlertConditionType defines a comparison against an SLO's burn rate. Burn rate
// is error-budget consumption relative to the rate allowed by the SLO.
type AlertConditionType struct {
	// Kind selects the condition calculation. OpenSLO defaults Kind to
	// [AlertConditionKindBurnRate]. This SDK does not apply that default.
	Kind AlertConditionKind `json:"kind"`
	// Operator compares the calculated burn rate with Threshold.
	Operator Operator `json:"op"`
	// Threshold sets the numeric burn-rate boundary.
	Threshold *float64 `json:"threshold"`
	// LookbackWindow sets the period for burn-rate calculation.
	LookbackWindow DurationShorthand `json:"lookbackWindow"`
	// AlertAfter sets how long the condition must remain valid before an alert is
	// triggered. OpenSLO treats an omitted value as "0m". This SDK leaves it unset.
	AlertAfter *DurationShorthand `json:"alertAfter,omitempty"`
}

// AlertConditionKind identifies the calculation used by an [AlertConditionType].
type AlertConditionKind string

const (
	// AlertConditionKindBurnRate compares an SLO's error-budget burn rate.
	AlertConditionKindBurnRate AlertConditionKind = "burnrate"
)

var alertConditionValidation = govy.New(
	validationRulesAPIVersion(func(a AlertCondition) openslo.Version { return a.APIVersion }),
	validationRulesKind(func(a AlertCondition) openslo.Kind { return a.Kind }, openslo.KindAlertCondition),
	validationRulesMetadata(func(a AlertCondition) Metadata { return a.Metadata }),
	govy.For(func(a AlertCondition) AlertConditionSpec { return a.Spec }).
		WithName("spec").
		Required().
		Include(alertConditionSpecValidation),
).WithNameFunc(internal.GetObjectName[AlertCondition])

var alertConditionSpecValidation = govy.New(
	govy.For(func(spec AlertConditionSpec) string { return spec.Description }).
		WithName("description").
		Rules(rules.StringMaxLength(1050)),
	govy.For(func(spec AlertConditionSpec) string { return spec.Severity }).
		WithName("severity").
		Required(),
	govy.For(func(spec AlertConditionSpec) AlertConditionType { return spec.Condition }).
		WithName("condition").
		Required().
		Include(
			alertConditionTypeValidation,
			alertConditionBurnRateValidation,
		),
)

var alertConditionTypeValidation = govy.New(
	govy.For(func(a AlertConditionType) AlertConditionKind { return a.Kind }).
		WithName("kind").
		Required().
		Rules(rules.OneOf(AlertConditionKindBurnRate)),
)

var alertConditionBurnRateValidation = govy.New(
	govy.For(func(a AlertConditionType) Operator { return a.Operator }).
		WithName("op").
		Required().
		Include(operatorValidation),
	govy.ForPointer(func(a AlertConditionType) *float64 { return a.Threshold }).
		WithName("threshold").
		Required(),
	govy.For(func(a AlertConditionType) DurationShorthand { return a.LookbackWindow }).
		WithName("lookbackWindow").
		Required().
		Include(durationShortHandValidation),
	govy.ForPointer(func(a AlertConditionType) *DurationShorthand { return a.AlertAfter }).
		WithName("alertAfter").
		Include(durationShortHandValidation),
).
	When(
		func(a AlertConditionType) bool { return a.Kind == AlertConditionKindBurnRate },
		govy.WhenDescription("'kind' is 'burnrate'"),
	)
