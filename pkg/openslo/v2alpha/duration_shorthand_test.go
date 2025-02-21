package v2alpha

import (
	"testing"
	"time"

	"github.com/nobl9/govy/pkg/govytest"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/OpenSLO/go-sdk/internal/assert"
	"github.com/OpenSLO/go-sdk/pkg/openslo"
)

var parseDurationShorthandTestCases = []struct {
	input    string
	expected DurationShorthand
	err      bool
}{
	{"", DurationShorthand{value: 0, unit: ""}, false},
	{"0w", DurationShorthand{value: 0, unit: DurationShorthandUnitWeek}, false},
	{"10m", DurationShorthand{value: 10, unit: DurationShorthandUnitMinute}, false},
	{"5h", DurationShorthand{value: 5, unit: DurationShorthandUnitHour}, false},
	{"2d", DurationShorthand{value: 2, unit: DurationShorthandUnitDay}, false},
	{"1w", DurationShorthand{value: 1, unit: DurationShorthandUnitWeek}, false},
	{"invalid", DurationShorthand{}, true},
}

func TestParseDurationShorthand(t *testing.T) {
	for _, tc := range parseDurationShorthandTestCases {
		d, err := ParseDurationShorthand(tc.input)
		if tc.err {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, d)
		}
	}
}

func TestDurationShorthandUnmarshalText(t *testing.T) {
	for _, tc := range parseDurationShorthandTestCases {
		var d DurationShorthand
		err := d.UnmarshalText([]byte(tc.input))
		if tc.err {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, d)
		}
	}
}

var encodeDurationShorthandTestCases = []struct {
	input    DurationShorthand
	expected string
}{
	{DurationShorthand{value: 0, unit: DurationShorthandUnitWeek}, ""},
	{DurationShorthand{value: 10, unit: DurationShorthandUnitMinute}, "10m"},
	{DurationShorthand{value: 5, unit: DurationShorthandUnitHour}, "5h"},
	{DurationShorthand{value: 2, unit: DurationShorthandUnitDay}, "2d"},
	{DurationShorthand{value: 1, unit: DurationShorthandUnitWeek}, "1w"},
}

func TestDurationShorthandMarshalText(t *testing.T) {
	for _, tc := range encodeDurationShorthandTestCases {
		text, err := tc.input.MarshalText()
		assert.NoError(t, err)
		assert.Equal(t, tc.expected, string(text))
	}
}

func TestDurationShorthandString(t *testing.T) {
	for _, tc := range encodeDurationShorthandTestCases {
		assert.Equal(t, tc.expected, tc.input.String())
	}
}

func TestDurationShorthandDuration(t *testing.T) {
	tests := []struct {
		input    DurationShorthand
		expected time.Duration
	}{
		{DurationShorthand{value: 0, unit: DurationShorthandUnitWeek}, 0},
		{DurationShorthand{value: 10, unit: DurationShorthandUnitMinute}, 10 * time.Minute},
		{DurationShorthand{value: 5, unit: DurationShorthandUnitHour}, 5 * time.Hour},
		{DurationShorthand{value: 2, unit: DurationShorthandUnitDay}, 2 * 24 * time.Hour},
		{DurationShorthand{value: 1, unit: DurationShorthandUnitWeek}, 7 * 24 * time.Hour},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.expected, tc.input.Duration())
	}
}

func TestDurationShorthand_GetUnit(t *testing.T) {
	dur := DurationShorthand{unit: DurationShorthandUnitWeek}
	assert.Equal(t, DurationShorthandUnitWeek, dur.GetUnit())
}

func TestDurationShorthand_GetValue(t *testing.T) {
	dur := DurationShorthand{value: 12}
	assert.Equal(t, 12, dur.GetValue())
}

func runDurationShorthandTests[T openslo.Object](t *testing.T, path string, objectGetter func(d DurationShorthand) T) {
	t.Helper()

	tests := []struct {
		input        DurationShorthand
		expectedErrs []govytest.ExpectedRuleError
	}{
		{DurationShorthand{value: 0, unit: DurationShorthandUnitWeek}, nil},
		{DurationShorthand{value: 10, unit: DurationShorthandUnitMinute}, nil},
		{DurationShorthand{value: 5, unit: DurationShorthandUnitHour}, nil},
		{DurationShorthand{value: 2, unit: DurationShorthandUnitDay}, nil},
		{DurationShorthand{value: 1, unit: DurationShorthandUnitWeek}, nil},
		{
			DurationShorthand{value: 1, unit: "M"},
			[]govytest.ExpectedRuleError{{PropertyName: "unit", Code: rules.ErrorCodeOneOf}},
		},
		{
			DurationShorthand{value: 1, unit: "Q"},
			[]govytest.ExpectedRuleError{{PropertyName: "unit", Code: rules.ErrorCodeOneOf}},
		},
		{
			DurationShorthand{value: 1, unit: "Y"},
			[]govytest.ExpectedRuleError{{PropertyName: "unit", Code: rules.ErrorCodeOneOf}},
		},
		{
			DurationShorthand{value: -1, unit: DurationShorthandUnitMinute},
			[]govytest.ExpectedRuleError{{PropertyName: "value", Code: rules.ErrorCodeGreaterThanOrEqualTo}},
		},
		{
			DurationShorthand{value: 1, unit: ""},
			[]govytest.ExpectedRuleError{{PropertyName: "unit", Code: rules.ErrorCodeRequired}},
		},
		{
			DurationShorthand{value: 1, unit: "invalid"},
			[]govytest.ExpectedRuleError{{PropertyName: "unit", Code: rules.ErrorCodeOneOf}},
		},
	}

	for _, tc := range tests {
		object := objectGetter(tc.input)
		err := object.Validate()
		if tc.expectedErrs != nil {
			for i := range tc.expectedErrs {
				tc.expectedErrs[i].PropertyName = path + "." + tc.expectedErrs[i].PropertyName
			}
			govytest.AssertError(t, err, tc.expectedErrs...)
		} else {
			govytest.AssertNoError(t, err)
		}
	}
}
