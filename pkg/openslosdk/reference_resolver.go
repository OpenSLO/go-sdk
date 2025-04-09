package openslosdk

import (
	"errors"
	"fmt"
	"sync"

	"github.com/OpenSLO/go-sdk/pkg/openslo"
	v1 "github.com/OpenSLO/go-sdk/pkg/openslo/v1"
)

func NewReferenceResolver(objects ...openslo.Object) *ReferenceResolver {
	return &ReferenceResolver{
		objects:                 objects,
		inlined:                 make([]openslo.Object, 0, len(objects)),
		referencedObjectIndexes: make(map[int]bool),
	}
}

// ReferenceResolver is a utility to inline referenced [openslo.Object] in referencing object(s).
type ReferenceResolver struct {
	objects                 []openslo.Object
	references              []openslo.Object
	inlined                 []openslo.Object
	referencedObjectIndexes map[int]bool
	removeRefs              bool
	err                     error
	inlineOnce              sync.Once
}

// RemoveReferencedObjects instructs [ReferenceResolver] to remove referenced objects
// from the result of [ReferenceResolver.Inline].
func (r *ReferenceResolver) RemoveReferencedObjects() *ReferenceResolver {
	r.removeRefs = true
	return r
}

// Inline finds all referenced objects in the provided slice of [openslo.Object]
// and replaces the references with an inlined version of the referenced [openslo.Object].
// If the referenced object is not found in the provided [openslo.Object] slice, an error will be returned.
//
// By default, it will not remove referenced objects from the result.
// If you want to remove referenced objects, you can use the [ReferenceResolver.RemoveReferencedObjects] option.
func (r *ReferenceResolver) Inline() ([]openslo.Object, error) {
	r.inlineOnce.Do(func() {
		r.inlined, r.err = r.resolve()
	})
	return r.inlined, r.err
}

func (r *ReferenceResolver) resolve() ([]openslo.Object, error) {
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

func (r *ReferenceResolver) inlineObject(object openslo.Object) error {
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

func (r *ReferenceResolver) removeReferencedObjects() []openslo.Object {
	objects := make([]openslo.Object, 0, len(r.inlined))
	for i := range r.inlined {
		if r.referencedObjectIndexes[i] {
			continue
		}
		objects = append(objects, r.inlined[i])
	}
	return objects
}

func (r *ReferenceResolver) inlineV1Object(object openslo.Object) (openslo.Object, error) {
	switch v := object.(type) {
	case v1.AlertPolicy:
		return r.inlineV1AlertPolicy(v)
	case v1.SLO:
		return r.inlineV1SLO(v)
	default:
		return object, nil
	}
}

func (r *ReferenceResolver) inlineV1AlertPolicy(alertPolicy v1.AlertPolicy) (v1.AlertPolicy, error) {
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

func (r *ReferenceResolver) inlineV1SLO(slo v1.SLO) (v1.SLO, error) {
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
	if slo.Spec.IndicatorRef != nil {
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
	}
	return slo, nil
}

func (r *ReferenceResolver) addResult(object openslo.Object) {
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
