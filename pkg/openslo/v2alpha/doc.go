// Package v2alpha contains Go representations and validators for the unstable
// v2alpha API. This package can change incompatibly.
//
// Objects use the "openslo.com/v2alpha" API version and Kubernetes-style
// [Metadata]. The metadata has one value per label and no display name.
// SLO indicator fields use the names "sli" and "sliRef".
// Metric source fields are "dataSourceRef", "dataSourceSpec", and "spec".
// Durations support minutes, hours, days, and weeks. Threshold-metric SLOs can
// have multiple objectives.
//
// The proposal also describes labels on individual SLO objectives, but
// [SLOObjective] does not expose an objective-label field. Exported fields,
// JSON tags, and validators define the SDK representation.
package v2alpha
