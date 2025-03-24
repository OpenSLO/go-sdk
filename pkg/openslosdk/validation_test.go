package openslosdk

import (
	"testing"

	"github.com/OpenSLO/go-sdk/internal/assert"
	"github.com/OpenSLO/go-sdk/pkg/openslo"
	v1 "github.com/OpenSLO/go-sdk/pkg/openslo/v1"
	"github.com/OpenSLO/go-sdk/pkg/openslo/v1alpha"
	"github.com/OpenSLO/go-sdk/pkg/openslo/v2alpha"
)

func TestValidate(t *testing.T) {
	objects := []openslo.Object{
		v1alpha.NewService(
			v1alpha.Metadata{},
			v1alpha.ServiceSpec{},
		),
		v1.NewService(
			v1.Metadata{
				Name: "service 1",
			},
			v1.ServiceSpec{},
		),
		v2alpha.NewService(
			v2alpha.Metadata{
				Name: "service2",
				Labels: v2alpha.Labels{
					"invalid key": "value",
				},
			},
			v2alpha.ServiceSpec{},
		),
	}

	err := Validate(objects...)

	assert.Require(t, assert.Error(t, err))

	// nolint: lll
	expectedError := `Validation for openslo/v1alpha Service at index 0 has failed for the following properties:
  - 'metadata':
    - property is required but was empty
Validation for openslo/v1 Service at index 1 has failed for the following properties:
  - 'metadata.name' with value 'service 1':
    - string must match regular expression: '^[a-z0-9]([-a-z0-9]*[a-z0-9])?$' (e.g. 'my-name', '123-abc'); an RFC-1123 compliant label name must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character
Validation for openslo.com/v2alpha Service at index 2 has failed for the following properties:
  - 'metadata.labels.invalid key' with key 'invalid key':
    - name part string must match regular expression: '^([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$' (e.g. 'my.domain/MyName', 'MyName', 'my.name', '123-abc'); Kubernetes Qualified Name must consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character with an optional DNS subdomain prefix and '/'`
	assert.Equal(t, expectedError, err.Error())
}
