package v2alpha

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/OpenSLO/go-sdk/internal"
	"github.com/OpenSLO/go-sdk/pkg/openslo"
)

var (
	_ = Object(AlertPolicy{})
	_ = openslo.ObjectValidator[AlertPolicy](AlertPolicy{})
)

// NewAlertPolicy returns an AlertPolicy from metadata and spec.
func NewAlertPolicy(metadata Metadata, spec AlertPolicySpec) AlertPolicy {
	return AlertPolicy{
		APIVersion: APIVersion,
		Kind:       openslo.KindAlertPolicy,
		Metadata:   metadata,
		Spec:       spec,
	}
}

// AlertPolicy defines when the system emits an SLO alert and where it sends it.
type AlertPolicy struct {
	APIVersion openslo.Version `json:"apiVersion"`
	Kind       openslo.Kind    `json:"kind"`
	Metadata   Metadata        `json:"metadata"`
	Spec       AlertPolicySpec `json:"spec"`
}

// GetVersion returns [APIVersion].
func (a AlertPolicy) GetVersion() openslo.Version {
	return APIVersion
}

// GetKind returns [openslo.KindAlertPolicy].
func (a AlertPolicy) GetKind() openslo.Kind {
	return openslo.KindAlertPolicy
}

// GetName returns the alert policy's metadata name.
func (a AlertPolicy) GetName() string {
	return a.Metadata.Name
}

// Validate returns an error for an invalid alert policy.
func (a AlertPolicy) Validate() error {
	return alertPolicyValidation.Validate(a)
}

// String returns the alert policy's formatted version, kind, and name.
func (a AlertPolicy) String() string {
	return internal.GetObjectName(a)
}

// GetMetadata returns the alert policy's metadata.
func (a AlertPolicy) GetMetadata() Metadata {
	return a.Metadata
}

// GetValidator returns the validator configured for [AlertPolicy].
func (a AlertPolicy) GetValidator() govy.Validator[AlertPolicy] {
	return alertPolicyValidation
}

// AlertPolicySpec defines the trigger states, condition, and notification
// destinations for an [AlertPolicy].
type AlertPolicySpec struct {
	// Description summarizes the alert policy.
	Description string `json:"description,omitempty"`
	// AlertWhenNoData enables notifications when the associated SLO has no
	// burn-rate value.
	AlertWhenNoData bool `json:"alertWhenNoData,omitempty"`
	// AlertWhenBreaching enables notifications when the condition starts breaching.
	AlertWhenBreaching bool `json:"alertWhenBreaching,omitempty"`
	// AlertWhenResolved enables notifications when the condition resolves.
	AlertWhenResolved bool `json:"alertWhenResolved,omitempty"`
	// Conditions supplies the policy's alert condition by reference or inline.
	Conditions []AlertPolicyCondition `json:"conditions,omitempty"`
	// NotificationTargets lists referenced or inline delivery destinations.
	NotificationTargets []AlertPolicyNotificationTarget `json:"notificationTargets,omitempty"`
}

// AlertPolicyCondition supplies a condition to an [AlertPolicy] by reference
// or inline definition.
type AlertPolicyCondition struct {
	*AlertPolicyConditionRef
	*AlertPolicyConditionInline
}

// AlertPolicyConditionInline is an alert-condition definition embedded in an
// [AlertPolicy].
type AlertPolicyConditionInline struct {
	Kind     openslo.Kind       `json:"kind"`
	Metadata Metadata           `json:"metadata"`
	Spec     AlertConditionSpec `json:"spec"`
}

// AlertPolicyConditionRef identifies a separately defined [AlertCondition].
type AlertPolicyConditionRef struct {
	// ConditionRef names the alert condition to use.
	ConditionRef string `json:"conditionRef"`
}

// AlertPolicyNotificationTarget supplies a notification destination to an
// [AlertPolicy] by reference or inline definition.
type AlertPolicyNotificationTarget struct {
	*AlertPolicyNotificationTargetRef
	*AlertPolicyNotificationTargetInline
}

// AlertPolicyNotificationTargetInline is an alert-notification-target
// definition embedded in an [AlertPolicy].
type AlertPolicyNotificationTargetInline struct {
	Kind     openslo.Kind                `json:"kind"`
	Metadata Metadata                    `json:"metadata"`
	Spec     AlertNotificationTargetSpec `json:"spec"`
}

// AlertPolicyNotificationTargetRef identifies a separately defined
// [AlertNotificationTarget].
type AlertPolicyNotificationTargetRef struct {
	// TargetRef names the notification target to use.
	TargetRef string `json:"targetRef"`
}

var alertPolicyValidation = govy.New(
	validationRulesAPIVersion(func(a AlertPolicy) openslo.Version { return a.APIVersion }),
	validationRulesKind(func(a AlertPolicy) openslo.Kind { return a.Kind }, openslo.KindAlertPolicy),
	validationRulesMetadata(func(a AlertPolicy) Metadata { return a.Metadata }),
	govy.For(func(a AlertPolicy) AlertPolicySpec { return a.Spec }).
		WithName("spec").
		Include(alertPolicySpecValidation),
).WithNameFunc(internal.GetObjectName[AlertPolicy])

