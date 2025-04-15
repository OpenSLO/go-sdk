package openslosdk

import (
	"errors"
	"fmt"
	"sync"

	"github.com/OpenSLO/go-sdk/pkg/openslo"
	v1 "github.com/OpenSLO/go-sdk/pkg/openslo/v1"
)

func NewReferenceInliner(objects ...openslo.Object) *ReferenceInliner {
	return &ReferenceInliner{
		config:                  defaultReferenceConfig(),
		objects:                 objects,
		inlined:                 make([]openslo.Object, 0, len(objects)),
		referencedObjectIndexes: make(map[int]bool),
	}
}

// ReferenceInliner is a utility to inline referenced [openslo.Object] in referencing object(s).
type ReferenceInliner struct {
	config                  ReferenceConfig
	objects                 []openslo.Object
	references              []openslo.Object
	inlined                 []openslo.Object
	referencedObjectIndexes map[int]bool
	removeRefs              bool
	err                     error
	once                    sync.Once
}

// Inline finds all referenced objects in the provided slice of [openslo.Object]
// and replaces the references with an inlined version of the referenced [openslo.Object].
// If the referenced object is not found in the provided [openslo.Object] slice, an error will be returned.
//
// By default, it will not remove referenced objects from the result.
// If you want to remove referenced objects, you can use the [ReferenceInliner.RemoveReferencedObjects] option.
func (r *ReferenceInliner) Inline() ([]openslo.Object, error) {
	r.once.Do(func() {
		r.inlined, r.err = r.inlineObjects()
	})
	return r.inlined, r.err
}

// RemoveReferencedObjects instructs [ReferenceInliner] to remove referenced objects
// from the result of [ReferenceInliner.Inline].
func (r *ReferenceInliner) RemoveReferencedObjects() *ReferenceInliner {
	r.removeRefs = true
	return r
}

// WithConfig allows providing a custom [ReferenceConfig] which can help limit
// the inlined references to a desired subset.
// Example:
//
//	 // only inline [v1.SLI] reference for [v1.SLO]
//	 ReferenceConfig{
//			V1: ReferenceConfigV1{
//				SLO: &ReferenceConfigV1SLO{
//					SLI:         true,
//				},
//			},
//		}
func (r *ReferenceInliner) WithConfig(config ReferenceConfig) *ReferenceInliner {
	r.config = config
	return r
}

func (r *ReferenceInliner) inlineObjects() ([]openslo.Object, error) {
	if r.references == nil {
		r.references = r.objects
	}
	for _, object := range r.objects {
		if err := r.inlineObject(object); err != nil {
			return nil, err
		}
	}
	if r.removeRefs {
		return r.removeReferencedObjects(), nil
	}
	return r.inlined, nil
}

func (r *ReferenceInliner) inlineObject(object openslo.Object) error {
	version := object.GetVersion()
	switch version {
	case openslo.VersionV1:
		inlinedObject, err := r.inlineV1Object(object)
		if err != nil {
			return fmt.Errorf("failed to inline %s: %w", object, err)
		}
		r.addResult(inlinedObject)
	default:
		r.addResult(object)
	}
	return nil
}

func (r *ReferenceInliner) removeReferencedObjects() []openslo.Object {
	objects := make([]openslo.Object, 0, len(r.inlined))
	for i := range r.inlined {
		if r.referencedObjectIndexes[i] {
			continue
		}
		objects = append(objects, r.inlined[i])
	}
	return objects
}

func (r *ReferenceInliner) inlineV1Object(object openslo.Object) (openslo.Object, error) {
	switch v := object.(type) {
	case v1.AlertPolicy:
		return r.inlineV1AlertPolicy(v)
	case v1.SLO:
		return r.inlineV1SLO(v)
	default:
		return object, nil
	}
}

func (r *ReferenceInliner) inlineV1AlertPolicy(alertPolicy v1.AlertPolicy) (v1.AlertPolicy, error) {
	var err error
	if r.config.V1.AlertPolicy.AlertNotificationTarget {
		alertPolicy, err = r.inlineV1AlertPolicyTargets(alertPolicy)
		if err != nil {
			return v1.AlertPolicy{}, err
		}
	}
	if r.config.V1.AlertPolicy.AlertCondition {
		alertPolicy, err = r.inlineV1AlertPolicyConditions(alertPolicy)
		if err != nil {
			return v1.AlertPolicy{}, err
		}
	}
	return alertPolicy, nil
}

func (r *ReferenceInliner) inlineV1AlertPolicyTargets(alertPolicy v1.AlertPolicy) (v1.AlertPolicy, error) {
	for i, target := range alertPolicy.Spec.NotificationTargets {
		if target.AlertPolicyNotificationTargetRef == nil {
			continue
		}
		alertNotificationTarget, idx := findObject[v1.AlertNotificationTarget](r.references, target.TargetRef)
		if idx == -1 {
			return v1.AlertPolicy{}, newReferenceNotFoundErr(
				alertNotificationTarget,
				fmt.Sprintf("spec.notificationTargets[%d].targetRef", i),
				target.TargetRef,
			)
		}
		target.AlertPolicyNotificationTargetRef = nil
		target.AlertPolicyNotificationTargetInline = &v1.AlertPolicyNotificationTargetInline{
			Kind:     alertNotificationTarget.GetKind(),
			Metadata: alertNotificationTarget.Metadata,
			Spec:     alertNotificationTarget.Spec,
		}
		r.referencedObjectIndexes[idx] = true
		alertPolicy.Spec.NotificationTargets[i] = target
	}
	return alertPolicy, nil
}

