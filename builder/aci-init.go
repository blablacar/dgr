package builder
import (
	"github.com/blablacar/cnt/log"
	"os"
	"strconv"
)

func (cnt *Cnt) Init() {
	log.Get().Info("Setting up files three")
	uid := os.Getenv("SUDO_UID")
	uidInt,err := strconv.Atoi(uid)
	gid := os.Getenv("SUDO_GID")
	gidInt ,err := strconv.Atoi(gid)

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
	for _,folder := range folderList  {
		fpath := cnt.path + "/" +folder
		os.MkdirAll(fpath,0777 )
		os.Lchown(fpath,uidInt,gidInt)
	}
}