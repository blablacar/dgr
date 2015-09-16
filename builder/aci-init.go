package builder

import (
	"github.com/blablacar/cnt/log"
	"os"
	"strconv"
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

	uidInt, err := strconv.Atoi(uid)
	gidInt, err := strconv.Atoi(gid)

	if err != nil {
		log.Get().Panic(err)
	}
	folderList := []string{
		RUNLEVELS,
		RUNLEVELS_PRESTART,
		RUNLEVELS_LATESTART,
		RUNLEVELS_BUILD,
		RUNLEVELS_BUILD_SETUP,
		RUNLEVELS_BUILD_INHERIT_EARLY,
		RUNLEVELS_BUILD_INHERIT_LATE,
		CONFD,
		CONFD_TEMPLATE,
		CONFD_CONFIG,
		ATTRIBUTES,
		FILES_PATH,
	}
	for _, folder := range folderList {
		fpath := initPath + "/" + folder
		os.MkdirAll(fpath, 0777)
		os.Lchown(fpath, uidInt, gidInt)
	}
}
