package utils
import (
	"github.com/appc/spec/schema"
	"io/ioutil"
	"log"
	"github.com/blablacar/cnt/spec"
    "github.com/appc/spec/schema/types"
)

const IMAGE_MANIFEST = `{
    "acKind": "ImageManifest",
    "acVersion": "0.6.1",
    "name": "xxx/xxx",
    "labels": [
        {
            "name": "version",
            "value": "0.0.0"
        },
        {
            "name": "os",
            "value": "linux"
        },
        {
            "name": "arch",
            "value": "amd64"
        }
    ],
    "app": {
        "exec": [
            "/bin/bash"
        ],
        "user": "0",
        "group": "0",
        "environment": [
            {
                "name": "PATH",
                "value": "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
            }
        ]
    }
}
`

func BasicImageManifest() *schema.ImageManifest {
	im := new(schema.ImageManifest)
	im.UnmarshalJSON([]byte(IMAGE_MANIFEST))
	return im
}

//
//func ReadManifest(path string) *schema.ImageManifest {
//    im := new(schema.ImageManifest)
//    content, err := ioutil.ReadFile(path)
//    if  err != nil {
//        panic(err)
////        config.GetConfig().Log.Panic("Cannot read manifest file", err)
//    }
//    im.UnmarshalJSON(content)
//    return im
//}

func WriteImageManifest(m *spec.AciManifest, targetFile string, projectName string, version string) {
    name, _ := types.NewACIdentifier(m.NameAndVersion.Name())

    labels := types.Labels{}
    labels = append(labels, types.Label{Name: "version", Value: m.NameAndVersion.Version()})

    if m.Aci.App.User == "" {
        m.Aci.App.User = "0"
    }
    if m.Aci.App.Group == "" {
        m.Aci.App.Group = "0"
    }

    im := schema.BlankImageManifest()
	im.Annotations = m.Aci.Annotations
    im.Dependencies = m.Aci.Dependencies
    im.Name = *name
    im.Labels = labels
    im.App = &types.App{
        Exec: m.Aci.App.Exec,
        EventHandlers: m.Aci.App.EventHandlers,
        User: m.Aci.App.User,
        Group: m.Aci.App.Group,
        WorkingDirectory: m.Aci.App.WorkingDirectory,
        Environment: m.Aci.App.Environment,
        MountPoints: m.Aci.App.MountPoints,
        Ports: m.Aci.App.Ports,
        Isolators: m.Aci.App.Isolators,
    }

	buff, err := im.MarshalJSON()
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(targetFile, buff, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
