package openslosdk

import (
	"sync"

	"github.com/OpenSLO/go-sdk/pkg/openslo"
	v1 "github.com/OpenSLO/go-sdk/pkg/openslo/v1"
)

func NewReferenceExporter(objects ...openslo.Object) *ReferenceExporter {
	return &ReferenceExporter{
		config:   defaultReferenceConfig(),
		objects:  objects,
		exported: make([]openslo.Object, 0, len(objects)),
	}
}

type ReferenceExporter struct {
	config   ReferenceConfig
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

// WithConfig allows providing a custom [ReferenceConfig] which can help limit
// the exported definitions to a desired subset.
// Example:
//
//	 // only export [v1.SLI] reference for [v1.SLO]
//	 ReferenceConfig{
//			V1: ReferenceConfigV1{
//				SLO: &ReferenceConfigV1SLO{
//					SLI:         true,
//				},
//			},
//		}
func (r *ReferenceExporter) WithConfig(config ReferenceConfig) *ReferenceExporter {
	r.config = config
	return r
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
	if r.config.V1.AlertPolicy.AlertNotificationTarget {
		exported = append(exported, r.exportV1AlertPolicyTargets(&alertPolicy)...)
	}
	if r.config.V1.AlertPolicy.AlertCondition {
		exported = append(exported, r.exportV1AlertPolicyConditions(&alertPolicy)...)
	}
	return append([]openslo.Object{alertPolicy}, exported...)
}

func (r *ReferenceExporter) exportV1AlertPolicyTargets(alertPolicy *v1.AlertPolicy) []openslo.Object {
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
	return exported
}

func (r *ReferenceExporter) exportV1AlertPolicyConditions(alertPolicy *v1.AlertPolicy) []openslo.Object {
	exported := make([]openslo.Object, 0)
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
	return exported
}

func (r *ReferenceExporter) exportV1SLO(slo v1.SLO) []openslo.Object {
	exported := make([]openslo.Object, 0)
	if r.config.V1.SLO.AlertPolicy {
		exported = append(exported, r.exportV1SLOAlertPolicies(&slo)...)
	}
	if r.config.V1.SLO.SLI {
		exported = append(exported, r.exportV1SLOSLI(&slo)...)
	}
	return append([]openslo.Object{slo}, exported...)
}

func (r *ReferenceExporter) exportV1SLOAlertPolicies(slo *v1.SLO) []openslo.Object {
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
	return exported
}

func (r *ReferenceExporter) exportV1SLOSLI(slo *v1.SLO) []openslo.Object {
	if slo.Spec.Indicator == nil {
		return nil
	}
	sli := v1.NewSLI(slo.Spec.Indicator.Metadata, slo.Spec.Indicator.Spec)
	slo.Spec.IndicatorRef = &slo.Spec.Indicator.Metadata.Name
	slo.Spec.Indicator = nil
	return []openslo.Object{sli}
}

func (r *ReferenceExporter) addResult(objects ...openslo.Object) {
	r.exported = append(r.exported, objects...)
}
