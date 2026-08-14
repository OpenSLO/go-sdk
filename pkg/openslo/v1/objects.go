package v1

import (
	"encoding/json"
	"regexp"
	"slices"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/OpenSLO/go-sdk/pkg/openslo"
)

// APIVersion is the OpenSLO v1 API version.
const APIVersion = openslo.VersionV1

var supportedKinds = []openslo.Kind{
	openslo.KindSLO,
	openslo.KindSLI,
	openslo.KindDataSource,
	openslo.KindService,
	openslo.KindAlertPolicy,
	openslo.KindAlertCondition,
	openslo.KindAlertNotificationTarget,
}

// GetSupportedKinds returns a copy of the object kinds supported by this package.
func GetSupportedKinds() []openslo.Kind {
	return slices.Clone(supportedKinds)
}

// Object is an OpenSLO v1 object with accessible [Metadata].
type Object interface {
	openslo.Object
	GetMetadata() Metadata
}

// Metadata identifies and describes an OpenSLO v1 object.
type Metadata struct {
	// Name identifies the object.
	Name string `json:"name"`
	// DisplayName is the object's human-readable name.
	DisplayName string `json:"displayName,omitempty"`
	// Labels contains user-defined metadata for describing and grouping the object.
	Labels Labels `json:"labels,omitempty"`
	// Annotations contains implementation- or system-specific metadata.
	Annotations Annotations `json:"annotations,omitempty"`
}

// Labels maps metadata keys to one or more values.
type Labels map[string]Label

// Annotations maps metadata keys to implementation- or system-specific values.
type Annotations map[string]string

// Label is the set of values assigned to one label key.
// JSON decoding accepts either a single string or an array of strings.
type Label []string

// UnmarshalJSON decodes a string or string array into a Label.
func (a *Label) UnmarshalJSON(data []byte) error {
	var multi []string
	if err := json.Unmarshal(data, &multi); err != nil {
		var single string
		if err = json.Unmarshal(data, &single); err != nil {
			return err
		}
		*a = []string{single}
	} else {
		*a = multi
	}
	return nil
}

// Operator identifies a comparison applied to a metric value.
type Operator string

const (
	// OperatorGT selects greater than.
	OperatorGT Operator = "gt"
	// OperatorLT selects less than.
	OperatorLT Operator = "lt"
	// OperatorGTE selects greater than or equal to.
	OperatorGTE Operator = "gte"
	// OperatorLTE selects less than or equal to.
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

// Validate returns an error for an unsupported operator.
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
				govy.For(func(m Metadata) string { return m.DisplayName }).
					WithName("displayName").
					OmitEmpty().
					Rules(rules.StringMaxLength(63)),
				govy.For(func(m Metadata) Labels { return m.Labels }).
					WithName("labels").
					Include(labelsValidator()),
				govy.For(func(m Metadata) Annotations { return m.Annotations }).
					WithName("annotations").
					Include(annotationsValidator()),
			),
		)
}

var (
	labelKeyRegexp            = regexp.MustCompile(`^[a-zA-Z0-9]([-._a-zA-Z0-9]{0,61}[a-zA-Z0-9])?$`)
	annotationKeyLengthRegexp = regexp.MustCompile(`^(.{0,253}/)?.{0,63}$`)
	// nolint: lll
	annotationKeyRegexp = regexp.MustCompile(
		`^([a-z0-9]([-a-z0-9]{0,61}[a-z0-9])?(\.[a-z0-9]([-a-z0-9]{0,61}[a-z0-9])?)*/)?[a-zA-Z0-9]([-._a-zA-Z0-9]{0,61}[a-zA-Z0-9])?$`,
	)
)

func labelsValidator() govy.Validator[Labels] {
	return govy.New(
		govy.ForMap(govy.GetSelf[Labels]()).
			RulesForKeys(rules.StringMatchRegexp(labelKeyRegexp)),
	)
}

func annotationsValidator() govy.Validator[Annotations] {
	return govy.New(
		govy.ForMap(govy.GetSelf[Annotations]()).
			Cascade(govy.CascadeModeStop).
			RulesForKeys(
				rules.StringMatchRegexp(annotationKeyLengthRegexp),
				rules.StringMatchRegexp(annotationKeyRegexp).
					WithExamples(
						"my-domain.org/my-key",
						"openslo.com/annotation",
					),
			),
	)
}
