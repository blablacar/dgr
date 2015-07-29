package builder
import (
	"github.com/blablacar/cnt/log"
	"github.com/appc/spec/schema"
	"github.com/blablacar/cnt/utils"
	"io/ioutil"
	"path/filepath"
)

type Pod struct {
	path string
	args BuildArgs
	manifest *schema.PodManifest
}

func OpenPod(path string, args BuildArgs) (*Pod, error) {
	pod := new(Pod)

	if fullPath, err := filepath.Abs(path); err != nil {
		log.Get().Panic("Cannot get fullpath of project", err)
	} else {
		pod.path = fullPath
	}
	pod.args = args
	pod.manifest = utils.ReadPodManifest(pod.path + "/pod-manifest.json")
	return pod, nil
}

func (p *Pod) Build() {
	log.Get().Info("Building POD ")

//	runner := runner.ChrootRunner{}

	files, _ := ioutil.ReadDir(p.path)
	for _, f := range files {
		if f.IsDir() {
			if cnt, err := OpenCnt(p.path + "/" + f.Name(), p.args); err == nil {
				cnt.Build()
			}
		}
	}
}

func (p *Pod) Install() {
	log.Get().Info("Installing POD ")

	files, _ := ioutil.ReadDir(p.path)
	for _, f := range files {
		if f.IsDir() {
			if cnt, err := OpenCnt(p.path + "/" + f.Name(), p.args); err == nil {
				cnt.Install()
			}
		}
	}
}

func (p *Pod) Push() {
	log.Get().Info("Push POD ")

	files, _ := ioutil.ReadDir(p.path)
	for _, f := range files {
		if f.IsDir() {
			if cnt, err := OpenCnt(p.path + "/" + f.Name(), p.args); err == nil {
				cnt.Push()
			}
		}
	}
}

func (p *Pod) Clean() {
	log.Get().Info("Cleaning POD ")

	files, _ := ioutil.ReadDir(p.path)
	for _, f := range files {
		if f.IsDir() {
			if cnt, err := OpenCnt(p.path + "/" + f.Name(), p.args); err == nil {
				cnt.Clean()
			}
		}
	}
}


func (p *Pod) Test() {
	log.Get().Info("Testing POD ")

	files, _ := ioutil.ReadDir(p.path)
	for _, f := range files {
		if f.IsDir() {
			if cnt, err := OpenCnt(p.path + "/" + f.Name(), p.args); err == nil {
				cnt.Test()
			}
		}
	}
}