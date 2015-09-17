package builder

import (
	"github.com/blablacar/cnt/log"
	"os"
	"strconv"
)

const INIT_BUILD_FILE = `#!/bin/bash
echo "I'm a build script that is run to install applications"
`

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

	uidInt, err := strconv.Atoi(uid)
	gidInt, err := strconv.Atoi(gid)

	if err != nil {
		log.Get().Panic(err)
	}
	folderList := []string{
		PATH_RUNLEVELS,
		PATH_RUNLEVELS+PATH_PRESTART_EARLY,
		PATH_RUNLEVELS+PATH_PRESTART_LATE,
		PATH_RUNLEVELS+PATH_BUILD,
		PATH_RUNLEVELS+PATH_BUILD_LATE,
		PATH_RUNLEVELS+PATH_BUILD_SETUP,
		PATH_RUNLEVELS+PATH_INHERIT_BUILD_EARLY,
		PATH_RUNLEVELS+PATH_INHERIT_BUILD_LATE,
		PATH_CONFD,
		PATH_CONFD+PATH_TEMPLATES,
		PATH_CONFD+PATH_CONFDOTD,
		PATH_ATTRIBUTES,
		PATH_FILES,
	}
	for _, folder := range folderList {
		fpath := initPath + "/" + folder
		os.MkdirAll(fpath, 0777)
		os.Lchown(fpath, uidInt, gidInt)
	}
}
