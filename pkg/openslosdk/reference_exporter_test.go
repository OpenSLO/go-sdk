package openslosdk

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/OpenSLO/go-sdk/internal"
	"github.com/OpenSLO/go-sdk/internal/assert"
)

func TestReferenceExporter_Export(t *testing.T) {
	root := internal.FindModuleRoot()
	testDataPath := filepath.Join(root, "pkg", "openslosdk", "test_data", "export")

	tests := map[string]struct {
		filename string
	}{
		"v1: Alert Policies": {
			filename: "v1_alert_policies.yaml",
		},
		"v1: SLO": {
			filename: "v1_slo.yaml",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// Read input.
			inputPath := filepath.Join(testDataPath, "inputs", test.filename)
			inputFileData, err := os.ReadFile(inputPath)
			assert.Require(t, assert.NoError(t, err))
			inputObjects, err := Decode(bytes.NewReader(inputFileData), FormatYAML)
			assert.Require(t, assert.NoError(t, err))
			err = Validate(inputObjects...)
			assert.Require(t, assert.NoError(t, err))

			// Inline objects.
			exporter := NewReferenceExporter(inputObjects...)
			inlinedObjects := exporter.Export()
			assert.Require(t, assert.NotEmpty(t, inlinedObjects))

			// Read output.
			outputPath := filepath.Join(testDataPath, "outputs", test.filename)
			outputsFileData, err := os.ReadFile(outputPath)
			assert.Require(t, assert.NoError(t, err))
			outputObjects, err := Decode(bytes.NewReader(outputsFileData), FormatYAML)
			assert.Require(t, assert.NoError(t, err))
			err = Validate(outputObjects...)
			assert.Require(t, assert.NoError(t, err))

			// Check.
			err = Validate(inlinedObjects...)
			assert.Require(t, assert.NoError(t, err))
			var buf bytes.Buffer
			err = Encode(&buf, FormatYAML, inlinedObjects...)
			assert.Require(t, assert.NoError(t, err))
			assert.Equal(t, string(outputsFileData), buf.String())
		})
	}
}
