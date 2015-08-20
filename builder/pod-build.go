package builder
import (
	"github.com/blablacar/cnt/log"
	"os"
	"github.com/blablacar/cnt/utils"
	"github.com/appc/spec/schema/types"
	"github.com/appc/spec/schema"
)


func (p *Pod) Build() {
	log.Get().Info("Building POD : ", p.manifest.NameAndVersion)

	os.MkdirAll(p.target, 0777)

	//	p.manifest.Pod.Apps.Get("yop").Image.ID.Set("sha512-3cf428b611c03a08d5732a1dd576fd5de257022023a21fd440ad42fc7ee006cd")
	//	p.manifest.Pod.Apps.Get("").Image.ID.Set("sha512-3cf428b611c03a08d5732a1dd576fd5de257022023a21fd440ad42fc7ee006cd")
	//	p.manifest.Pod.Apps.Get("").Name.Set("yopla")
//	for _, element := range p.manifest.Pod.Apps {
//		if err := element.Name.Set("yopla"); err != nil {
//			log.Get().Panic(err)
//		}
//		element.Image.ID.Set("sha512-3cf428b611c03a08d5732a1dd576fd5de257022023a21fd440ad42fc7ee006cd")
//	}

	if err := p.manifest.Pod.Apps.Get("yoplaboom").Image.ID.Set("sha512-3cf428b611c03a08d5732a1dd576fd5de257022023a21fd440ad42fc7ee006cd"); err != nil {
		log.Get().Panic(err)
	}

	tmp, _ := types.NewHash("sha512-3cf428b611c03a08d5732a1dd576fd5de257022023a21fd440ad42fc7ee006cd")
//	tmp.Set("sha512-aaf428b611c03a08d5732a1dd576fd5de257022023a21fd440ad42fc7ee006cd")
//	log.Get().Panic(tmp)


	p.manifest.Pod.Apps.Get("yoplaboom").Image.ID.Set("sha512-aaf428b611c03a08d5732a1dd576fd5de257022023a21fd440ad42fc7ee006cd")
	p.manifest.Pod.Apps.Get("yoplaboom").Image.ID = *tmp
	log.Get().Warn(">>", p.manifest.Pod.Apps.Get("yoplaboom").Image.ID.Empty())
	log.Get().Warn(">>", p.manifest.Pod.Apps.Get("yoplaboom").Image.ID.String())

	old := p.manifest.Pod.Apps.Get("yoplaboom").Image

	ttmp := schema.RuntimeImage{Name: old.Name, ID: *tmp, Labels: old.Labels}
//	p.manifest.Pod.Apps.Get("yoplaboom").Image

	p.manifest.Pod.Apps = []schema.RuntimeApp{}
	name, _ := types.NewACName("ouda")
	p.manifest.Pod.Apps = append(p.manifest.Pod.Apps, schema.RuntimeApp{Name: *name, Image: ttmp})
//	*new(schema.AppList)
//	p.manifest.Pod.Apps.

	log.Get().Warn(p.manifest.Pod.Apps.Get("ouda"))

	p.writePodManifest()
}

func (p *Pod) writePodManifest() {
	utils.WritePodManifest(p.manifest.Pod, p.target + "/manifest")
}
