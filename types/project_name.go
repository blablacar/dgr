package types
import (
	"encoding/json"
	"fmt"
	"strings"
)


var (
	ErrNoACKind = ProjectNameError("ProjectName must be set")
)

// ACKind wraps a string to define a field which must be set with one of
// several ACKind values. If it is unset, or has an invalid value, the field
// will refuse to marshal/unmarshal.
type AciName string

func (a AciName) String() string {
	return string(a)
}

func (a AciName) ShortName() string {
	split := strings.Split(string(a), "/")
	return split[1]
}

func (a AciName) assertValid() error {
	s := a.String()
	switch s {
	case "ImageManifest", "PodManifest":
		return nil
	case "":
		return ErrNoACKind
	default:
		msg := fmt.Sprintf("bad ACKind: %s", s)
		return ProjectNameError(msg)
	}
}

func (a AciName) MarshalJSON() ([]byte, error) {
	if err := a.assertValid(); err != nil {
		return nil, err
	}
	return json.Marshal(a.String())
}

func (a *AciName) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	na := AciName(s)
	if err := na.assertValid(); err != nil {
		return err
	}
	*a = na
	return nil
}
