package config
import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"github.com/blablacar/cnt/utils"
	"runtime"
	"github.com/blablacar/cnt/log"
)

var cntConfig = new(CntConfig)

type CntConfig struct {
	Push struct {
			 Type     string                `yaml:"type,omitempty"`
			 Url      string                `yaml:"url,omitempty"`
			 Username string                `yaml:"username,omitempty"`
			 Password string                `yaml:"password,omitempty"`
		 }                            `yaml:"push,omitempty"`
}

func GetConfig() *CntConfig {
	return cntConfig
}

func (c *CntConfig) Load() {
	var cntHome string
	switch runtime.GOOS {
	case "windows":
		cntHome = utils.UserHomeOrFatal() + "/AppData/Local/Cnt";
	case "darwin":
		cntHome = utils.UserHomeOrFatal() + "/Library/Cnt";
	case "linux":
		cntHome = utils.UserHomeOrFatal() + "/.config/cnt";
	default:
		log.Get().Panic("Unsupported OS, please fill a bug repost")
	}
	log.Get().Debug("Home folder is " + cntHome)

	if source, err := ioutil.ReadFile(cntHome + "/config.yml"); err == nil {
		err = yaml.Unmarshal([]byte(source), &c)
		if err != nil {
			panic(err)
		}
	}
}
