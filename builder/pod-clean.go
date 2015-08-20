package builder
import (
	"github.com/blablacar/cnt/log"
	"os"
)

func (p *Pod) Clean() {
	log.Get().Info("Cleaning POD ", p.manifest.NameAndVersion)

	if err := os.RemoveAll(p.target + "/"); err != nil {
		log.Get().Panic("Cannot clean ", p.manifest.NameAndVersion, err)
	}

//	files, _ := ioutil.ReadDir(p.path)
//	for _, f := range files {
//		if f.IsDir() {
//			if cnt, err := OpenAci(p.path + "/" + f.Name(), p.args); err == nil {
//				cnt.Clean()
//			}
//		}
//	}
}