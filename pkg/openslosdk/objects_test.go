package openslosdk

import (
	"testing"

	"github.com/OpenSLO/go-sdk/internal/assert"
	"github.com/OpenSLO/go-sdk/pkg/openslo"
	v1 "github.com/OpenSLO/go-sdk/pkg/openslo/v1"
)

func TestFilterByKind(t *testing.T) {
	tests := map[string]struct {
		objects  []openslo.Object
		expected any
	}{
		"Filter v1.SLO": {
			objects: []openslo.Object{
				v1.Service{},
				v1.DataSource{},
				mockObject{},
				v1.Service{},
			},
			expected: []v1.SLO(nil),
		},
		"Filter v1.Service": {
			objects: []openslo.Object{
				v1.Service{Metadata: v1.Metadata{Name: "service1"}},
				v1.DataSource{Metadata: v1.Metadata{Name: "dataSource"}},
				mockObject{},
				v1.Service{Metadata: v1.Metadata{Name: "service2"}},
			},
			expected: []v1.Service{
				{Metadata: v1.Metadata{Name: "service1"}},
				{Metadata: v1.Metadata{Name: "service2"}},
			},
		},
		"Filter v1.DataSource": {
			objects: []openslo.Object{
				v1.Service{},
				v1.DataSource{Metadata: v1.Metadata{Name: "dataSource"}},
				mockObject{},
				v1.Service{},
			},
			expected: []v1.DataSource{
				{Metadata: v1.Metadata{Name: "dataSource"}},
			},
		},
		"Filter mockObject": {
			objects: []openslo.Object{
				v1.Service{},
				v1.DataSource{},
				mockObject{},
				v1.Service{},
			},
			expected: []mockObject{
				{},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var result any
			switch name {
			case "Filter v1.SLO":
				result = FilterByKind[v1.SLO](tc.objects)
			case "Filter v1.Service":
				result = FilterByKind[v1.Service](tc.objects)
			case "Filter v1.DataSource":
				result = FilterByKind[v1.DataSource](tc.objects)
			case "Filter mockObject":
				result = FilterByKind[mockObject](tc.objects)
			default:
				t.Fatalf("unexpected test case: %s", name)
			}
			assert.Equal(t, tc.expected, result)
		})
	}
}

type mockObject struct {
	openslo.Object
}
