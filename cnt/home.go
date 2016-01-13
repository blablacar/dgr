package cnt

import (
	"github.com/blablacar/cnt/utils"
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
			panic(err)
		}
	}

	return HomeStruct{
		path:   path,
		Config: config,
	}
}

func DefaultHomeFolder() string {
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

	path := "/root/.config/cnt"
	user := os.Getenv("SUDO_USER")
	if user != "" {
		home, err := utils.ExecCmdGetOutput("bash", "-c", "echo ~"+user)
		if err != nil {
			panic("Cannot find user home" + err.Error())
		}
		path = home + "/.config/cnt"
	}
	return path
}
