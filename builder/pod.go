package builder
import (
	"github.com/blablacar/cnt/log"
	"github.com/appc/spec/schema"
	"github.com/blablacar/cnt/utils"
	"io/ioutil"
	"github.com/blablacar/cnt/runner"
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
	log.Get().Info("Building pod ")

	runner := runner.ChrootRunner{}

	files, _ := ioutil.ReadDir(p.path)
	for _, f := range files {
		if f.IsDir() {
			if cnt, err := OpenCnt(p.path + "/" + f.Name(), p.args); err == nil {
				cnt.Build(&runner)
			}
		}
	}
}

func (p *Pod) Push() {
	log.Get().Info("Push pod ")

	files, _ := ioutil.ReadDir(p.path)
	for _, f := range files {
		if f.IsDir() {
			if cnt, err := OpenCnt(p.path + "/" + f.Name(), p.args); err == nil {
				cnt.Push()
			}
		}
	}

}
