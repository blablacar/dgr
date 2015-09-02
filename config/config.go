package config

import (
	"github.com/blablacar/cnt/log"
	"github.com/blablacar/cnt/utils"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"runtime"
)

var cntConfig CntConfig

type CntConfig struct {
	Path    string
	AciPath string
	Push    struct {
		Type     string `yaml:"type,omitempty"`
		Url      string `yaml:"url,omitempty"`
		Username string `yaml:"username,omitempty"`
		Password string `yaml:"password,omitempty"`
	} `yaml:"push,omitempty"`
}

func GetConfig() *CntConfig {
	return &cntConfig
}

func (c *CntConfig) Load() {
}

func init() {
	cntConfig = CntConfig{}
	switch runtime.GOOS {
	case "windows":
		cntConfig.Path = utils.UserHomeOrFatal() + "/AppData/Local/Cnt"
	case "darwin":
		cntConfig.Path = utils.UserHomeOrFatal() + "/Library/Cnt"
	case "linux":
		cntConfig.Path = utils.UserHomeOrFatal() + "/.config/cnt"
	default:
		log.Get().Panic("Unsupported OS, please fill a bug repost")
	}
	cntConfig.AciPath = cntConfig.Path + "/aci"

	if source, err := ioutil.ReadFile(cntConfig.Path + "/config.yml"); err == nil {
		err = yaml.Unmarshal([]byte(source), &cntConfig)
		if err != nil {
			panic(err)
		}
	}

	log.Get().Debug("Home folder is " + cntConfig.Path)
}
