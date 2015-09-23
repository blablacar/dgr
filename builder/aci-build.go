package builder

import (
	"github.com/blablacar/cnt/dist"
	"github.com/blablacar/cnt/log"
	"github.com/blablacar/cnt/utils"
	"io/ioutil"
	"os"
	"strings"
)

func (cnt *Img) Build() error {
	log.Get().Info("Building Image : ", cnt.manifest.NameAndVersion)

	os.MkdirAll(cnt.rootfs, 0777)

	cnt.processFrom()
	cnt.copyInternals()
	cnt.copyRunlevelsScripts()

	cnt.runLevelBuildSetup()

	cnt.writeImgManifest()
	cnt.writeCntManifest() // TODO move that, here because we update the version number to generated version

	cnt.runBuild()
	cnt.copyAttributes()
	cnt.copyConfd()
	cnt.copyFiles()
	cnt.runBuildLate()

	cnt.tarAci()
	//	ExecCmd("chown " + os.Getenv("SUDO_USER") + ": " + target + "/*") //TODO chown
	return nil
}

func (i *Img) CheckBuilt() {
	if _, err := os.Stat(i.target + PATH_IMAGE_ACI); os.IsNotExist(err) {
		if err := i.Build(); err != nil {
			log.Get().Panic("Cannot continue since build failed")
		}
	}
}

///////////////////////////////////////////////////////

func (cnt *Img) writeCntManifest() {
	utils.CopyFile(cnt.path+PATH_CNT_MANIFEST, cnt.target+PATH_CNT_MANIFEST)
}

func (cnt *Img) runBuildLate() {
	res, err := utils.IsDirEmpty(cnt.target + PATH_RUNLEVELS + PATH_BUILD_LATE)
	res2, err2 := utils.IsDirEmpty(cnt.rootfs + PATH_CNT + PATH_RUNLEVELS + PATH_INHERIT_BUILD_LATE)
	if (res && res2) || (err != nil && err2 != nil) {
		return
	}

	{
		rootfs := "${TARGET}/rootfs"
		if cnt.manifest.Build.NoBuildImage() {
			rootfs = ""
		}
		build := strings.Replace(BUILD_SCRIPT_LATE, "%%ROOTFS%%", rootfs, 1)
		ioutil.WriteFile(cnt.target+"/build-late.sh", []byte(build), 0777)
	}

	if err := utils.ExecCmd("systemd-nspawn", "--version"); err == nil {
		log.Get().Info("Run with systemd-nspawn")
		if err := utils.ExecCmd("systemd-nspawn", "--directory="+cnt.rootfs, "--capability=all",
			"--bind="+cnt.target+"/:/target", "--share-system", "target/build-late.sh"); err != nil {
			log.Get().Panic("Build step did not succeed", err)
		}
	} else {
		log.Get().Panic("systemd-nspawn is required")
	}
}

func (cnt *Img) runBuild() {
	if res, err := utils.IsDirEmpty(cnt.target + PATH_RUNLEVELS + PATH_BUILD); res || err != nil {
		return
	}
	if err := utils.ExecCmd("systemd-nspawn", "--version"); err != nil {
		log.Get().Panic("systemd-nspawn is required")
	}

	rootfs := "${TARGET}/rootfs"
	if cnt.manifest.Build.NoBuildImage() {
		rootfs = ""
	}
	build := strings.Replace(BUILD_SCRIPT, "%%ROOTFS%%", rootfs, 1)
	ioutil.WriteFile(cnt.target+"/build.sh", []byte(build), 0777)

	if err := utils.ExecCmd("systemd-nspawn", "--directory="+cnt.rootfs, "--capability=all",
		"--bind="+cnt.target+"/:/target", "--share-system", "target/build.sh"); err != nil {
		log.Get().Panic("Build step did not succeed", err)
	}
}

func (cnt *Img) processFrom() {
	if cnt.manifest.From == "" {
		return
	}
	if err := utils.ExecCmd("bash", "-c", "rkt image list --fields name --no-legend | grep "+cnt.manifest.From.String()); err != nil {
		utils.ExecCmd("rkt", "--insecure-skip-verify=true", "fetch", cnt.manifest.From.String())
	}
	if err := utils.ExecCmd("rkt", "image", "render", "--overwrite", cnt.manifest.From.String(), cnt.target); err != nil {
		log.Get().Panic("Cannot render from image"+cnt.manifest.From.String(), err)
	}
}

