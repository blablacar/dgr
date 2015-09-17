package builder

import (
	"errors"
	"github.com/appc/spec/schema"
	"github.com/appc/spec/schema/types"
	"github.com/blablacar/cnt/log"
	"github.com/blablacar/cnt/spec"
	"github.com/blablacar/cnt/utils"
	"os"
	"strconv"
	"text/template"
)

func (p *Pod) Build() {
	log.Get().Info("Building POD : ", p.manifest.Name)

	os.RemoveAll(p.target)
	os.MkdirAll(p.target, 0777)

	apps := p.processAci()

	p.writeSystemdUnit(apps)
}

func (p *Pod) processAci() []schema.RuntimeApp {
	apps := []schema.RuntimeApp{}
	for _, e := range p.manifest.Pod.Apps {

		aciName := p.buildAciIfNeeded(e)
		// TODO: support not FS override by only storing info pod manifest
		//		if aciName == nil {
		//			aciName = &e.Image
		//		}

		name, _ := types.NewACName(e.Name)

		sum, err := utils.Sha512sum(p.path + "/" + e.Name + "/target/image.aci")
		if err != nil {
			log.Get().Panic(err)
		}

		tmp, _ := types.NewHash("sha512-" + sum)

		labels := types.Labels{}
		labels = append(labels, types.Label{Name: "version", Value: aciName.Version()})
		identifier, _ := types.NewACIdentifier(aciName.Name())
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
				EventHandlers:    e.App.EventHandlers,
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

func (p *Pod) buildAciIfNeeded(e spec.RuntimeApp) *spec.ACFullname {
	if dir, err := os.Stat(p.path + "/" + e.Name); err == nil && dir.IsDir() {
		aci, err := NewAciWithManifest(p.path+"/"+e.Name, p.args, p.toAciManifest(e))
		if err != nil {
			log.Get().Panic(err)
		}
		aci.PodName = &p.manifest.Name
		aci.Build()
		return &aci.manifest.NameAndVersion
	}
	return nil
}

//func (p *Pod) writePodManifest(apps []schema.RuntimeApp) {
//	m := p.manifest.Pod
//	ver, _ := types.NewSemVer("0.6.1")
//	manifest := schema.PodManifest{
//		ACKind:      "PodManifest",
//		ACVersion:   *ver,
//		Apps:        apps,
//		Volumes:     m.Volumes,
//		Isolators:   m.Isolators,
//		Annotations: m.Annotations,
//		Ports:       m.Ports}
//	utils.WritePodManifest(&manifest, p.target+POD_TARGET_MANIFEST)
//}

const SYSTEMD_TEMPLATE = `[Unit]
Description={{ .Shortname }} %i

[Service]
ExecStartPre=/opt/bin/rkt gc --grace-period=0s --expire-prepared=0s
ExecStart=/opt/bin/rkt --insecure-skip-verify run \
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

func (p *Pod) writeSystemdUnit(apps []schema.RuntimeApp) {
	volumes := []string{}
	envs := []string{}
	acilist := []string{}

	for _, env := range p.manifest.Envs {
		envs = append(envs, "--set-env="+env.Name+"='"+env.Value+"'")
	}
	for _, app := range apps {
		for _, mount := range app.App.MountPoints {
			mountPoint, err := p.getVolumeMountValue(mount.Name)
			if err != nil {
				log.Get().Panic(err)
			}
			volumes = append(volumes,
				"--volume="+mount.Name.String()+
					",kind="+mountPoint.Kind+
					",source="+mountPoint.Source+
					",readOnly="+strconv.FormatBool(*mountPoint.ReadOnly))
		}
		version, _ := app.Image.Labels.Get("version")
		acilist = append(acilist, app.Image.Name.String()+":"+version)
	}

	commands := []string{}
	if p.manifest.PrivateNet != "" {
		commands = append(commands, "--private-net="+p.manifest.PrivateNet)
	}

	commands = append(commands, volumes...)
	commands = append(commands, envs...)

	info := SystemdUnit{Shortname: p.manifest.Name.ShortName(), Commands: commands, Acilist: acilist}
	tmpl, _ := template.New("test").Parse(SYSTEMD_TEMPLATE)
	w, _ := os.Create(p.target + "/" + p.manifest.Name.ShortName() + "@.service")
	tmpl.Execute(w, info)
}

func (p *Pod) getVolumeMountValue(mountName types.ACName) (*types.Volume, error) {
	for _, volume := range p.manifest.Pod.Volumes {
		if volume.Name.Equals(mountName) {
			return &volume, nil
		}
	}
	return nil, errors.New("Volume mount point not set :" + mountName.String())
}
