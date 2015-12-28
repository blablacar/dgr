package builder

import (
	"github.com/Sirupsen/logrus"
	"github.com/blablacar/cnt/dist"
	"github.com/blablacar/cnt/utils"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func (aci *Aci) Build() error {
	aci.log.Info("Building")

	os.MkdirAll(aci.rootfs, 0777)

	aci.fullyResolveDependencies()
	aci.processFrom()
	aci.copyInternals()
	aci.copyRunlevelsScripts()

	aci.runLevelBuildSetup()

	aci.writeImgManifest()

	aci.runBuild()
	aci.copyAttributes()
	aci.copyConfd()
	aci.copyFiles()
	aci.runBuildLate()

	aci.tarAci(false)
	//	ExecCmd("chown " + os.Getenv("SUDO_USER") + ": " + target + "/*") //TODO chown
	return nil
}

func (aci *Aci) CheckBuilt() {
	if _, err := os.Stat(aci.target + PATH_IMAGE_ACI); os.IsNotExist(err) {
		if err := aci.Build(); err != nil {
			panic("Cannot continue since build failed")
		}
	}
}

///////////////////////////////////////////////////////

func (aci *Aci) fullyResolveDependencies() {
	if !aci.FullyResolveDep {
		return
	}

	for i, dep := range aci.manifest.Aci.Dependencies {
		resolved, err := dep.FullyResolved()
		if err != nil {
			aci.log.WithField("dependency", dep).WithError(err).Fatal("Cannot fully resolve dependency")
		}
		aci.manifest.Aci.Dependencies[i] = *resolved
	}
}

func (aci *Aci) runBuildLate() {
	res, err := utils.IsDirEmpty(aci.target + PATH_RUNLEVELS + PATH_BUILD_LATE)
	res2, err2 := utils.IsDirEmpty(aci.rootfs + PATH_CNT + PATH_RUNLEVELS + PATH_INHERIT_BUILD_LATE)
	if (res && res2) || (err != nil && err2 != nil) {
		return
	}

	{
		rootfs := "${TARGET}/rootfs"
		if aci.manifest.Build.NoBuildImage() {
			rootfs = ""
		}
		build := strings.Replace(BUILD_SCRIPT_LATE, "%%ROOTFS%%", rootfs, 1)
		ioutil.WriteFile(aci.target+"/build-late.sh", []byte(build), 0777)
	}

	checkSystemdNspawn()

	aci.log.Info("Starting systemd-nspawn to run Build late scripts")
	if err := utils.ExecCmd("systemd-nspawn", "--directory="+aci.rootfs, "--capability=all",
		"--bind="+aci.target+"/:/target", "target/build-late.sh"); err != nil {
		aci.log.WithError(err).Fatal("Build late part failed")
	}
}

func (aci *Aci) runBuild() {
	if res, err := utils.IsDirEmpty(aci.target + PATH_RUNLEVELS + PATH_BUILD); res || err != nil {
		return
	}

	checkSystemdNspawn()

	rootfs := "${TARGET}/rootfs"
	if aci.manifest.Build.NoBuildImage() {
		rootfs = ""
	}
	build := strings.Replace(BUILD_SCRIPT, "%%ROOTFS%%", rootfs, 1)
	ioutil.WriteFile(aci.target+"/build.sh", []byte(build), 0777)

	aci.log.Info("Starting systemd-nspawn to run Build scripts")
	if err := utils.ExecCmd("systemd-nspawn", "--directory="+aci.rootfs, "--capability=all",
		"--bind="+aci.target+"/:/target", "target/build.sh"); err != nil {
		aci.log.WithError(err).Fatal("Build part failed")
	}
}

func (aci *Aci) processFrom() {
	if aci.manifest.From == "" {
		return
	}
	if err := utils.ExecCmd("bash", "-c", "rkt image list --fields name --no-legend | grep -q "+aci.manifest.From.String()); err != nil {
		utils.ExecCmd("rkt", "--insecure-options=image", "fetch", aci.manifest.From.String())
	}
	if err := utils.ExecCmd("rkt", "image", "render", "--overwrite", aci.manifest.From.String(), aci.target); err != nil {
		panic("Cannot render from image" + aci.manifest.From.String() + err.Error())
	}
	os.Remove(aci.target + PATH_MANIFEST)
}

func (aci *Aci) copyInternals() {
	aci.log.Debug("Copy internals")
	os.MkdirAll(aci.rootfs+PATH_CNT+PATH_BIN, 0755)
	os.MkdirAll(aci.rootfs+"/bin", 0755)     // this is required or systemd-nspawn will create symlink on it
	os.MkdirAll(aci.rootfs+"/usr/bin", 0755) // this is required by systemd-nspawn

	busybox, _ := dist.Asset("dist/bindata/busybox")
	if err := ioutil.WriteFile(aci.rootfs+PATH_CNT+PATH_BIN+"/busybox", busybox, 0777); err != nil {
		panic(err)
	}

	confd, _ := dist.Asset("dist/bindata/confd")
	if err := ioutil.WriteFile(aci.rootfs+PATH_CNT+PATH_BIN+"/confd", confd, 0777); err != nil {
		panic(err)
	}

	attributeMerger, _ := dist.Asset("dist/bindata/attributes-merger")
	if err := ioutil.WriteFile(aci.rootfs+PATH_CNT+PATH_BIN+"/attributes-merger", attributeMerger, 0777); err != nil {
		panic(err)
	}

	confdFile := `backend = "env"
confdir = "/cnt"
prefix = "/confd"
log-level = "debug"
`
	os.MkdirAll(aci.rootfs+PATH_CNT+"/prestart", 0755)
	if err := ioutil.WriteFile(aci.rootfs+PATH_CNT+"/prestart/confd.toml", []byte(confdFile), 0777); err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile(aci.rootfs+PATH_CNT+PATH_BIN+"/prestart", []byte(PRESTART), 0777); err != nil {
		panic(err)
	}
}

func (aci *Aci) copyRunlevelsScripts() {
	aci.log.Debug("Copy Runlevels scripts")
	utils.CopyDir(aci.path+PATH_RUNLEVELS+PATH_BUILD, aci.target+PATH_RUNLEVELS+PATH_BUILD)
	utils.CopyDir(aci.path+PATH_RUNLEVELS+PATH_BUILD_LATE, aci.target+PATH_RUNLEVELS+PATH_BUILD_LATE)
	utils.CopyDir(aci.path+PATH_RUNLEVELS+PATH_BUILD_SETUP, aci.target+PATH_RUNLEVELS+PATH_BUILD_SETUP)
	utils.CopyDir(aci.path+PATH_RUNLEVELS+PATH_PRESTART_EARLY, aci.target+PATH_RUNLEVELS+PATH_PRESTART_EARLY)
	utils.CopyDir(aci.path+PATH_RUNLEVELS+PATH_PRESTART_LATE, aci.target+PATH_RUNLEVELS+PATH_PRESTART_LATE)

	utils.CopyDir(aci.path+PATH_RUNLEVELS+PATH_PRESTART_EARLY, aci.target+PATH_ROOTFS+PATH_CNT+PATH_RUNLEVELS+PATH_PRESTART_EARLY)
	utils.CopyDir(aci.path+PATH_RUNLEVELS+PATH_PRESTART_LATE, aci.target+PATH_ROOTFS+PATH_CNT+PATH_RUNLEVELS+PATH_PRESTART_LATE)
	utils.CopyDir(aci.path+PATH_RUNLEVELS+PATH_INHERIT_BUILD_EARLY, aci.target+PATH_ROOTFS+PATH_CNT+PATH_RUNLEVELS+PATH_INHERIT_BUILD_EARLY)
	utils.CopyDir(aci.path+PATH_RUNLEVELS+PATH_INHERIT_BUILD_LATE, aci.target+PATH_ROOTFS+PATH_CNT+PATH_RUNLEVELS+PATH_INHERIT_BUILD_LATE)
}

func (aci *Aci) runLevelBuildSetup() {
	files, err := ioutil.ReadDir(aci.path + PATH_RUNLEVELS + PATH_BUILD_SETUP)
	if err != nil {
		return
	}

	os.Setenv("BASEDIR", aci.path)
	os.Setenv("TARGET", aci.target)
	for _, f := range files {
		if !f.IsDir() {
			aci.log.WithField("file", f.Name()).Info("Running Build setup level script")
			if err := utils.ExecCmd(aci.path + PATH_RUNLEVELS + PATH_BUILD_SETUP + "/" + f.Name()); err != nil {
				panic(err)
			}
		}
	}
}

func (aci *Aci) copyConfd() {
	utils.CopyDir(aci.path+PATH_CONFD+PATH_CONFDOTD, aci.rootfs+PATH_CNT+PATH_CONFDOTD)
	utils.CopyDir(aci.path+PATH_CONFD+PATH_TEMPLATES, aci.rootfs+PATH_CNT+PATH_TEMPLATES)
}

func (aci *Aci) copyFiles() {
	utils.CopyDir(aci.path+PATH_FILES, aci.rootfs)
}

func (aci *Aci) copyAttributes() {
	files, err := utils.AttributeFiles(aci.path + PATH_ATTRIBUTES)
	if err != nil {
		aci.log.WithError(err).Fatal("Cannot read attribute files")
	}
	for _, file := range files {
		targetPath := aci.rootfs + PATH_CNT + PATH_ATTRIBUTES + "/" + aci.manifest.NameAndVersion.ShortName()
		err = os.MkdirAll(targetPath, 0755)
		if err != nil {
			aci.log.WithField("path", targetPath).WithError(err).Fatal("Cannot create target attribute directory")
		}
		if err := utils.CopyFile(file, targetPath+"/"+filepath.Base(file)); err != nil {
			aci.log.WithField("file", file).WithError(err).Fatal("Cannot copy attribute file")
		}
	}
}

func (aci *Aci) writeImgManifest() {
	aci.log.Debug("Writing aci manifest")
	utils.WriteImageManifest(&aci.manifest, aci.target+PATH_MANIFEST, aci.manifest.NameAndVersion.Name())
}

func checkSystemdNspawn() {
	_, err := utils.ExecCmdGetOutput("systemd-nspawn", "--version")
	if err != nil {
		logrus.WithError(err).Fatal("systemd-nspawn is required")
	}
}