func (cnt *Img) copyInternals() {
	log.Get().Info("Copy internals")
	os.MkdirAll(cnt.rootfs+PATH_CNT+PATH_BIN, 0755)
	os.MkdirAll(cnt.rootfs+"/usr/bin", 0755) // this is required by systemd-nspawn

	busybox, _ := dist.Asset("dist/bindata/busybox")
	if err := ioutil.WriteFile(cnt.rootfs+PATH_CNT+PATH_BIN+"/busybox", busybox, 0777); err != nil {
		log.Get().Panic(err)
	}

	confd, _ := dist.Asset("dist/bindata/confd")
	if err := ioutil.WriteFile(cnt.rootfs+PATH_CNT+PATH_BIN+"/confd", confd, 0777); err != nil {
		log.Get().Panic(err)
	}

	attributeMerger, _ := dist.Asset("dist/bindata/attributes-merger")
	if err := ioutil.WriteFile(cnt.rootfs+PATH_CNT+PATH_BIN+"/attributes-merger", attributeMerger, 0777); err != nil {
		log.Get().Panic(err)
	}

	confdFile := `backend = "env"
confdir = "/cnt"
prefix = "/confd"
log-level = "debug"
`
	os.MkdirAll(cnt.rootfs+PATH_CNT+"/prestart", 0755)
	if err := ioutil.WriteFile(cnt.rootfs+PATH_CNT+"/prestart/confd.toml", []byte(confdFile), 0777); err != nil {
		log.Get().Panic(err)
	}

	if err := ioutil.WriteFile(cnt.rootfs+PATH_CNT+PATH_BIN+"/prestart", []byte(PRESTART), 0777); err != nil {
		log.Get().Panic(err)
	}
}

func (cnt *Img) copyRunlevelsScripts() {
	log.Get().Info("Copy Runlevels scripts")
	utils.CopyDir(cnt.path+PATH_RUNLEVELS+PATH_BUILD, cnt.target+PATH_RUNLEVELS+PATH_BUILD)
	utils.CopyDir(cnt.path+PATH_RUNLEVELS+PATH_BUILD_LATE, cnt.target+PATH_RUNLEVELS+PATH_BUILD_LATE)
	utils.CopyDir(cnt.path+PATH_RUNLEVELS+PATH_BUILD_SETUP, cnt.target+PATH_RUNLEVELS+PATH_BUILD_SETUP)
	utils.CopyDir(cnt.path+PATH_RUNLEVELS+PATH_PRESTART_EARLY, cnt.target+PATH_RUNLEVELS+PATH_PRESTART_EARLY)
	utils.CopyDir(cnt.path+PATH_RUNLEVELS+PATH_PRESTART_LATE, cnt.target+PATH_RUNLEVELS+PATH_PRESTART_LATE)

	utils.CopyDir(cnt.path+PATH_RUNLEVELS+PATH_PRESTART_EARLY, cnt.target+PATH_ROOTFS+PATH_CNT+PATH_RUNLEVELS+PATH_PRESTART_EARLY)
	utils.CopyDir(cnt.path+PATH_RUNLEVELS+PATH_PRESTART_LATE, cnt.target+PATH_ROOTFS+PATH_CNT+PATH_RUNLEVELS+PATH_PRESTART_LATE)
	utils.CopyDir(cnt.path+PATH_RUNLEVELS+PATH_INHERIT_BUILD_EARLY, cnt.target+PATH_ROOTFS+PATH_CNT+PATH_RUNLEVELS+PATH_INHERIT_BUILD_EARLY)
	utils.CopyDir(cnt.path+PATH_RUNLEVELS+PATH_INHERIT_BUILD_LATE, cnt.target+PATH_ROOTFS+PATH_CNT+PATH_RUNLEVELS+PATH_INHERIT_BUILD_LATE)
}

func (cnt *Img) runLevelBuildSetup() {
	files, err := ioutil.ReadDir(cnt.path + PATH_RUNLEVELS + PATH_BUILD_SETUP)
	if err != nil {
		return
	}

	os.Setenv("BASEDIR", cnt.path)
	os.Setenv("TARGET", cnt.target)
	for _, f := range files {
		if !f.IsDir() {
			log.Get().Info("Running Build setup level : ", f.Name())
			if err := utils.ExecCmd(cnt.path + PATH_RUNLEVELS + PATH_BUILD_SETUP + "/" + f.Name()); err != nil {
				log.Get().Panic(err)
			}
		}
	}
}

func (cnt *Img) tarAci() {
	dir, _ := os.Getwd()
	log.Get().Debug("chdir to", cnt.target)
	os.Chdir(cnt.target)
	utils.Tar(cnt.args.Zip, PATH_IMAGE_ACI[1:], PATH_MANIFEST[1:], PATH_ROOTFS[1:])
	log.Get().Debug("chdir to", dir)
	os.Chdir(dir)
}

func (cnt *Img) copyConfd() {
	utils.CopyDir(cnt.path+PATH_CONFD+PATH_CONFDOTD, cnt.rootfs+PATH_CNT+PATH_CONFDOTD)
	utils.CopyDir(cnt.path+PATH_CONFD+PATH_TEMPLATES, cnt.rootfs+PATH_CNT+PATH_TEMPLATES)
}

func (cnt *Img) copyFiles() {
	utils.CopyDir(cnt.path+PATH_FILES, cnt.rootfs)
}

func (cnt *Img) copyAttributes() {
	utils.CopyDir(cnt.path+PATH_ATTRIBUTES, cnt.rootfs+PATH_CNT+PATH_ATTRIBUTES+"/"+cnt.manifest.NameAndVersion.ShortName())
}

func (cnt *Img) writeImgManifest() {
	log.Get().Debug("Writing aci manifest")
	version := cnt.manifest.NameAndVersion.Version()
	if version == "" {
		version = utils.GenerateVersion()
	}
	utils.WriteImageManifest(&cnt.manifest, cnt.target+PATH_MANIFEST, cnt.manifest.NameAndVersion.Name(), version)
}
