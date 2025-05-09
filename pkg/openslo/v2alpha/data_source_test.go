package v2alpha

import (
	"strings"
	"testing"

	"github.com/nobl9/govy/pkg/govytest"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/OpenSLO/go-sdk/internal/assert"
	"github.com/OpenSLO/go-sdk/pkg/openslo"
)

var dataSourceValidationMessageRegexp = getValidationMessageRegexp(openslo.KindDataSource)

func TestDataSource_Validate_Ok(t *testing.T) {
	err := validDataSource().Validate()
	govytest.AssertNoError(t, err)
}

func TestDataSource_Validate_VersionAndKind(t *testing.T) {
	dataSource := validDataSource()
	dataSource.APIVersion = "v0.1"
	dataSource.Kind = openslo.KindSLO
	err := dataSource.Validate()
	assert.Require(t, assert.Error(t, err))
	assert.True(t, dataSourceValidationMessageRegexp.MatchString(err.Error()))
	govytest.AssertError(t, err,
		govytest.ExpectedRuleError{
			PropertyName: "apiVersion",
			Code:         rules.ErrorCodeEqualTo,
		},
		govytest.ExpectedRuleError{
			PropertyName: "kind",
			Code:         rules.ErrorCodeEqualTo,
		},
	)
}

func TestDataSource_Validate_Metadata(t *testing.T) {
	runMetadataTests(t, "metadata", func(m Metadata) DataSource {
		dataSource := validDataSource()
		dataSource.Metadata = m
		return dataSource
	})
}

func TestDataSource_Validate_Spec(t *testing.T) {
	runDataSourceSpecTests(t, "spec", func(d DataSourceSpec) DataSource {
		dataSource := validDataSource()
		dataSource.Spec = d
		return dataSource
	})
}

func runDataSourceSpecTests[T openslo.Object](
	t *testing.T,
	path string,
	objectGetter func(d DataSourceSpec) T,
) {
	t.Helper()

	t.Run("description ok", func(t *testing.T) {
		dataSource := validDataSource()
		dataSource.Spec.Description = strings.Repeat("A", 1050)
		object := objectGetter(dataSource.Spec)
		err := object.Validate()
		govytest.AssertNoError(t, err)
	})
	t.Run("description too long", func(t *testing.T) {
		dataSource := validDataSource()
		dataSource.Spec.Description = strings.Repeat("A", 1051)
		object := objectGetter(dataSource.Spec)
		err := object.Validate()
		govytest.AssertError(t, err, govytest.ExpectedRuleError{
			PropertyName: path + ".description",
			Code:         rules.ErrorCodeStringMaxLength,
		})
	})
	t.Run("missing fields", func(t *testing.T) {
		dataSource := validDataSource()
		dataSource.Spec.Type = ""
		dataSource.Spec.ConnectionDetails = nil
		object := objectGetter(dataSource.Spec)
		err := object.Validate()
		govytest.AssertError(t, err,
			govytest.ExpectedRuleError{
				PropertyName: path + ".type",
				Code:         rules.ErrorCodeRequired,
			},
			govytest.ExpectedRuleError{
				PropertyName: path + ".connectionDetails",
				Code:         rules.ErrorCodeRequired,
			},
		)
	})
}

func validDataSource() DataSource {
	return NewDataSource(
		Metadata{
			Name: "prometheus",
			Labels: Labels{
				"team": "team-a",
				"env":  "prod",
			},
			Annotations: Annotations{
				"key": "value",
			},
		},
		DataSourceSpec{
			Type:              "Prometheus",
			ConnectionDetails: []byte(`[{"url":"http://prometheus.example.com"}]`),
		},
	)
}
