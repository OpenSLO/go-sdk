package openslosdk

import (
	"fmt"

	"github.com/OpenSLO/go-sdk/pkg/openslo"
	v1 "github.com/OpenSLO/go-sdk/pkg/openslo/v1"
)

func InlineObjects(objects ...openslo.Object) ([]openslo.Object, error) {
	inliner := &objectsInliner{
		objects:                 objects,
		inlined:                 make([]openslo.Object, 0, len(objects)),
		referencedObjectIndexes: make(map[int]bool),
	}
	for _, object := range objects {
		if err := inliner.inlineObject(object); err != nil {
			return nil, err
		}
	}
	return inliner.removeReferencedObjects(), nil
}

type objectsInliner struct {
	objects                 []openslo.Object
	inlined                 []openslo.Object
	referencedObjectIndexes map[int]bool
}

func (o *objectsInliner) inlineObject(object openslo.Object) error {
	version := object.GetVersion()
	switch version {
	case openslo.VersionV1:
		inlinedObject, err := o.inlineV1Object(object)
		if err != nil {
			return fmt.Errorf("failed to inline %s %s %s: %w",
				object.GetVersion(), object.GetKind(), object.GetName(), err)
		}
		o.addResult(inlinedObject)
	default:
		o.addResult(object)
	}
	return nil
}

func (o *objectsInliner) removeReferencedObjects() []openslo.Object {
	objects := make([]openslo.Object, 0, len(o.inlined))
	for i := range o.inlined {
		if o.referencedObjectIndexes[i] {
			continue
		}
		objects = append(objects, o.inlined[i])
	}
	return objects
}

func (o *objectsInliner) inlineV1Object(object openslo.Object) (openslo.Object, error) {
	switch v := object.(type) {
	case v1.AlertPolicy:
		return o.inlineV1AlertPolicy(v)
	default:
		return object, nil
	}
}

func (o *objectsInliner) inlineV1AlertPolicy(alertPolicy v1.AlertPolicy) (openslo.Object, error) {
	for i, target := range alertPolicy.Spec.NotificationTargets {
		if target.AlertPolicyNotificationTargetRef == nil {
			continue
		}
		alertNotificationTarget, idx := findObject[v1.AlertNotificationTarget](o.objects, target.TargetRef)
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
		o.referencedObjectIndexes[idx] = true
		alertPolicy.Spec.NotificationTargets[i] = target
	}
	for i, condition := range alertPolicy.Spec.Conditions {
		if condition.AlertPolicyConditionRef == nil {
			continue
		}
		alertCondition, idx := findObject[v1.AlertCondition](o.objects, condition.ConditionRef)
		if idx == -1 {
			return nil, newReferenceNotFoundErr(
				alertCondition,
				fmt.Sprintf("spec.conditions[%d].targetRef", i),
				condition.ConditionRef,
			)
		}
		condition.AlertPolicyConditionRef = nil
		condition.AlertPolicyConditionInline = &v1.AlertPolicyConditionInline{
			Kind:     alertCondition.GetKind(),
			Metadata: alertCondition.Metadata,
			Spec:     alertCondition.Spec,
		}
		o.referencedObjectIndexes[idx] = true
		alertPolicy.Spec.Conditions[i] = condition
	}
	return alertPolicy, nil
}

func (o *objectsInliner) addResult(object openslo.Object) {
	o.inlined = append(o.inlined, object)
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
	return fmt.Errorf("%s %s %s object referenced at '%s' does not exist",
		object.GetVersion(), object.GetKind(), name, path)
}