var alertPolicySpecValidation = govy.New(
	govy.For(func(spec AlertPolicySpec) string { return spec.Description }).
		WithName("description").
		Rules(rules.StringMaxLength(1050)),
	govy.ForSlice(func(spec AlertPolicySpec) []AlertPolicyCondition { return spec.Conditions }).
		WithName("conditions").
		Rules(rules.SliceLength[[]AlertPolicyCondition](1, 1)).
		IncludeForEach(alertPolicyConditionValidation),
	govy.ForSlice(func(spec AlertPolicySpec) []AlertPolicyNotificationTarget { return spec.NotificationTargets }).
		WithName("notificationTargets").
		Rules(rules.SliceMinLength[[]AlertPolicyNotificationTarget](1)).
		IncludeForEach(alertPolicyNotificationTargetValidation),
)

var alertPolicyConditionValidation = govy.New(
	govy.For(govy.GetSelf[AlertPolicyCondition]()).
		Rules(rules.MutuallyExclusive(true, map[string]func(a AlertPolicyCondition) any{
			"conditionRef": func(a AlertPolicyCondition) any { return a.AlertPolicyConditionRef },
			// It's impossible to list all fields that constitute the inlined version in the error message,
			// therefore 'spec' must suffice.
			"spec": func(a AlertPolicyCondition) any { return a.AlertPolicyConditionInline },
		}).
			WithDescription("exactly one of 'conditionRef' and 'spec' must be set")),
	govy.ForPointer(func(a AlertPolicyCondition) *AlertPolicyConditionRef { return a.AlertPolicyConditionRef }).
		Include(govy.New(
			govy.For(func(ref AlertPolicyConditionRef) string { return ref.ConditionRef }).
				WithName("conditionRef").
				Required().
				Rules(rules.StringDNSLabel()),
		)).Cascade(govy.CascadeModeContinue),
	govy.ForPointer(func(a AlertPolicyCondition) *AlertPolicyConditionInline { return a.AlertPolicyConditionInline }).
		Include(govy.New(
			govy.For(func(inline AlertPolicyConditionInline) openslo.Kind { return inline.Kind }).
				WithName("kind").
				Required().
				Rules(rules.EQ(openslo.KindAlertCondition)),
			validationRulesMetadata(func(a AlertPolicyConditionInline) Metadata { return a.Metadata }),
			govy.For(func(inline AlertPolicyConditionInline) AlertConditionSpec { return inline.Spec }).
				WithName("spec").
				Required().
				Include(alertConditionSpecValidation),
		)).Cascade(govy.CascadeModeContinue),
).Cascade(govy.CascadeModeStop)

var alertPolicyNotificationTargetValidation = govy.New(
	govy.For(govy.GetSelf[AlertPolicyNotificationTarget]()).
		Rules(rules.MutuallyExclusive(true, map[string]func(a AlertPolicyNotificationTarget) any{
			"targetRef": func(a AlertPolicyNotificationTarget) any { return a.AlertPolicyNotificationTargetRef },
			// It's impossible to list all fields that constitute the inlined version in the error message,
			// therefore 'spec' must suffice.
			"spec": func(a AlertPolicyNotificationTarget) any { return a.AlertPolicyNotificationTargetInline },
		}).
			WithDescription("exactly one of 'targetRef' and 'spec' must be set")),
	govy.ForPointer(func(a AlertPolicyNotificationTarget) *AlertPolicyNotificationTargetRef {
		return a.AlertPolicyNotificationTargetRef
	}).
		Include(govy.New(
			govy.For(func(ref AlertPolicyNotificationTargetRef) string { return ref.TargetRef }).
				WithName("targetRef").
				Required().
				Rules(rules.StringDNSLabel()),
		)).Cascade(govy.CascadeModeContinue),
	govy.ForPointer(func(a AlertPolicyNotificationTarget) *AlertPolicyNotificationTargetInline {
		return a.AlertPolicyNotificationTargetInline
	}).
		Include(govy.New(
			govy.For(func(inline AlertPolicyNotificationTargetInline) openslo.Kind { return inline.Kind }).
				WithName("kind").
				Required().
				Rules(rules.EQ(openslo.KindAlertNotificationTarget)),
			validationRulesMetadata(func(a AlertPolicyNotificationTargetInline) Metadata { return a.Metadata }),
			govy.For(func(inline AlertPolicyNotificationTargetInline) AlertNotificationTargetSpec { return inline.Spec }).
				WithName("spec").
				Required().
				Include(alertNotificationTargetSpecValidation),
		)).Cascade(govy.CascadeModeContinue),
).Cascade(govy.CascadeModeStop)
