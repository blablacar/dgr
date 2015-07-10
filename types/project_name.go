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
type ProjectName string

func (a ProjectName) String() string {
	return string(a)
}

func (a ProjectName) ShortName() string {
	split := strings.Split(string(a), "/")
	return split[1]
}

func (a ProjectName) assertValid() error {
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

func (a ProjectName) MarshalJSON() ([]byte, error) {
	if err := a.assertValid(); err != nil {
		return nil, err
	}
	return json.Marshal(a.String())
}

func (a *ProjectName) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	na := ProjectName(s)
	if err := na.assertValid(); err != nil {
		return err
	}
	*a = na
	return nil
}
