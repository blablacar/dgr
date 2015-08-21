package builder
import (
	"github.com/blablacar/cnt/log"
	"os"
	"github.com/blablacar/cnt/utils"
	"github.com/appc/spec/schema/types"
	"github.com/appc/spec/schema"
	"github.com/blablacar/cnt/spec"
)


func (p *Pod) Build() {
	log.Get().Info("Building POD : ", p.manifest.NameAndVersion)

	os.MkdirAll(p.target, 0777)
	os.Remove(p.target + POD_TARGET_MANIFEST)

	apps := p.processAci()

	p.writePodManifest(apps)
}

func (p *Pod) processAci() []schema.RuntimeApp {
	apps := []schema.RuntimeApp{}
	for _, e := range p.manifest.Pod.Apps {

		p.buildAciIfNeeded(e)

		name, _ := types.NewACName(e.Image.ShortName())

		sum, err := utils.Sha512sum(p.path + "/" + e.Name + "/target/image.aci")
		if (err != nil) {
			log.Get().Panic(err)
		}

		tmp, _ := types.NewHash("sha512-" + sum)

		labels := types.Labels{}
		labels = append(labels, types.Label{Name: "version", Value: e.Image.Version()})
		identifier, _ := types.NewACIdentifier(e.Image.Name())
		ttmp := schema.RuntimeImage{Name: identifier, ID: *tmp, Labels: labels}
		apps = append(apps, schema.RuntimeApp{Name: *name, Image: ttmp, App: e.App, Mounts: e.Mounts, Annotations: e.Annotations})

	}
	return apps
}

func (p *Pod) buildAciIfNeeded(e spec.RuntimeApp) bool {
	if dir, err := os.Stat(p.path + "/" + e.Name); err == nil && dir.IsDir() {
		aci, err := NewAciWithManifest(p.path + "/" + e.Name, p.args, p.toAciManifest(e))
		if (err != nil) {
			log.Get().Panic(err)
		}
		aci.Build()
		return true
	}
	return false
}

func (p *Pod) writePodManifest(apps []schema.RuntimeApp) {

	m := p.manifest.Pod
	ver, _ := types.NewSemVer("0.6.1")
	manifest := schema.PodManifest{
		ACKind: "PodManifest",
		ACVersion: *ver,
		Apps: apps,
		Volumes: m.Volumes,
		Isolators: m.Isolators,
		Annotations: m.Annotations,
		Ports: m.Ports}
	utils.WritePodManifest(&manifest, p.target + POD_TARGET_MANIFEST)
}
