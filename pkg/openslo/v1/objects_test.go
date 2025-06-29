package v1

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/nobl9/govy/pkg/govytest"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/OpenSLO/go-sdk/pkg/openslo"
)

func getValidationMessageRegexp(kind openslo.Kind) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf(strings.TrimSpace(`
(?s)Validation for v1.%s '.*' has failed for the following properties:
.*
`), kind))
}

func runMetadataTests[T openslo.Object](t *testing.T, path string, objectGetter func(m Metadata) T) {
	t.Run("name and display name", func(t *testing.T) {
		object := objectGetter(Metadata{
			Name:        strings.Repeat("MY SERVICE", 20),
			DisplayName: strings.Repeat("my-service", 20),
		})
		err := object.Validate()
		govytest.AssertError(t, err,
			govytest.ExpectedRuleError{
				PropertyName: path + ".name",
				Code:         rules.ErrorCodeStringDNSLabel,
			},
			govytest.ExpectedRuleError{
				PropertyName: path + ".displayName",
				Code:         rules.ErrorCodeStringMaxLength,
			},
		)
	})
	t.Run("labels", func(t *testing.T) {
		for name, test := range getLabelsTestCases(t, path+".labels") {
			t.Run(name, func(t *testing.T) {
				object := objectGetter(Metadata{
					Name:   "ok",
					Labels: test.labels,
				})
				test.Test(t, object)
			})
		}
	})
	t.Run("annotations", func(t *testing.T) {
		for name, test := range getAnnotationsTestCases(t, path+".annotations") {
			t.Run(name, func(t *testing.T) {
				object := objectGetter(Metadata{
					Name:        "ok",
					Annotations: test.annotations,
				})
				test.Test(t, object)
			})
		}
	})
}

type labelsTestCase struct {
	labels  Labels
	isValid bool
	error   govytest.ExpectedRuleError
}

func (tc labelsTestCase) Test(t *testing.T, object openslo.Object) {
	err := object.Validate()
	if tc.isValid {
		govytest.AssertNoError(t, err)
	} else {
		govytest.AssertError(t, err, tc.error)
	}
}

func getLabelsTestCases(t *testing.T, propertyPath string) map[string]labelsTestCase {
	t.Helper()
	validLabels := []Labels{
		{strings.Repeat("l", 63): {}},
		{"net": {"this", "that"}},
		{"net": {""}},
		{"net": {}},
		{"net_this.that-this": {}},
		{"net_.-this": {}},
		{"9net": {}},
		{"net9": {}},
		{"nEt": {}},
	}
	invalidLabels := []Labels{
		{strings.Repeat("l", 64): {}},
		{"net_": {}},
		{"net.": {}},
		{"net-": {}},
		{"_net": {}},
		{"-net": {}},
		{".net": {}},
	}
	testCases := make(map[string]labelsTestCase, len(validLabels)+len(invalidLabels))
	for _, labels := range validLabels {
		testCases[fmt.Sprintf("valid: %v", labels)] = labelsTestCase{
			labels:  labels,
			isValid: true,
		}
	}
	for _, labels := range invalidLabels {
		testCases[fmt.Sprintf("invalid: %v", labels)] = labelsTestCase{
			labels: labels,
			error: govytest.ExpectedRuleError{
				PropertyName: propertyPath + "." + escapeJSONPathKey(getMapFirstKey(labels)),
				IsKeyError:   true,
				Code:         rules.ErrorCodeStringMatchRegexp,
			},
		}
	}
	return testCases
}

type annotationsTestCase struct {
	annotations Annotations
	isValid     bool
	error       govytest.ExpectedRuleError
}

func (tc annotationsTestCase) Test(t *testing.T, object openslo.Object) {
	err := object.Validate()
	if tc.isValid {
		govytest.AssertNoError(t, err)
	} else {
		govytest.AssertError(t, err, tc.error)
	}
}

func getAnnotationsTestCases(t *testing.T, propertyPath string) map[string]annotationsTestCase {
	t.Helper()
	theLongestPrefix := strings.Repeat(strings.Repeat("l", 63)+".", 3)
	theLongestPrefix = theLongestPrefix[:len(theLongestPrefix)-1]
	validAnnotations := []Annotations{
		{strings.Repeat("l", 63): "this"},
		{theLongestPrefix + "/" + strings.Repeat("l", 63): ""},
		{"net": "this"},
		{"net": ""},
		{"9net": ""},
		{"net9": ""},
		{"nEt": ""},
		{"openslo.com/service": ""},
		{"domain/service": ""},
		{"domain.org/service": ""},
		{"domain.this.org/service": ""},
		{"domain.this.org/service.foo": ""},
		{"my-org.com/spec.indicator.metricSource": ""},
	}
	invalidAnnotations := []Annotations{
		{strings.Repeat("l", 64): ""},
		{strings.Repeat("l", 254) + "/net": ""},
		{strings.Repeat("l", 253) + "/" + strings.Repeat("l", 64): ""},
		{strings.Repeat("l", 254) + "/" + strings.Repeat("l", 63): ""},
		{"net_": ""},
		{"net.": ""},
		{"net-": ""},
		{"_net": ""},
		{"-net": ""},
		{".net": ""},
		{"openslo.com/": ""},
		{"openslo.com!/service": ""},
		{"-openslo.com/service": ""},
		{"_openslo.com/service": ""},
		{".openslo.com/service": ""},
		{"openslo.com./service": ""},
		{"openslo.com_/service": ""},
		{"openslo.com-/service": ""},
		{"openslo..this.com/service": ""},
		{"openslo.-this.com/service": ""},
		{"openslo._this.com/service": ""},
		{"openslo.this..com/service": ""},
		{"openslo.this-.com/service": ""},
		{"openslo.this_.com/service": ""},
		{"openslo_this.org/service": ""},
		{"openslo_this.my-org.com/service": ""},
	}
	testCases := make(map[string]annotationsTestCase, len(validAnnotations)+len(invalidAnnotations))
	for _, annotations := range validAnnotations {
		testCases[fmt.Sprintf("valid: %v", annotations)] = annotationsTestCase{
			annotations: annotations,
			isValid:     true,
		}
	}
	for _, annotations := range invalidAnnotations {
		testCases[fmt.Sprintf("invalid: %v", annotations)] = annotationsTestCase{
			annotations: annotations,
			error: govytest.ExpectedRuleError{
				PropertyName: propertyPath + "." + escapeJSONPathKey(getMapFirstKey(annotations)),
				IsKeyError:   true,
				Code:         rules.ErrorCodeStringMatchRegexp,
			},
		}
	}
	return testCases
}

func runOperatorTests[T openslo.Object](t *testing.T, path string, objectGetter func(o Operator) T) {
	t.Run("valid operator values", func(t *testing.T) {
		for _, op := range validOperators {
			object := objectGetter(op)
			err := object.Validate()
			govytest.AssertNoError(t, err)
		}
	})
	t.Run("invalid operator value", func(t *testing.T) {
		object := objectGetter("lessThan")
		err := object.Validate()
		govytest.AssertError(t, err, govytest.ExpectedRuleError{
			PropertyName: path,
			Code:         rules.ErrorCodeOneOf,
		})
	})
}

func getMapFirstKey[V any](l map[string]V) string {
	for k := range l {
		return k
	}
	return ""
}

func escapeJSONPathKey(v string) string {
	if strings.Contains(v, ".") {
		return "['" + v + "']"
	}
	return v
}
