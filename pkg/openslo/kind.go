package openslo

import (
	"fmt"
)

// Kind represents all the object kinds defined by OpenSLO specification.
// Keep in mind not all specification versions support every [Kind].
type Kind string

const (
	KindSLO                     Kind = "SLO"
	KindSLI                     Kind = "SLI"
	KindDataSource              Kind = "DataSource"
	KindService                 Kind = "Service"
	KindAlertPolicy             Kind = "AlertPolicy"
	KindAlertCondition          Kind = "AlertCondition"
	KindAlertNotificationTarget Kind = "AlertNotificationTarget"
)

func ParseKind(s string) (Kind, error) {
	kind := Kind(s)
	if err := kind.Validate(); err != nil {
		return "", err
	}
	return kind, nil
}

func (k Kind) String() string {
	return string(k)
}

func (k Kind) Validate() error {
	switch k {
	case KindSLO,
		KindSLI,
		KindDataSource,
		KindService,
		KindAlertPolicy,
		KindAlertCondition,
		KindAlertNotificationTarget:
		return nil
	default:
		return fmt.Errorf("unsupported %[1]T: %[1]s", k)
	}
}

// UnmarshalText implements the text [encoding.TextUnmarshaler] interface.
func (k *Kind) UnmarshalText(text []byte) error {
	tmp, err := ParseKind(string(text))
	if err != nil {
		return err
	}
	*k = tmp
	return nil
}