func (r *ReferenceInliner) inlineV1AlertPolicyConditions(alertPolicy v1.AlertPolicy) (v1.AlertPolicy, error) {
	for i, condition := range alertPolicy.Spec.Conditions {
		if condition.AlertPolicyConditionRef == nil {
			continue
		}
		alertCondition, idx := findObject[v1.AlertCondition](r.references, condition.ConditionRef)
		if idx == -1 {
			return v1.AlertPolicy{}, newReferenceNotFoundErr(
				alertCondition,
				fmt.Sprintf("spec.conditions[%d].conditionRef", i),
				condition.ConditionRef,
			)
		}
		condition.AlertPolicyConditionRef = nil
		condition.AlertPolicyConditionInline = &v1.AlertPolicyConditionInline{
			Kind:     alertCondition.GetKind(),
			Metadata: alertCondition.Metadata,
			Spec:     alertCondition.Spec,
		}
		r.referencedObjectIndexes[idx] = true
		alertPolicy.Spec.Conditions[i] = condition
	}
	return alertPolicy, nil
}

func (r *ReferenceInliner) inlineV1SLO(slo v1.SLO) (v1.SLO, error) {
	var err error
	if r.config.V1.SLO.AlertPolicy {
		slo, err = r.inlineV1SLOAlertPolicies(slo)
		if err != nil {
			return v1.SLO{}, err
		}
	}
	if r.config.V1.SLO.SLI {
		slo, err = r.inlineV1SLOSLI(slo)
		if err != nil {
			return v1.SLO{}, err
		}
	}
	return slo, nil
}

func (r *ReferenceInliner) inlineV1SLOAlertPolicies(slo v1.SLO) (v1.SLO, error) {
	for i, ap := range slo.Spec.AlertPolicies {
		var alertPolicy v1.AlertPolicy
		switch {
		case ap.SLOAlertPolicyInline != nil:
			alertPolicy = v1.NewAlertPolicy(ap.Metadata, ap.Spec)
		default:
			var idx int
			alertPolicy, idx = findObject[v1.AlertPolicy](r.references, ap.AlertPolicyRef)
			if idx == -1 {
				return v1.SLO{}, newReferenceNotFoundErr(
					alertPolicy,
					fmt.Sprintf("spec.alertPolicies[%d].alertPolicyRef", i),
					ap.AlertPolicyRef,
				)
			}
			r.referencedObjectIndexes[idx] = true
		}

		inlinedAlertPolicy, err := r.inlineV1AlertPolicy(alertPolicy)
		if err != nil {
			var refErr referenceNotFoundErr
			if errors.As(err, &refErr) {
				refErr.fieldPath = fmt.Sprintf("spec.alertPolicies[%d].%s", i, refErr.fieldPath)
				return v1.SLO{}, refErr
			}
			return v1.SLO{}, fmt.Errorf(
				"failed to inline %s referenced at 'spec.alertPolicies[%d].alertPolicyRef': %w",
				alertPolicy, i, err)
		}
		ap.SLOAlertPolicyRef = nil
		ap.SLOAlertPolicyInline = &v1.SLOAlertPolicyInline{
			Kind:     inlinedAlertPolicy.GetKind(),
			Metadata: inlinedAlertPolicy.Metadata,
			Spec:     inlinedAlertPolicy.Spec,
		}
		slo.Spec.AlertPolicies[i] = ap
	}
	return slo, nil
}

func (r *ReferenceInliner) inlineV1SLOSLI(slo v1.SLO) (v1.SLO, error) {
	if slo.Spec.IndicatorRef == nil {
		return slo, nil
	}
	sli, idx := findObject[v1.SLI](r.references, *slo.Spec.IndicatorRef)
	if idx == -1 {
		return v1.SLO{}, newReferenceNotFoundErr(
			sli,
			"spec.indicatorRef",
			*slo.Spec.IndicatorRef,
		)
	}
	slo.Spec.IndicatorRef = nil
	slo.Spec.Indicator = &v1.SLOIndicatorInline{
		Metadata: sli.Metadata,
		Spec:     sli.Spec,
	}
	r.referencedObjectIndexes[idx] = true
	return slo, nil
}

func (r *ReferenceInliner) addResult(object openslo.Object) {
	r.inlined = append(r.inlined, object)
}

func findObject[T openslo.Object](objects []openslo.Object, name string) (object T, objectIndex int) {
	for i := range objects {
		if objects[i].GetName() != name {
			continue
		}
		if v, ok := objects[i].(T); ok {
			return v, i
		}
	}
	return object, -1
}

func newReferenceNotFoundErr(object openslo.Object, path, name string) error {
	return referenceNotFoundErr{
		objectName: name,
		fieldPath:  path,
		object:     object,
	}
}

type referenceNotFoundErr struct {
	objectName string
	fieldPath  string
	object     openslo.Object
}

func (r referenceNotFoundErr) Error() string {
	return fmt.Sprintf("%s '%s' referenced at '%s' does not exist", r.object, r.objectName, r.fieldPath)
}
