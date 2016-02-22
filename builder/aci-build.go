package builder

import (
	"github.com/blablacar/dgr/dgr"
	"github.com/blablacar/dgr/dist"
	"github.com/blablacar/dgr/utils"
	"github.com/n0rad/go-erlog/logs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func (aci *Aci) Build() error {
	aci.Clean()

	logs.WithF(aci.fields).Info("Building")

	os.MkdirAll(aci.rootfs, 0777)
	os.MkdirAll(aci.target+PATH_BUILDER, 0777)

	aci.fullyResolveDependencies()
	aci.processBuilder()
	aci.processFrom()
	aci.copyInternals()
	aci.copyRunlevelsScripts()

	aci.runLevelBuildSetup()

	aci.runBuild()
	aci.copyAttributes()
	aci.copyTemplates()
	aci.copyFiles()
	aci.runBuildLate()

	aci.writeAciManifest()

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

func (aci *Aci) processBuilder() {
	if aci.manifest.Builder.String() == "" {
		return
	}
	logs.WithF(aci.fields).WithField("builder", aci.manifest.Builder.String()).Debug("Process builder")

	if err := utils.ExecCmd("bash", "-c", "rkt image list --fields name --no-legend | grep -q "+aci.manifest.Builder.String()); err != nil {
		utils.ExecCmd("rkt", "--insecure-options=image", "fetch", aci.manifest.Builder.String())
	}
	if err := utils.ExecCmd("rkt", "image", "render", "--overwrite", aci.manifest.Builder.String(), aci.target+PATH_TMP); err != nil {
		logs.WithEF(err, aci.fields).WithField("builder", aci.manifest.Builder.String()).Fatal("Failed to render builder aci")
	}

	err := utils.ExecCmd("bash", "mv", aci.target+PATH_TMP+PATH_ROOTFS, aci.target+PATH_BUILDER)
	if err != nil {
		logs.WithEF(err, aci.fields).Fatal("Failed to move rendered build rootfs")
	}
	os.RemoveAll(aci.target + PATH_TMP)
}

func (aci *Aci) fullyResolveDependencies() {
	if !aci.FullyResolveDep {
		return
	}

	for i, dep := range aci.manifest.Aci.Dependencies {
		resolved, err := dep.FullyResolved()
		if err != nil {
			logs.WithEF(err, aci.fields.WithField("dependency", dep)).Fatal("Cannot fully resolve dependency")
		}
		aci.manifest.Aci.Dependencies[i] = *resolved
	}
}

func (aci *Aci) runBuildLate() {
	res, err := utils.IsDirEmpty(aci.target + PATH_RUNLEVELS + PATH_BUILD_LATE)
	res2, err2 := utils.IsDirEmpty(aci.rootfs + PATH_DGR + PATH_RUNLEVELS + PATH_INHERIT_BUILD_LATE)
	if (res && res2) || (err != nil && err2 != nil) {
		return
	}
	logs.WithF(aci.fields).Debug("Running build late")

	checkSystemdNspawn()

	rootfs := ""
	builderRootfs := aci.target + PATH_ROOTFS
	if res, _ := utils.IsDirEmpty(aci.target + PATH_BUILDER); !res && err == nil {
		builderRootfs = aci.target + PATH_BUILDER
		rootfs = "/target/rootfs"
	}

	build := strings.Replace(BUILD_SCRIPT_LATE, "%%ROOTFS%%", rootfs, -1)
	ioutil.WriteFile(aci.target+"/build-late.sh", []byte(build), 0777)

	logs.WithF(aci.fields).Info("Starting systemd-nspawn to run Build late scripts")
	if err := utils.ExecCmd("systemd-nspawn",
		"--setenv=LOG_LEVEL="+logs.GetLevel().String(),
		"--register=no",
		"-q",
		"--directory="+builderRootfs,
		"--capability=all",
		"--bind="+aci.target+"/:/target",
		"target/build-late.sh"); err != nil {
		logs.WithEF(err, aci.fields).Fatal("Build late part failed")
	}
}

func (aci *Aci) runBuild() { // TODO merge with runBuildLate
	if res, err := utils.IsDirEmpty(aci.target + PATH_RUNLEVELS + PATH_BUILD); res || err != nil {
		return
	}
	logs.WithF(aci.fields).Debug("Running build")

	checkSystemdNspawn()

	rootfs := ""
	builderRootfs := aci.target + PATH_ROOTFS
	if res, err := utils.IsDirEmpty(aci.target + PATH_BUILDER); !res && err == nil {
		builderRootfs = aci.target + PATH_BUILDER
		rootfs = "/target/rootfs"
	}

	build := strings.Replace(BUILD_SCRIPT, "%%ROOTFS%%", rootfs, -1)
	ioutil.WriteFile(aci.target+"/build.sh", []byte(build), 0777)

	logs.WithF(aci.fields).Info("Starting systemd-nspawn to run Build scripts")
	if err := utils.ExecCmd("systemd-nspawn",
		"--setenv=LOG_LEVEL="+logs.GetLevel().String(),
		"--register=no",
		"-q",
		"--directory="+builderRootfs,
		"--capability=all",
		"--bind="+aci.target+"/:/target",
		"target/build.sh"); err != nil {
		logs.WithEF(err, aci.fields).Fatal("Build part failed")
	}
}

func (aci *Aci) processFrom() {
	froms, err := aci.manifest.GetFroms()
	if err != nil {
		logs.WithEF(err, aci.fields).Fatal("Cannot process from")
	}
	for _, from := range froms {
		if from == "" {
			continue
		}

		logs.WithF(aci.fields).WithField("from", from).Info("Rendering from")

		if err := utils.ExecCmd("bash", "-c", "rkt image list --fields name --no-legend | grep -q "+from.String()); err != nil {
			utils.ExecCmd("rkt", "--insecure-options=image", "fetch", from.String())
		}
		if err := utils.ExecCmd("rkt", "image", "render", "--overwrite", from.String(), aci.target); err != nil {
			logs.WithEF(err, aci.fields).WithField("from", from.String()).Fatal("Failed to render from")
		}
	}

	os.Remove(aci.target + PATH_MANIFEST)
}

func (aci *Aci) copyInternals() {
	logs.WithF(aci.fields).Debug("Copy internals")
	os.MkdirAll(aci.rootfs+PATH_DGR+PATH_BIN, 0755)
	//	os.MkdirAll(aci.rootfs+PATH_BIN, 0755)   // this is required or systemd-nspawn will create symlink on it
	os.MkdirAll(aci.rootfs+"/usr/bin", 0755) // this is required by systemd-nspawn

	busybox, _ := dist.Asset("dist/bindata/busybox")
	if err := ioutil.WriteFile(aci.rootfs+PATH_DGR+PATH_BIN+"/busybox", busybox, 0777); err != nil {
		panic(err)
	}

	templater, _ := dist.Asset("dist/bindata/templater")
	if err := ioutil.WriteFile(aci.rootfs+PATH_DGR+PATH_BIN+"/templater", templater, 0777); err != nil {
		logs.WithEF(err, aci.fields).Fatal("Failed to copy templater")
	}

	os.MkdirAll(aci.rootfs+PATH_DGR+"/prestart", 0755)
	if err := ioutil.WriteFile(aci.rootfs+PATH_DGR+PATH_BIN+"/prestart", []byte(PRESTART), 0777); err != nil {
		logs.WithEF(err, aci.fields).Fatal("Failed to write prestart")
	}

	if err := ioutil.WriteFile(aci.rootfs+PATH_DGR+PATH_BIN+"/functions.sh", []byte(SH_FUNCTIONS), 0777); err != nil {
		logs.WithEF(err, aci.fields).Fatal("Failed to write functions.sh")
	}
}

func (aci *Aci) copyRunlevelsScripts() {
	logs.WithF(aci.fields).Debug("Copy Runlevels scripts")
	utils.CopyDir(aci.path+PATH_RUNLEVELS+PATH_BUILD, aci.target+PATH_RUNLEVELS+PATH_BUILD)
	utils.CopyDir(aci.path+PATH_RUNLEVELS+PATH_BUILD_LATE, aci.target+PATH_RUNLEVELS+PATH_BUILD_LATE)
	utils.CopyDir(aci.path+PATH_RUNLEVELS+PATH_BUILD_SETUP, aci.target+PATH_RUNLEVELS+PATH_BUILD_SETUP)
	utils.CopyDir(aci.path+PATH_RUNLEVELS+PATH_PRESTART_EARLY, aci.target+PATH_RUNLEVELS+PATH_PRESTART_EARLY)
	utils.CopyDir(aci.path+PATH_RUNLEVELS+PATH_PRESTART_LATE, aci.target+PATH_RUNLEVELS+PATH_PRESTART_LATE)

	utils.CopyDir(aci.path+PATH_RUNLEVELS+PATH_PRESTART_EARLY, aci.target+PATH_ROOTFS+PATH_DGR+PATH_RUNLEVELS+PATH_PRESTART_EARLY)
	utils.CopyDir(aci.path+PATH_RUNLEVELS+PATH_PRESTART_LATE, aci.target+PATH_ROOTFS+PATH_DGR+PATH_RUNLEVELS+PATH_PRESTART_LATE)
	utils.CopyDir(aci.path+PATH_RUNLEVELS+PATH_INHERIT_BUILD_EARLY, aci.target+PATH_ROOTFS+PATH_DGR+PATH_RUNLEVELS+PATH_INHERIT_BUILD_EARLY)
	utils.CopyDir(aci.path+PATH_RUNLEVELS+PATH_INHERIT_BUILD_LATE, aci.target+PATH_ROOTFS+PATH_DGR+PATH_RUNLEVELS+PATH_INHERIT_BUILD_LATE)
}

func (aci *Aci) runLevelBuildSetup() {
	_, err := ioutil.ReadDir(aci.path + PATH_RUNLEVELS + PATH_BUILD_SETUP)
	if err != nil {
		return
	}

	logs.WithF(aci.fields).Info("Running build setup scripts")

	if err := ioutil.WriteFile(aci.target+"/build-setup.sh", []byte(BUILD_SETUP), 0777); err != nil {
		logs.WithEF(err, aci.fields).Fatal("Failed to write build setup script")
	}

	os.Setenv("BASEDIR", aci.path)
	os.Setenv("TARGET", aci.target)
	os.Setenv("LOG_LEVEL", logs.GetLevel().String())

	if err := utils.ExecCmd(aci.target + "/build-setup.sh"); err != nil {
		logs.WithEF(err, aci.fields).Fatal("Build setup failed")
	}
}

func (aci *Aci) copyTemplates() {
	utils.CopyDir(aci.path+PATH_TEMPLATES, aci.rootfs+PATH_DGR+PATH_TEMPLATES)
}

func (aci *Aci) copyFiles() {
	utils.CopyDir(aci.path+PATH_FILES, aci.rootfs)
}

func (aci *Aci) copyAttributes() {
	files, err := utils.AttributeFiles(aci.path + PATH_ATTRIBUTES)
	if err != nil {
		logs.WithEF(err, aci.fields).Fatal("Cannot read attribute files")
	}
	for _, file := range files {
		targetPath := aci.rootfs + PATH_DGR + PATH_ATTRIBUTES + "/" + aci.manifest.NameAndVersion.ShortName()
		err = os.MkdirAll(targetPath, 0755)
		if err != nil {
			logs.WithEF(err, aci.fields.WithField("path", targetPath)).Fatal("Cannot create target attribute directory")
		}
		if err := utils.CopyFile(file, targetPath+"/"+filepath.Base(file)); err != nil {
			logs.WithEF(err, aci.fields.WithField("file", file)).Fatal("Cannot copy attribute file")
		}
	}
}

func (aci *Aci) writeAciManifest() {
	logs.WithF(aci.fields).Debug("Writing aci manifest")
	utils.WriteImageManifest(&aci.manifest, aci.target+PATH_MANIFEST, aci.manifest.NameAndVersion.Name(), dgr.Version)
}

func checkSystemdNspawn() {
	_, err := utils.ExecCmdGetOutput("systemd-nspawn", "--version")
	if err != nil {
		logs.WithE(err).Fatal("systemd-nspawn is required")
	}
}
