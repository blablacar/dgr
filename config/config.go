package config
import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"github.com/blablacar/cnt/utils"
)

var cntConfig = new(CntConfig)

type CntConfig struct {
	Push     struct {
					Type      	string                `yaml:"type,omitempty"`
					Url 		string                `yaml:"url,omitempty"`
					Username    string                `yaml:"username,omitempty"`
					Password	string                `yaml:"password,omitempty"`
				}  							`yaml:"push,omitempty"`
}

func GetConfig() *CntConfig {
	return cntConfig
}

func (c *CntConfig) Load() {
	if source, err := ioutil.ReadFile(utils.UserHomeOrFatal() + "/.config/cnt/config.yml"); err == nil {
		err = yaml.Unmarshal([]byte(source), &c)
		if err != nil {
			panic(err)
		}
	}
}
