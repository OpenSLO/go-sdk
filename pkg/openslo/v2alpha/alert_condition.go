package v2alpha

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

// AlertCondition defines when an SLO alert condition is breaching.
// [AlertPolicySpec.AlertWhenBreaching] controls whether that state triggers an
// alert.
//
// This type is the SDK's v2alpha representation. The living, unstable
// [OpenSLO v2alpha proposal] does not define alert objects.
//
// [OpenSLO v2alpha proposal]: https://github.com/OpenSLO/OpenSLO/blob/e74b589cc98b98a5413611176d659a72318e7519/enhancements/v2alpha.md
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

// GetName returns the alert condition's metadata name.
func (a AlertCondition) GetName() string {
	return a.Metadata.Name
}

// Validate returns an error for an invalid alert condition.
func (a AlertCondition) Validate() error {
	return alertConditionValidation.Validate(a)
}

// String returns the alert condition's formatted version and kind.
// It also returns the metadata name when set.
func (a AlertCondition) String() string {
	return internal.GetObjectName(a)
}

// GetMetadata returns the alert condition's metadata.
func (a AlertCondition) GetMetadata() Metadata {
	return a.Metadata
}

// GetValidator returns the validator configured for [AlertCondition].
func (a AlertCondition) GetValidator() govy.Validator[AlertCondition] {
	return alertConditionValidation
}

// AlertConditionSpec defines an alert's severity and breach condition.
type AlertConditionSpec struct {
	// Severity is a consumer-defined alert classification.
	Severity  string             `json:"severity"`
	Condition AlertConditionType `json:"condition"`
	// Description optionally summarizes the alert condition in at most 1,050 characters.
	Description string `json:"description,omitempty"`
}

// AlertConditionType defines a burn-rate comparison over a lookback window.
// Burn rate is error-budget consumption relative to the rate allowed by the
// SLO.
type AlertConditionType struct {
	// Kind selects the condition algorithm. This SDK requires Kind and does not
	// apply the OpenSLO v1 default of [AlertConditionKindBurnRate].
	Kind AlertConditionKind `json:"kind"`
	// Operator compares the calculated burn rate with Threshold.
	Operator Operator `json:"op"`
	// Threshold sets the numeric burn-rate boundary.
	Threshold *float64 `json:"threshold"`
	// LookbackWindow sets the period for burn-rate calculation.
	LookbackWindow DurationShorthand `json:"lookbackWindow"`
	// AlertAfter sets how long the burn-rate comparison must remain true before
	// the condition becomes breaching. This SDK requires AlertAfter and does not
	// apply the OpenSLO v1 default of "0m".
	AlertAfter DurationShorthand `json:"alertAfter"`
}

// AlertConditionKind identifies the evaluation algorithm for an
// [AlertConditionType].
type AlertConditionKind string

const (
	// AlertConditionKindBurnRate selects an error-budget burn-rate comparison.
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
		OmitEmpty().
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
	govy.For(func(a AlertConditionType) DurationShorthand { return a.AlertAfter }).
		WithName("alertAfter").
		Required().
		Include(durationShortHandValidation),
).
	When(
		func(a AlertConditionType) bool { return a.Kind == AlertConditionKindBurnRate },
		govy.WhenDescription("'kind' is 'burnrate'"),
	)
