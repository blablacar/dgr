package builder

import (
	"github.com/blablacar/cnt/log"
	"path/filepath"
	"os"
	"github.com/blablacar/cnt/utils"
	"io/ioutil"
)

func (cnt *Img) Init() {
	initPath := cnt.path
	if cnt.args.Path != "" {
		initPath = cnt.args.Path
	}
	log.Get().Info("Setting up files three")
	uid := "0"
	gid := "0"

	if os.Getenv("SUDO_UID") != "" {
		uid = os.Getenv("SUDO_UID")
		gid = os.Getenv("SUDO_GID")
	}


	files := make(map[string]string)

	files[PATH_RUNLEVELS+PATH_PRESTART_EARLY+"/10.prestart-early.sh"] = `#!/bin/bash
echo "I'm a prestart early script that is run before templating"
`
	files[PATH_RUNLEVELS+PATH_PRESTART_LATE+"/10.prestart-late.sh"] = `#!/bin/bash
echo "I'm a prestart late script that is run after templating"
`
	files[PATH_RUNLEVELS+PATH_BUILD+"/10.install.sh"] = `#!/bin/bash
echo "I'm a build script that is run to install applications"
`
	files[PATH_RUNLEVELS+PATH_BUILD_SETUP+"/10.setup.sh"] = `#!/bin/bash -x
echo "I'm build setup script file that is run to prepare $TARGET/rootfs before running build scripts"
mkdir -p $TARGET/rootfs/{bin,lib64,sys,usr,usr/lib,usr/bin/,cnt,cnt/bin}
curl 'https://github.com/kelseyhightower/confd/releases/download/v0.10.0/confd-0.10.0-linux-amd64' -O $TARGET/rootfs/cnt/bin/confd
curl 'https://github.com/blablacar/attribute-merger/releases/download/0.1/attributes-merger' -O $TARGET/rootfs/cnt/bin/attributes-merger
cp /bin/dirname $TARGET/rootfs/bin/
cp /bin/bash $TARGET/rootfs/bin/
cp --preserve=links /usr/lib/libc.so.* $TARGET/rootfs/usr/lib
cp --preserve=links /usr/lib/libreadline.so.* $TARGET/rootfs/usr/lib
cp --preserve=links /usr/lib/libncursesw.so.* $TARGET/rootfs/usr/lib
cp --preserve=links /usr/lib/libdl.so.* $TARGET/rootfs/usr/lib
cp --preserve=links /usr/lib/libc.so.* $TARGET/rootfs/usr/lib
cp --preserve=links /lib64/ld-linux-x86-64.so.* $TARGET/rootfs/lib64
`
	files[PATH_RUNLEVELS+PATH_BUILD_LATE+"/10.setup.sh"] = `#!/bin/bash
echo "I'm a build late script that is run to install applications after the copy of files,template,etc..."
`
	files[PATH_RUNLEVELS+PATH_INHERIT_BUILD_EARLY+"/10.inherit-build-early.sh"] = `#!/bin/bash
echo "I'm a inherit build early script that is run on this image and all images that have me as From during build"
`
	files[PATH_RUNLEVELS+PATH_INHERIT_BUILD_LATE+"/10.inherit-build-early.sh"] = `#!/bin/bash
echo "I'm a inherit build late script that is run on this image and all images that have me as From during build"
`
	files[PATH_FILES+"/dummy"] = `Dummy file
`
	files[PATH_ATTRIBUTES+"/attributes.yml"] = `default:
  dummy: world
`
	files[PATH_CONFD+PATH_CONFDOTD+"/dummy.toml"] = `[template]
src = "dummy.tmpl"
dest = "/dummy"
uid = 0
gid = 0
mode = "0644"
keys = ["/"]
`
	files[PATH_CONFD+PATH_TEMPLATES+"/dummy.tmpl"] = `{{$data := json (getv "/data")}}Hello {{ $data.dummy }}
`
	files[".gitignore"] = `target/
`
	files["cnt-manifest.yml"] = `from: ""
name: aci.example.com/aci-dummy:1
aci:
  app:
    exec: [ "/bin/bash" ]
    eventHandlers:
      - { name: pre-start, exec: [ "/cnt/bin/prestart" ] }
`
	files[PATH_FILES+PATH_CNT+"/bin/prestart"] = `#!/bin/bash

BASEDIR=${0%/*}
CNT_PATH=/cnt

execute_files() {
  fdir=$1
  [ -d "$fdir" ] || return 0

  find "$fdir" -mindepth 1 -maxdepth 1 -type f -print0 | sort -z -n |
  while read -r -d $'\0' file; do
      echo "$file"
      [ -x "$file" ] && "$file"
  done
}

execute_files ${CNT_PATH}/prestart-early

${BASEDIR}/attributes-merger -i ${CNT_PATH}/attributes -e CONFD_OVERRIDE
export CONFD_DATA=$(cat attributes.json)
${BASEDIR}/confd -onetime -config-file=${CNT_PATH}/prestart/confd.toml

execute_files ${CNT_PATH}/prestart-late
`
	files[PATH_FILES+PATH_CNT+"/prestart/confd.toml"] = `backend = "env"
confdir = "/cnt"
prefix = "/confd"
log-level = "debug"
`

	for filePath, data := range files {
		fpath := initPath+"/"+filePath
		os.MkdirAll(filepath.Dir(fpath), 0777)
		ioutil.WriteFile(fpath, []byte(data), 0777)
	}
	utils.ExecCmd("chown", "-R", uid+":"+gid, initPath)
}
