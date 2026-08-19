package v1

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

// AlertPolicy defines which alert-condition states trigger an SLO alert. It also
// defines where the consuming system delivers the resulting notifications.
// It is distinct from an organization's policy for responding to error-budget
// consumption.
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

// GetName returns the name in the AlertPolicy's [Metadata].
func (a AlertPolicy) GetName() string {
	return a.Metadata.Name
}

// Validate returns an error for an invalid alert policy.
func (a AlertPolicy) Validate() error {
	return alertPolicyValidation.Validate(a)
}

// String returns the alert policy's formatted version and kind. It also returns
// [Metadata.Name] when set.
func (a AlertPolicy) String() string {
	return internal.GetObjectName(a)
}

// GetMetadata returns the AlertPolicy's [Metadata].
func (a AlertPolicy) GetMetadata() Metadata {
	return a.Metadata
}

// GetValidator returns the validator for AlertPolicy objects.
func (a AlertPolicy) GetValidator() govy.Validator[AlertPolicy] {
	return alertPolicyValidation
}

// AlertPolicySpec defines which condition states trigger an SLO alert and where
// the consuming system delivers the resulting notifications.
// The trigger fields have a false zero value. JSON encoding omits false values,
// and this SDK accepts all three fields as false.
type AlertPolicySpec struct {
	// Description summarizes the alert policy.
	Description string `json:"description,omitempty"`
	// AlertWhenNoData reports whether to trigger an alert when the associated
	// [SLO] has no burn-rate value. Its zero value is false.
	AlertWhenNoData bool `json:"alertWhenNoData,omitempty"`
	// AlertWhenBreaching reports whether to trigger an alert when the condition is
	// breaching. Its zero value is false.
	AlertWhenBreaching bool `json:"alertWhenBreaching,omitempty"`
	// AlertWhenResolved reports whether to trigger an alert when the condition
	// resolves. Its zero value is false.
	AlertWhenResolved bool `json:"alertWhenResolved,omitempty"`
	// Conditions contains exactly one alert condition, specified inline or by
	// reference.
	Conditions []AlertPolicyCondition `json:"conditions,omitempty"`
	// NotificationTargets contains one or more notification destinations. Specify
	// each destination inline or by reference.
	NotificationTargets []AlertPolicyNotificationTarget `json:"notificationTargets,omitempty"`
}

// AlertPolicyCondition supplies exactly one alert condition representation to an
// [AlertPolicySpec]. Set [AlertPolicyConditionInline] or
// [AlertPolicyConditionRef], but not both.
type AlertPolicyCondition struct {
	*AlertPolicyConditionRef
	*AlertPolicyConditionInline
}

// AlertPolicyConditionInline is the inline form of an [AlertCondition]. It omits
// [AlertCondition.APIVersion].
type AlertPolicyConditionInline struct {
	Kind     openslo.Kind       `json:"kind"`
	Metadata Metadata           `json:"metadata"`
	Spec     AlertConditionSpec `json:"spec"`
}

// AlertPolicyConditionRef identifies an existing [AlertCondition] by
// [Metadata.Name].
type AlertPolicyConditionRef struct {
	// ConditionRef matches the [Metadata.Name] of an existing [AlertCondition].
	ConditionRef string `json:"conditionRef"`
}

// AlertPolicyNotificationTarget supplies exactly one notification-target
// representation to an [AlertPolicySpec]. Set
// [AlertPolicyNotificationTargetInline] or [AlertPolicyNotificationTargetRef],
// but not both.
type AlertPolicyNotificationTarget struct {
	*AlertPolicyNotificationTargetRef
	*AlertPolicyNotificationTargetInline
}

// AlertPolicyNotificationTargetInline is the inline form of an
// [AlertNotificationTarget]. It omits [AlertNotificationTarget.APIVersion].
type AlertPolicyNotificationTargetInline struct {
	Kind     openslo.Kind                `json:"kind"`
	Metadata Metadata                    `json:"metadata"`
	Spec     AlertNotificationTargetSpec `json:"spec"`
}

// AlertPolicyNotificationTargetRef identifies an existing
// [AlertNotificationTarget] by [Metadata.Name].
type AlertPolicyNotificationTargetRef struct {
	// TargetRef matches the [Metadata.Name] of an existing
	// [AlertNotificationTarget].
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
		OmitEmpty().
		Rules(rules.StringMaxLength(1050)),
	govy.For(func(spec AlertPolicySpec) bool { return spec.AlertWhenNoData }).
		WithName("alertWhenNoData").
		OmitEmpty(),
	govy.For(func(spec AlertPolicySpec) bool { return spec.AlertWhenBreaching }).
		WithName("alertWhenBreaching").
		OmitEmpty(),
	govy.For(func(spec AlertPolicySpec) bool { return spec.AlertWhenResolved }).
		WithName("alertWhenResolved").
		OmitEmpty(),
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
