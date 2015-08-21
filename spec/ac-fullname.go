package spec
import (
	"strings"
	"encoding/json"
	"github.com/blablacar/cnt/log"
)

type ACFullname string

func (n *ACFullname) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		log.Get().Panic(err)
		return err
	}
	nn, err := NewACFullName(s)
	if err != nil {
		log.Get().Panic(err)
		return err
	}
	*n = *nn
	return nil
}

func (n ACFullname) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.String())
}

func (n ACFullname) String() string {
	return string(n)
}

/* example.com/yopla:1 */
func NewACFullName(s string) (*ACFullname, error) {
	n := ACFullname(s)
	return &n, nil
}

/* 1 */
func (n ACFullname) Version() string {
	split := strings.Split(string(n), ":")
	if (len(split) == 1) {
		return ""
	}
	return split[1]
}

/* yopla:1 */
func (n ACFullname) ShortNameId() string {
	return strings.Split(string(n), "/")[1]
}

/* yopla */
func (n ACFullname) ShortName() string {
	return strings.Split(n.Name(), "/")[1]
}

/* example.com/yopla */
func (n ACFullname) Name() string {
	return strings.Split(string(n), ":")[0]
}