package v2alpha

import (
	"regexp"
	"slices"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/OpenSLO/go-sdk/pkg/openslo"
)

// APIVersion is the OpenSLO v2alpha API version.
const APIVersion = openslo.VersionV2alpha

var supportedKinds = []openslo.Kind{
	openslo.KindSLO,
	openslo.KindSLI,
	openslo.KindDataSource,
	openslo.KindService,
	openslo.KindAlertPolicy,
	openslo.KindAlertCondition,
	openslo.KindAlertNotificationTarget,
}

// GetSupportedKinds returns a copy of the OpenSLO object kinds supported by v2alpha.
func GetSupportedKinds() []openslo.Kind {
	return slices.Clone(supportedKinds)
}

// Object is implemented by every OpenSLO v2alpha object and exposes its
// version-specific [Metadata].
type Object interface {
	openslo.Object
	// GetMetadata returns the object's version-specific metadata.
	GetMetadata() Metadata
}

// Metadata is the Kubernetes-style identifying metadata used by v2alpha
// objects. It contains a name and optional labels and annotations.
type Metadata struct {
	// Name identifies the object when other OpenSLO objects refer to it.
	Name string `json:"name"`
	// Labels optionally classifies the object with Kubernetes-style,
	// single-valued labels.
	Labels Labels `json:"labels,omitempty"`
	// Annotations optionally attaches non-identifying metadata with qualified
	// keys.
	Annotations Annotations `json:"annotations,omitempty"`
}

// Labels maps label keys to one string value each.
type Labels map[string]string

// Annotations maps annotation keys to arbitrary string values.
type Annotations map[string]string

// Operator specifies a comparison operation for an SLO objective or alert condition.
type Operator string

const (
	OperatorGT  Operator = "gt"
	OperatorLT  Operator = "lt"
	OperatorGTE Operator = "gte"
	OperatorLTE Operator = "lte"
)

var validOperators = []Operator{
	OperatorGT,
	OperatorLT,
	OperatorGTE,
	OperatorLTE,
}

var operatorValidation = govy.New(
	govy.For(govy.GetSelf[Operator]()).
		Rules(rules.OneOf(validOperators...)),
)

// Validate returns an error for an unsupported comparison operator.
func (o Operator) Validate() error {
	return operatorValidation.Validate(o)
}

func validationRulesAPIVersion[T openslo.Object](
	getter func(T) openslo.Version,
) govy.PropertyRules[openslo.Version, T] {
	return govy.For(getter).
		WithName("apiVersion").
		Required().
		Rules(rules.EQ(APIVersion))
}

func validationRulesKind[T openslo.Object](
	getter func(T) openslo.Kind, kind openslo.Kind,
) govy.PropertyRules[openslo.Kind, T] {
	return govy.For(getter).
		WithName("kind").
		Required().
		Rules(rules.EQ(kind))
}

func validationRulesMetadata[T any](getter func(T) Metadata) govy.PropertyRules[Metadata, T] {
	return govy.For(getter).
		WithName("metadata").
		Required().
		Include(
			govy.New(
				govy.For(func(m Metadata) string { return m.Name }).
					WithName("name").
					Required().
					Rules(rules.StringDNSLabel()),
				govy.For(func(m Metadata) Labels { return m.Labels }).
					WithName("labels").
					OmitEmpty().
					Include(labelsValidator()),
				govy.For(func(m Metadata) Annotations { return m.Annotations }).
					WithName("annotations").
					OmitEmpty().
					Include(annotationsValidator()),
			),
		)
}

var labelValueRegexp = regexp.MustCompile(`^([a-z0-9]([-._a-z0-9]{0,61}[a-z0-9])?)?$`)

func labelsValidator() govy.Validator[Labels] {
	return govy.New(
		govy.ForMap(govy.GetSelf[Labels]()).
			Cascade(govy.CascadeModeStop).
			RulesForKeys(rules.StringKubernetesQualifiedName()).
			RulesForValues(rules.StringMatchRegexp(labelValueRegexp).WithExamples("my-label", "my.domain_123-label")),
	)
}

func annotationsValidator() govy.Validator[Annotations] {
	return govy.New(
		govy.ForMap(govy.GetSelf[Annotations]()).
			Cascade(govy.CascadeModeStop).
			RulesForKeys(rules.StringKubernetesQualifiedName()),
	)
}
