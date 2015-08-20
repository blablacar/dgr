package builder
import (
	"github.com/blablacar/cnt/log"
	"io/ioutil"
)

func (p *Pod) Test() {
	log.Get().Info("Testing POD ")

	files, _ := ioutil.ReadDir(p.path)
	for _, f := range files {
		if f.IsDir() {
			if cnt, err := OpenAci(p.path + "/" + f.Name(), p.args); err == nil {
				cnt.Test()
			}
		}
	}
}