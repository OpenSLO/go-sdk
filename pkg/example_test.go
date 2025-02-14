package pkg_test

import (
	"bytes"
	"os"

	v1 "github.com/OpenSLO/go-sdk/pkg/openslo/v1"
	"github.com/OpenSLO/go-sdk/pkg/openslosdk"
)

const serviceDefinition = `
apiVersion: openslo/v1
kind: Service
metadata:
  name: web-app
  displayName: React Web Application
spec:
  description: Web application built in React
`

func Example() {
	// Decode the Service.
	objects, err := openslosdk.Decode(bytes.NewBufferString(serviceDefinition), openslosdk.FormatYAML)
	if err != nil {
		panic(err)
	}

	// Define Data Source in code.
	dataSource := v1.NewDataSource(
		v1.Metadata{
			Name: "prometheus",
			Labels: v1.Labels{
				"env": {"prod"},
			},
		},
		v1.DataSourceSpec{
			Description:       "Production Prometheus",
			Type:              "Prometheus",
			ConnectionDetails: []byte(`[{"url":"http://prometheus.example.com"}]`),
		},
	)

	// Add Data Source to objects.
	objects = append(objects, dataSource)

	// Validate the objects.
	if err = openslosdk.Validate(objects...); err != nil {
		panic(err)
	}

	// Write objects to stdout in JSON format.
	err = openslosdk.Encode(os.Stdout, openslosdk.FormatJSON, objects...)
	if err != nil {
		panic(err)
	}

	// Output:
	// [
	//   {
	//     "apiVersion": "openslo/v1",
	//     "kind": "Service",
	//     "metadata": {
	//       "name": "web-app",
	//       "displayName": "React Web Application"
	//     },
	//     "spec": {
	//       "description": "Web application built in React"
	//     }
	//   },
	//   {
	//     "apiVersion": "openslo/v1",
	//     "kind": "DataSource",
	//     "metadata": {
	//       "name": "prometheus",
	//       "labels": {
	//         "env": [
	//           "prod"
	//         ]
	//       }
	//     },
	//     "spec": {
	//       "description": "Production Prometheus",
	//       "type": "Prometheus",
	//       "connectionDetails": [
	//         {
	//           "url": "http://prometheus.example.com"
	//         }
	//       ]
	//     }
	//   }
	// ]
}

func ptr[T any](v T) *T { return &v }
