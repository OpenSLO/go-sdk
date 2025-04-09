package openslosdk

import (
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

type ReferenceResolver struct {
	objects                 []openslo.Object
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
	default:
		return object, nil
	}
}

func (r *ReferenceResolver) inlineV1AlertPolicy(alertPolicy v1.AlertPolicy) (openslo.Object, error) {
	for i, target := range alertPolicy.Spec.NotificationTargets {
		if target.AlertPolicyNotificationTargetRef == nil {
			continue
		}
		alertNotificationTarget, idx := findObject[v1.AlertNotificationTarget](r.objects, target.TargetRef)
		if idx == -1 {
			return nil, newReferenceNotFoundErr(
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
		alertCondition, idx := findObject[v1.AlertCondition](r.objects, condition.ConditionRef)
		if idx == -1 {
			return nil, newReferenceNotFoundErr(
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
	return fmt.Errorf("%s '%s' referenced at '%s' does not exist", object, name, path)
}
