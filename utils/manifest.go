package utils
import (
	"github.com/appc/spec/schema"
	"io/ioutil"
    "strings"
    "log"
    "github.com/appc/spec/schema/types"
)

const(
	imageManifest = `{
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
)

//"eventHandlers": [
//{
//"exec": [
//"/usr/local/bin/prestart"
//],
//"name": "pre-start"
//}
//],


func BasicManifest() *schema.ImageManifest {
    im := new(schema.ImageManifest)
    im.UnmarshalJSON([]byte(imageManifest))
    return im
}

func ReadPodManifest(path string) *schema.PodManifest {
    im := new(schema.PodManifest)
    content, err := ioutil.ReadFile(path)
    if  err != nil {
        panic(err)
        //        config.GetConfig().Log.Panic("Cannot read manifest file", err)
    }
    im.UnmarshalJSON(content)
    return im
}

func ReadManifest(path string) *schema.ImageManifest {
    im := new(schema.ImageManifest)
    content, err := ioutil.ReadFile(path)
    if  err != nil {
        panic(err)
//        config.GetConfig().Log.Panic("Cannot read manifest file", err)
    }
    im.UnmarshalJSON(content)
    return im
}

func WriteImageManifest(im *schema.ImageManifest, targetFile string, projectName types.ACIdentifier, version string) {
	buff, err := im.MarshalJSON()
    res := strings.Replace(string(buff), "0.0.0", version, 1)
    res = strings.Replace(res, "__VERSION__", version, 1)
    res = strings.Replace(res, "xxx/xxx", string(projectName), 1)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(targetFile, []byte(res), 0644)
	if err != nil {
		log.Fatal(err)
	}
}
