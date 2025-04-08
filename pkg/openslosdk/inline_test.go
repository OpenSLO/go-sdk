package openslosdk

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/OpenSLO/go-sdk/internal"
	"github.com/OpenSLO/go-sdk/internal/assert"
)

func TestInlineObjects(t *testing.T) {
	root := internal.FindModuleRoot()
	testDataPath := filepath.Join(root, "pkg", "openslosdk", "test_data", "inline")

	inputs := listAllFilePathsInDir(t, filepath.Join(testDataPath, "inputs"))
	outputs := listAllFilePathsInDir(t, filepath.Join(testDataPath, "outputs"))
	assert.Require(t, assert.Len(t, inputs, len(outputs)))

	for i, inputPath := range inputs {
		t.Run(filepath.Base(inputPath), func(t *testing.T) {
			inputFileData, err := os.ReadFile(inputPath)
			assert.Require(t, assert.NoError(t, err))

			outputsFileData, err := os.ReadFile(outputs[i])
			assert.Require(t, assert.NoError(t, err))

			inputObjects, err := Decode(bytes.NewReader(inputFileData), FormatYAML)
			assert.Require(t, assert.NoError(t, err))
			err = Validate(inputObjects...)
			assert.Require(t, assert.NoError(t, err))

			outputObjects, err := Decode(bytes.NewReader(outputsFileData), FormatYAML)
			assert.Require(t, assert.NoError(t, err))
			err = Validate(outputObjects...)
			assert.Require(t, assert.NoError(t, err))

			inlinedObjects, err := InlineObjects(inputObjects...)
			assert.Require(t, assert.NoError(t, err))
			assert.Require(t, assert.NotEmpty(t, inlinedObjects))

			err = Validate(inlinedObjects...)
			assert.Require(t, assert.NoError(t, err))

			var buf bytes.Buffer
			err = Encode(&buf, FormatYAML, inlinedObjects...)
			assert.Require(t, assert.NoError(t, err))

			assert.Equal(t, string(outputsFileData), buf.String())
		})
	}
}

func listAllFilePathsInDir(t *testing.T, dir string) []string {
	t.Helper()

	entries, err := os.ReadDir(dir)
	assert.Require(t, assert.NoError(t, err))

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		files = append(files, filepath.Join(dir, entry.Name()))
	}
	return files
}
