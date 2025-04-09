package openslosdk

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/OpenSLO/go-sdk/internal"
	"github.com/OpenSLO/go-sdk/internal/assert"
)

func TestReferenceResolver_Inline(t *testing.T) {
	root := internal.FindModuleRoot()
	testDataPath := filepath.Join(root, "pkg", "openslosdk", "test_data", "inline")

	tests := map[string]struct {
		filename    string
		resolverMod func(*ReferenceResolver) *ReferenceResolver
		err         error
	}{
		"valid Alert Policies - keep refs": {
			filename: "v1_alert_policies_keep_refs.yaml",
		},
		"valid Alert Policies - remove refs": {
			filename:    "v1_alert_policies_remove_refs.yaml",
			resolverMod: func(r *ReferenceResolver) *ReferenceResolver { return r.RemoveReferencedObjects() },
		},
		"non-existing AlertNotificationTarget for Alert Policies": {
			filename: "v1_alert_policies_invalid_target.yaml",
			err: errors.New("failed to inline v1.AlertPolicy 'invalid-target': v1.AlertNotificationTarget" +
				" 'devs-email-notification' referenced at 'spec.notificationTargets[1].targetRef' does not exist"),
		},
		"non-existing AlertCondition for Alert Policies": {
			filename: "v1_alert_policies_invalid_condition.yaml",
			err: errors.New("failed to inline v1.AlertPolicy 'invalid-condition': v1.AlertCondition" +
				" 'cpu-usage-breach' referenced at 'spec.conditions[0].conditionRef' does not exist"),
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
			resolver := NewReferenceResolver(inputObjects...)
			if test.resolverMod != nil {
				resolver = test.resolverMod(resolver)
			}
			inlinedObjects, err := resolver.Inline()
			switch {
			case test.err == nil:
				assert.Require(t, assert.NoError(t, err))
				assert.Require(t, assert.NotEmpty(t, inlinedObjects))
			default:
				assert.Require(t, assert.Error(t, err))
				assert.Equal(t, test.err.Error(), err.Error())
				return
			}

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
