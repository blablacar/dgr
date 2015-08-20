package builder
import (
	"github.com/blablacar/cnt/log"
	"github.com/appc/spec/schema"
	"path/filepath"
	"io/ioutil"
	"github.com/ghodss/yaml"
	"github.com/blablacar/cnt/utils"
)

const POD_MANIFEST = "cnt-pod-manifest.yml"

type Pod struct {
	path     string
	args     BuildArgs
	target	 string
	manifest PodManifest
}

type PodManifest struct {
	NameAndVersion string                      `json:"name"`
	Pod            *schema.PodManifest          `json:"pod"`
}

func OpenPod(path string, args BuildArgs) (*Pod, error) {
	pod := new(Pod)

	if fullPath, err := filepath.Abs(path); err != nil {
		log.Get().Panic("Cannot get fullpath of project", err)
	} else {
		pod.path = fullPath
	}
	pod.args = args
	pod.target = pod.path + "/target"
	pod.manifest.Pod = utils.BasicPodManifest()
	pod.readManifest(pod.path + "/"+ POD_MANIFEST)
	return pod, nil
}

func (p *Pod) readManifest(manifestPath string) {
	source, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		log.Get().Panic(err)
	}
	err = yaml.Unmarshal([]byte(source), &p.manifest)
	if err != nil {
		log.Get().Panic(err)
	}
	log.Get().Trace("Pod manifest : ", p.manifest.NameAndVersion, p.manifest)
}