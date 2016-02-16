package dgr

import (
	"github.com/blablacar/dgr/utils"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/logs"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type Config struct {
	Path string
	Push struct {
		Type     string `yaml:"type,omitempty"`
		Url      string `yaml:"url,omitempty"`
		Username string `yaml:"username,omitempty"`
		Password string `yaml:"password,omitempty"`
	} `yaml:"push,omitempty"`
	TargetWorkDir string `yaml:"targetWorkDir,omitempty"`
}

type HomeStruct struct {
	path   string
	Config Config
}

func NewHome(path string) HomeStruct {
	logs.WithField("path", path).Debug("Loading home")

	var config Config
	if source, err := ioutil.ReadFile(path + "/config.yml"); err == nil {
		err = yaml.Unmarshal([]byte(source), &config)
		if err != nil {
			logs.WithEF(err, data.WithField("path", path+"/config.yml")).Fatal("Failed to process configuration file")
		}
	} else if source, err := ioutil.ReadFile(DefaultHomeFolder("cnt") + "/config.yml"); err == nil {
		logs.WithField("old", DefaultHomeFolder("cnt")+"/config.yml").WithField("new", DefaultHomeFolder("")).Warn("You are using old home folder")
		err = yaml.Unmarshal([]byte(source), &config)
		if err != nil {
			logs.WithEF(err, data.WithField("path", path+"/config.yml")).Fatal("Failed to process configuration file")
		}
	}

	return HomeStruct{
		path:   path,
		Config: config,
	}
}

func DefaultHomeFolder(programName string) string {
	if programName == "" {
		programName = "dgr"
	}
	//	switch runtime.GOOS {
	//	case "windows":
	//		dgrConfig.Path = utils.UserHomeOrFatal() + "/AppData/Local/dgr"
	//	case "darwin":
	//		dgrConfig.Path = utils.UserHomeOrFatal() + "/Library/dgr"
	//	case "linux":
	//		dgrConfig.Path = utils.UserHomeOrFatal() + "/.config/dgr"
	//	default:
	//		log.Get().Panic("Unsupported OS, please fill a bug repost")
	//	}

	path := "/root/.config/" + programName
	user := os.Getenv("SUDO_USER")
	if user != "" {
		home, err := utils.ExecCmdGetOutput("bash", "-c", "echo ~"+user)
		if err != nil {
			panic("Cannot find user home" + err.Error())
		}
		path = home + "/.config/" + programName
	}
	return path
}
