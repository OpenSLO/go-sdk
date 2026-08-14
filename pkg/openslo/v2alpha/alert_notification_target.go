package v2alpha

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/OpenSLO/go-sdk/internal"
	"github.com/OpenSLO/go-sdk/pkg/openslo"
)

var (
	_ = Object(AlertNotificationTarget{})
	_ = openslo.ObjectValidator[AlertNotificationTarget](AlertNotificationTarget{})
)

// NewAlertNotificationTarget returns a notification target from metadata and spec.
func NewAlertNotificationTarget(metadata Metadata, spec AlertNotificationTargetSpec) AlertNotificationTarget {
	return AlertNotificationTarget{
		APIVersion: APIVersion,
		Kind:       openslo.KindAlertNotificationTarget,
		Metadata:   metadata,
		Spec:       spec,
	}
}

// AlertNotificationTarget names a destination for alert delivery.
type AlertNotificationTarget struct {
	APIVersion openslo.Version             `json:"apiVersion"`
	Kind       openslo.Kind                `json:"kind"`
	Metadata   Metadata                    `json:"metadata"`
	Spec       AlertNotificationTargetSpec `json:"spec"`
}

// GetVersion returns [APIVersion].
func (a AlertNotificationTarget) GetVersion() openslo.Version {
	return APIVersion
}

// GetKind returns [openslo.KindAlertNotificationTarget].
func (a AlertNotificationTarget) GetKind() openslo.Kind {
	return openslo.KindAlertNotificationTarget
}

// GetName returns the notification target's metadata name.
func (a AlertNotificationTarget) GetName() string {
	return a.Metadata.Name
}

// Validate returns an error for an invalid notification target.
func (a AlertNotificationTarget) Validate() error {
	return alertNotificationTargetValidation.Validate(a)
}

// String returns the notification target's formatted version, kind, and name.
func (a AlertNotificationTarget) String() string {
	return internal.GetObjectName(a)
}

// GetMetadata returns the notification target's metadata.
func (a AlertNotificationTarget) GetMetadata() Metadata {
	return a.Metadata
}

// GetValidator returns the validator configured for [AlertNotificationTarget].
func (a AlertNotificationTarget) GetValidator() govy.Validator[AlertNotificationTarget] {
	return alertNotificationTargetValidation
}

// AlertNotificationTargetSpec identifies the consumer-defined notification
// destination.
type AlertNotificationTargetSpec struct {
	// Description summarizes the destination.
	Description string `json:"description,omitempty"`
	// Target identifies the consumer-defined destination for alert delivery.
	Target string `json:"target"`
}

var alertNotificationTargetValidation = govy.New(
	validationRulesAPIVersion(
		func(a AlertNotificationTarget) openslo.Version { return a.APIVersion },
	),
	validationRulesKind(
		func(a AlertNotificationTarget) openslo.Kind { return a.Kind },
		openslo.KindAlertNotificationTarget,
	),
	validationRulesMetadata(func(a AlertNotificationTarget) Metadata { return a.Metadata }),
	govy.For(func(a AlertNotificationTarget) AlertNotificationTargetSpec { return a.Spec }).
		WithName("spec").
		Required().
		Include(alertNotificationTargetSpecValidation),
).WithNameFunc(internal.GetObjectName[AlertNotificationTarget])

var alertNotificationTargetSpecValidation = govy.New(
	govy.For(func(spec AlertNotificationTargetSpec) string { return spec.Target }).
		WithName("target").
		Required(),
	govy.For(func(spec AlertNotificationTargetSpec) string { return spec.Description }).
		WithName("description").
		Rules(rules.StringMaxLength(1050)),
)
