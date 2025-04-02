package openslo

import (
	"errors"
	"testing"

	"github.com/OpenSLO/go-sdk/internal/assert"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Version
		err   error
	}{
		{
			name:  "Valid v1",
			input: "openslo/v1",
			want:  VersionV1,
		},
		{
			name:  "Valid v2alpha",
			input: "openslo.com/v2alpha",
			want:  VersionV2alpha,
		},
		{
			name:  "Unsupported version",
			input: "openslo/v2alpha",
			want:  "",
			err:   errors.New("unsupported openslo.Version: openslo/v2alpha"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseVersion(tc.input)
			if tc.err != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.err, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestAllVersionEnumValuesAreValid(t *testing.T) {
	filePath := "../openslo/version.go"
	enumName := "Version"

	assert.AllEnumValuesAreEqual(t, filePath, enumName, func(s string) error {
		_, err := ParseVersion(s)
		return err
	})
}
