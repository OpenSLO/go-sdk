package openslo

import (
	"errors"
	"testing"

	"github.com/OpenSLO/go-sdk/internal/assert"
)

func TestParseKind(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Kind
		err   error
	}{
		{
			name:  "Valid SLO",
			input: "SLO",
			want:  KindSLO,
		},
		{
			name:  "Valid Service",
			input: "Service",
			want:  KindService,
		},
		{
			name:  "Unsupported Kind",
			input: "UnknownKind",
			want:  "",
			err:   errors.New("unsupported openslo.Kind: UnknownKind"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseKind(tc.input)
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

func TestAllKindEnumValuesAreValid(t *testing.T) {
	filePath := "../openslo/kind.go"
	enumName := "Kind"

	assert.AllEnumValuesAreEqual(t, filePath, enumName, func(s string) error {
		_, err := ParseKind(s)
		return err
	})
}
