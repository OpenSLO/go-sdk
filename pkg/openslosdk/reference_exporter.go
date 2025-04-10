package openslosdk

import (
	"sync"

	"github.com/OpenSLO/go-sdk/pkg/openslo"
	v1 "github.com/OpenSLO/go-sdk/pkg/openslo/v1"
)

func NewReferenceExporter(objects ...openslo.Object) *ReferenceExporter {
	return &ReferenceExporter{
		objects:  objects,
		exported: make([]openslo.Object, 0, len(objects)),
	}
}

type ReferenceExporter struct {
	objects  []openslo.Object
	exported []openslo.Object
	once     sync.Once
}

// Export replaces all the inlined objects with references and returns
// the original objects along with the exported, previously inlined, objects.
func (r *ReferenceExporter) Export() []openslo.Object {
	r.once.Do(func() {
		r.exported = r.exportObjects()
	})
	return r.exported
}

func (r *ReferenceExporter) exportObjects() []openslo.Object {
	for _, object := range r.objects {
		r.exportObject(object)
	}
	return r.exported
}

func (r *ReferenceExporter) exportObject(object openslo.Object) {
	version := object.GetVersion()
	switch version {
	case openslo.VersionV1:
		r.addResult(r.exportV1Object(object)...)
	default:
		r.addResult(object)
	}
}

func (r *ReferenceExporter) exportV1Object(object openslo.Object) []openslo.Object {
	switch v := object.(type) {
	case v1.AlertPolicy:
		return r.exportV1AlertPolicy(v)
	case v1.SLO:
		return r.exportV1SLO(v)
	default:
		return []openslo.Object{object}
	}
}

func (r *ReferenceExporter) exportV1AlertPolicy(alertPolicy v1.AlertPolicy) []openslo.Object {
	exported := make([]openslo.Object, 0)
	for i, target := range alertPolicy.Spec.NotificationTargets {
		if target.AlertPolicyNotificationTargetInline == nil {
			continue
		}
		exported = append(exported, v1.NewAlertNotificationTarget(target.Metadata, target.Spec))
		target.AlertPolicyNotificationTargetRef = &v1.AlertPolicyNotificationTargetRef{
			TargetRef: target.Metadata.Name,
		}
		target.AlertPolicyNotificationTargetInline = nil
		alertPolicy.Spec.NotificationTargets[i] = target
	}
	for i, condition := range alertPolicy.Spec.Conditions {
		if condition.AlertPolicyConditionInline == nil {
			continue
		}
		exported = append(exported, v1.NewAlertCondition(condition.Metadata, condition.Spec))
		condition.AlertPolicyConditionRef = &v1.AlertPolicyConditionRef{
			ConditionRef: condition.Metadata.Name,
		}
		condition.AlertPolicyConditionInline = nil
		alertPolicy.Spec.Conditions[i] = condition
	}
	return append([]openslo.Object{alertPolicy}, exported...)
}

func (r *ReferenceExporter) exportV1SLO(slo v1.SLO) []openslo.Object {
	exported := make([]openslo.Object, 0)
	for i, ap := range slo.Spec.AlertPolicies {
		if ap.SLOAlertPolicyInline == nil {
			continue
		}
		alertPolicy := v1.NewAlertPolicy(ap.Metadata, ap.Spec)
		exported = append(exported, r.exportV1AlertPolicy(alertPolicy)...)
		ap.SLOAlertPolicyRef = &v1.SLOAlertPolicyRef{
			AlertPolicyRef: alertPolicy.Metadata.Name,
		}
		ap.SLOAlertPolicyInline = nil
		slo.Spec.AlertPolicies[i] = ap
	}
	if slo.Spec.Indicator != nil {
		exported = append(exported, v1.NewSLI(slo.Spec.Indicator.Metadata, slo.Spec.Indicator.Spec))
		slo.Spec.IndicatorRef = &slo.Spec.Indicator.Metadata.Name
		slo.Spec.Indicator = nil
	}
	return append([]openslo.Object{slo}, exported...)
}

func (r *ReferenceExporter) addResult(objects ...openslo.Object) {
	r.exported = append(r.exported, objects...)
}
