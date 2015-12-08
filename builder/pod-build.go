package builder

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/appc/spec/schema"
	"github.com/appc/spec/schema/types"
	"github.com/blablacar/cnt/spec"
	"github.com/blablacar/cnt/utils"
	"os"
)

const PATH_POD_MANIFEST = "/pod-manifest.json"

func (p *Pod) Build() {
	log.Info("Building POD : ", p.manifest.Name)

	os.RemoveAll(p.target)
	os.MkdirAll(p.target, 0777)

	p.preparePodVersion()
	apps := p.processAci()

	p.writePodManifest(apps)
}

func (p *Pod) preparePodVersion() {
	if p.manifest.Name.Version() == "" {
		p.manifest.Name = *spec.NewACFullName(p.manifest.Name.Name() + ":" + utils.GenerateVersion())
	}
}

func (p *Pod) processAci() []schema.RuntimeApp {
	apps := []schema.RuntimeApp{}
	for _, e := range p.manifest.Pod.Apps {

		aci := p.buildAci(e)

		name, _ := types.NewACName(e.Name)

		sum, err := utils.Sha512sum(aci.target + "/image.aci")
		if err != nil {
			panic(err)
		}

		tmp, _ := types.NewHash("sha512-" + sum)

		labels := types.Labels{}
		labels = append(labels, types.Label{Name: "version", Value: aci.manifest.NameAndVersion.Version()})
		identifier, _ := types.NewACIdentifier(aci.manifest.NameAndVersion.Name())
		ttmp := schema.RuntimeImage{Name: identifier, ID: *tmp, Labels: labels}

		if e.App.User == "" {
			e.App.User = "0"
		}
		if e.App.Group == "" {
			e.App.Group = "0"
		}

		apps = append(apps, schema.RuntimeApp{
			Name:  *name,
			Image: ttmp,
			App: &types.App{
				Exec:             e.App.Exec,
				User:             e.App.User,
				Group:            e.App.Group,
				WorkingDirectory: e.App.WorkingDirectory,
				Environment:      e.App.Environment,
				MountPoints:      e.App.MountPoints,
				Ports:            e.App.Ports,
				Isolators:        e.App.Isolators,
			},
			Mounts:      e.Mounts,
			Annotations: e.Annotations})
	}

	return apps
}

func (p *Pod) buildAci(e spec.RuntimeApp) *Aci {
	if dir, err := os.Stat(p.path + "/" + e.Name); err == nil && dir.IsDir() {
		aci, err := NewAciWithManifest(p.path+"/"+e.Name, p.args, p.toAciManifest(e), nil)
		if err != nil {
			panic(err)
		}
		aci.podName = &p.manifest.Name
		aci.Build()
		return aci
	}
	panic("Cannot found Pod's aci directory :" + p.path + "/" + e.Name)
	return nil
}

func (p *Pod) writePodManifest(apps []schema.RuntimeApp) {
	m := p.manifest.Pod
	ver, _ := types.NewSemVer("0.6.1")
	manifest := schema.PodManifest{
		ACKind:      "PodManifest",
		ACVersion:   *ver,
		Apps:        apps,
		Volumes:     m.Volumes,
		Isolators:   m.Isolators,
		Annotations: m.Annotations,
		Ports:       m.Ports}
	utils.WritePodManifest(&manifest, p.target+PATH_POD_MANIFEST)
}

const SYSTEMD_TEMPLATE = `[Unit]
Description={{ .Shortname }} %i

[Service]
ExecStartPre=/opt/bin/rkt gc --grace-period=0s --expire-prepared=0s
ExecStart=/opt/bin/rkt --insecure-options=image run \
{{range $i, $e := .Commands}}  {{$e}} \
{{end}}{{range $i, $e := .Acilist}}{{if $i}} \
{{end}}  {{$e}}{{end}}

[Install]
WantedBy=multi-user.target
`

type SystemdUnit struct {
	Shortname string
	Commands  []string
	Acilist   []string
}

func (p *Pod) getVolumeMountValue(mountName types.ACName) (*types.Volume, error) {
	for _, volume := range p.manifest.Pod.Volumes {
		if volume.Name.Equals(mountName) {
			return &volume, nil
		}
	}
	return nil, errors.New("Volume mount point not set :" + mountName.String())
}
