package v2alpha

import (
	"fmt"
	"time"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// ParseDurationShorthand parses s into a [DurationShorthand].
func ParseDurationShorthand(s string) (DurationShorthand, error) {
	d := new(DurationShorthand)
	err := d.UnmarshalText([]byte(s))
	return *d, err
}

// NewDurationShorthand returns a shorthand with the supplied value and unit
// without validating them.
func NewDurationShorthand(value int, unit DurationShorthandUnit) DurationShorthand {
	return DurationShorthand{
		unit:  unit,
		value: value,
	}
}

// DurationShorthand represents a duration as an integer with suffix "m", "h",
// "d", or "w". This SDK accepts zero and encodes it as empty text. OpenSLO
// specifies a positive integer.
type DurationShorthand struct {
	unit  DurationShorthandUnit
	value int
}

// GetUnit returns the shorthand's [DurationShorthandUnit].
func (d *DurationShorthand) GetUnit() DurationShorthandUnit {
	return d.unit
}

// GetValue returns the shorthand's integer value.
func (d *DurationShorthand) GetValue() int {
	return d.value
}

// UnmarshalText decodes text into d as an [encoding.TextUnmarshaler].
func (d *DurationShorthand) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return nil
	}
	if n, err := fmt.Sscanf(string(text), "%d%s", &d.value, &d.unit); err != nil || n != 2 {
		return fmt.Errorf("invalid duration shorthand: %s, expected [0-9]+[mhdwMQY]", text)
	}
	return nil
}

// MarshalText encodes d as an [encoding.TextMarshaler].
func (d DurationShorthand) MarshalText() ([]byte, error) {
	if d.value == 0 {
		return []byte{}, nil
	}
	return []byte(d.String()), nil
}

// String returns the encoded shorthand as required by [fmt.Stringer].
func (d DurationShorthand) String() string {
	if d.value == 0 {
		return ""
	}
	return fmt.Sprintf("%d%s", d.value, d.unit)
}

// Duration returns the equivalent [time.Duration] and panics for an unsupported unit.
func (d DurationShorthand) Duration() time.Duration {
	switch d.unit {
	case DurationShorthandUnitMinute:
		return time.Duration(d.value) * time.Minute
	case DurationShorthandUnitHour:
		return time.Duration(d.value) * time.Hour
	case DurationShorthandUnitDay:
		return time.Duration(d.value) * 24 * time.Hour
	case DurationShorthandUnitWeek:
		return time.Duration(d.value) * 7 * 24 * time.Hour
	default:
		panic("invalid unit")
	}
}

// DurationShorthandUnit identifies a minute, hour, day, or week unit for
// [DurationShorthand].
type DurationShorthandUnit string

const (
	DurationShorthandUnitMinute DurationShorthandUnit = "m"
	DurationShorthandUnitHour   DurationShorthandUnit = "h"
	DurationShorthandUnitDay    DurationShorthandUnit = "d"
	DurationShorthandUnitWeek   DurationShorthandUnit = "w"
)

var validDurationUnits = []DurationShorthandUnit{
	DurationShorthandUnitMinute,
	DurationShorthandUnitHour,
	DurationShorthandUnitDay,
	DurationShorthandUnitWeek,
}

// Validate returns an error for an invalid duration shorthand.
func (d DurationShorthand) Validate() error {
	return durationShortHandValidation.Validate(d)
}

var durationShortHandValidation = govy.New(
	govy.For(func(d DurationShorthand) DurationShorthandUnit { return d.unit }).
		WithName("unit").
		Required().
		Rules(rules.OneOf(validDurationUnits...)),
	govy.For(func(d DurationShorthand) int { return d.value }).
		WithName("value").
		Rules(rules.GTE(0)),
)
