package openslosdk

// ReferenceConfig configures [openslo.Object] references resolution.
// It is used both by [ReferenceInliner] and [ReferenceExporter].
// By default, all references are resolved.
type ReferenceConfig struct {
	V1 ReferenceConfigV1
}

// ReferenceConfigV1 configures [openslo.VersionV1] references resolution.
type ReferenceConfigV1 struct {
	SLO         ReferenceConfigV1SLO
	AlertPolicy ReferenceConfigV1AlertPolicy
}

// ReferenceConfigV1SLO configures [v1.SLO] references resolution.
type ReferenceConfigV1SLO struct {
	// AlertPolicy controls whether [openslo.KindAlertPolicy] references should be resolved.
	AlertPolicy bool
	// SLI controls whether [openslo.KindSLI] references should be resolved.
	SLI bool
}

// ReferenceConfigV1AlertPolicy configures [v1.AlertPolicy] references resolution.
type ReferenceConfigV1AlertPolicy struct {
	// AlertPolicy controls whether [openslo.KindAlertCondition] references should be resolved.
	AlertCondition bool
	// AlertPolicy controls whether [openslo.KindAlertNotificationTarget] references should be resolved.
	AlertNotificationTarget bool
}

func defaultReferenceConfig() ReferenceConfig {
	return ReferenceConfig{
		V1: ReferenceConfigV1{
			SLO: ReferenceConfigV1SLO{
				AlertPolicy: true,
				SLI:         true,
			},
			AlertPolicy: ReferenceConfigV1AlertPolicy{
				AlertCondition:          true,
				AlertNotificationTarget: true,
			},
		},
	}
}
