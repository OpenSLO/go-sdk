package internal_test

import (
	"testing"

	"github.com/OpenSLO/go-sdk/internal"
	"github.com/OpenSLO/go-sdk/internal/assert"
	"github.com/OpenSLO/go-sdk/pkg/openslo"
	v1 "github.com/OpenSLO/go-sdk/pkg/openslo/v1"
	"github.com/OpenSLO/go-sdk/pkg/openslo/v1alpha"
	"github.com/OpenSLO/go-sdk/pkg/openslo/v2alpha"
)

func TestGetObjectName(t *testing.T) {
	tests := []struct {
		expected string
		object   openslo.Object
	}{
		{"v1alpha.Service 'foo'", v1alpha.Service{Metadata: v1alpha.Metadata{Name: "foo"}}},
		{"v1.Service 'foo'", v1.Service{Metadata: v1.Metadata{Name: "foo"}}},
		{"v1.Service", v1.Service{}},
		{"v2alpha.Service 'foo'", v2alpha.Service{Metadata: v2alpha.Metadata{Name: "foo"}}},
	}

	for _, tc := range tests {
		name := internal.GetObjectName(tc.object)
		assert.Equal(t, tc.expected, name)
	}
}
