// Package v2alpha contains Go representations and validators for the unstable
// [OpenSLO v2alpha proposal]. This package can change incompatibly as the
// proposal changes.
//
// Objects use the "openslo.com/v2alpha" API version and Kubernetes-style
// [Metadata]. The metadata has one value per label and no display name.
// Compared with v1, SLO indicator fields use the names "sli" and "sliRef".
// Metric source fields are "dataSourceRef", "dataSourceSpec", and "spec".
// Durations support minutes, hours, days, and weeks. Threshold-metric SLOs can
// have multiple objectives.
//
// The proposal also describes labels on individual SLO objectives, but
// [SLOObjective] does not currently expose an objective-label field.
// Where the proposal is incomplete, the exported fields, JSON tags, and
// validators define the SDK representation.
//
// [OpenSLO v2alpha proposal]: https://github.com/OpenSLO/OpenSLO/blob/e74b589cc98b98a5413611176d659a72318e7519/enhancements/v2alpha.md
package v2alpha
