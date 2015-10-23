package builder

import (
	log "github.com/Sirupsen/logrus"
	"github.com/blablacar/cnt/utils"
	"io/ioutil"
	"os"
	"path/filepath"
)

func (cnt *Img) Init() {
	initPath := cnt.path
	if cnt.args.Path != "" {
		initPath = cnt.args.Path
	}
	log.Info("Setting up files three")
	uid := "0"
	gid := "0"

	if os.Getenv("SUDO_UID") != "" {
		uid = os.Getenv("SUDO_UID")
		gid = os.Getenv("SUDO_GID")
	}

	files := make(map[string]string)

	files[PATH_RUNLEVELS+PATH_PRESTART_EARLY+"/10.prestart-early.sh"] = `#!/cnt/bin/busybox sh
echo "I'm a prestart early script that is run before templating"
`
	files[PATH_RUNLEVELS+PATH_PRESTART_LATE+"/10.prestart-late.sh"] = `#!/cnt/bin/busybox sh
echo "I'm a prestart late script that is run after templating"
`
	files[PATH_RUNLEVELS+PATH_BUILD+"/10.install.sh"] = `#!/cnt/bin/busybox sh
echo "I'm a build script that is run to install applications"
`
	files[PATH_RUNLEVELS+PATH_BUILD_SETUP+"/10.setup.sh"] = `#!/bin/sh
echo "I'm build setup script file that is run from $BASEDIR to prepare $TARGET/rootfs before running build scripts"
`
	files[PATH_RUNLEVELS+PATH_BUILD_LATE+"/10.setup.sh"] = `#!/cnt/bin/busybox sh
echo "I'm a build late script that is run to install applications after the copy of files,template,etc..."
`
	files[PATH_RUNLEVELS+PATH_INHERIT_BUILD_EARLY+"/10.inherit-build-early.sh"] = `#!/cnt/bin/busybox sh
echo "I'm a inherit build early script that is run on this image and all images that have me as From during build"
`
	files[PATH_RUNLEVELS+PATH_INHERIT_BUILD_LATE+"/10.inherit-build-early.sh"] = `#!/cnt/bin/busybox sh
echo "I'm a inherit build late script that is run on this image and all images that have me as From during build"
`
	files[PATH_FILES+"/dummy"] = `Dummy file
`
	files[PATH_ATTRIBUTES+"/attributes.yml"] = `default:
  dummy: world
`
	files[PATH_CONFD+PATH_CONFDOTD+"/templated.toml"] = `[template]
src = "templated.tmpl"
dest = "/templated"
uid = 0
gid = 0
mode = "0644"
keys = ["/"]
`
	files[PATH_CONFD+PATH_TEMPLATES+"/templated.tmpl"] = `{{$data := json (getv "/data")}}Hello {{ $data.dummy }}
`
	files[".gitignore"] = `target/
`
	files["cnt-manifest.yml"] = `from: ""
name: aci.example.com/aci-dummy:1
aci:
  app:
    exec: [ "/cnt/bin/busybox", "sh" ]
    eventHandlers:
      - { name: pre-start, exec: [ "/cnt/bin/prestart" ] }
`
	files[PATH_TESTS+"/dummy.bats"] = `#!/cnt/bin/bats -x

@test "Prestart should template" {
  result="$(cat /templated)"
  [ "$result" == "Hello world" ]
}

@test "Cnt should copy files" {
  result="$(cat /dummy)"
  [ "$result" == "Dummy file" ]
}
`

	files[PATH_TESTS+"/wait.sh"] = `exit 0`

	for filePath, data := range files {
		fpath := initPath + "/" + filePath
		os.MkdirAll(filepath.Dir(fpath), 0777)
		ioutil.WriteFile(fpath, []byte(data), 0777)
	}
	utils.ExecCmd("chown", "-R", uid+":"+gid, initPath)
}
