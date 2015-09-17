package config

import (
	"github.com/blablacar/cnt/log"
	"github.com/blablacar/cnt/utils"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

var cntConfig CntConfig

type CntConfig struct {
	Path string
	Push struct {
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
	cntConfig = CntConfig{Path:"/root/.config/cnt"}
	user := os.Getenv("SUDO_USER")
	if user != "" {
		home, err := utils.ExecCmdGetOutput("bash", "-c", "echo ~"+user)
		if err != nil {
			log.Get().Panic("Cannot find user home", err)
		}
		cntConfig.Path = home+"/.config/cnt"
	}
	//	switch runtime.GOOS {
	//	case "windows":
	//		cntConfig.Path = utils.UserHomeOrFatal() + "/AppData/Local/Cnt"
	//	case "darwin":
	//		cntConfig.Path = utils.UserHomeOrFatal() + "/Library/Cnt"
	//	case "linux":
	//		cntConfig.Path = utils.UserHomeOrFatal() + "/.config/cnt"
	//	default:
	//		log.Get().Panic("Unsupported OS, please fill a bug repost")
	//	}

	if source, err := ioutil.ReadFile(cntConfig.Path + "/config.yml"); err == nil {
		err = yaml.Unmarshal([]byte(source), &cntConfig)
		if err != nil {
			panic(err)
		}
	}

	log.Get().Debug("Home folder is " + cntConfig.Path)
}
