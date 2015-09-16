package builder

import (
	"github.com/appc/spec/discovery"
	"github.com/blablacar/cnt/config"
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
	cnt.copyRunlevelsScripts()

	cnt.runlevelBuildSetup()

	cnt.writeImgManifest()
	cnt.writeCntManifest() // TODO move that, here because we update the version number to generated version

	cnt.runBuild()
	cnt.copyRunlevelsPrestart()
	cnt.copyAttributes()
	cnt.copyConfd()
	cnt.copyFiles()
	cnt.runBuildLate()

	cnt.tarAci()
	//	ExecCmd("chown " + os.Getenv("SUDO_USER") + ": " + target + "/*") //TODO chown
	return nil
}

///////////////////////////////////////////////////////

func (cnt *Img) writeCntManifest() {
	utils.CopyFile(cnt.path+IMG_MANIFEST, cnt.target+IMG_MANIFEST)
}

func (cnt *Img) runBuildLate() {
	res, err := utils.IsDirEmpty(cnt.target + RUNLEVELS_BUILD_LATE)
	res2, err2 := utils.IsDirEmpty(cnt.target + RUNLEVELS_BUILD_INHERIT_LATE)
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
	if res, err := utils.IsDirEmpty(cnt.target + RUNLEVELS_BUILD); res || err != nil {
		return
	}

	{
		rootfs := "${TARGET}/rootfs"
		if cnt.manifest.Build.NoBuildImage() {
			rootfs = ""
		}
		build := strings.Replace(BUILD_SCRIPT, "%%ROOTFS%%", rootfs, 1)
		ioutil.WriteFile(cnt.target+"/build.sh", []byte(build), 0777)
	}

	if err := utils.ExecCmd("systemd-nspawn", "--version"); err == nil {
		log.Get().Info("Run with systemd-nspawn")
		if err := utils.ExecCmd("systemd-nspawn", "--directory="+cnt.rootfs, "--capability=all",
			"--bind="+cnt.target+"/:/target", "--share-system", "target/build.sh"); err != nil {
			log.Get().Panic("Build step did not succeed", err)
		}
	} else {
		log.Get().Panic("systemd-nspawn is required")
		//		log.Get().Info("Run with docker")
		//
		//		//
		//		log.Get().Info("Prepare Docker")
		//		first := exec.Command("bash", "-c", "cd "+cnt.rootfs+" && tar cf - .")
		//		second := exec.Command("docker", "import", "-", "")
		//
		//		reader, writer := io.Pipe()
		//		first.Stdout = writer
		//		second.Stdin = reader
		//
		//		var buff bytes.Buffer
		//		second.Stdout = &buff
		//
		//		first.Start()
		//		second.Start()
		//		first.Wait()
		//		writer.Close()
		//		second.Wait()
		//		imgId := strings.TrimSpace(buff.String())
		//
		//		//
		//		log.Get().Info("Run Docker\n")
		//		cmd := []string{"run", "--name=" + cnt.manifest.NameAndVersion.ShortName(), "-v", cnt.target + ":/target", imgId, "/target/build.sh"}
		//		utils.ExecCmd("docker", "rm", cnt.manifest.NameAndVersion.ShortName())
		//		if err := utils.ExecCmd("docker", cmd...); err != nil {
		//			panic(err)
		//		}
		//
		//		//
		//		log.Get().Info("Release Docker")
		//		if cnt.manifest.Build.NoBuildImage() {
		//			os.RemoveAll(cnt.rootfs)
		//			os.Mkdir(cnt.rootfs, 0777)
		//
		//			if err := utils.ExecCmd("docker", "export", "-o", cnt.target+"/dockerfs.tar", cnt.manifest.NameAndVersion.ShortName()); err != nil {
		//				panic(err)
		//			}
		//
		//			utils.ExecCmd("tar", "xpf", cnt.target+"/dockerfs.tar", "-C", cnt.rootfs)
		//		}
		//		if err := utils.ExecCmd("docker", "rm", cnt.manifest.NameAndVersion.ShortName()); err != nil {
		//			panic(err)
		//		}
		//		if err := utils.ExecCmd("docker", "rmi", imgId); err != nil {
		//			panic(err)
		//		}
	}
}

func (cnt *Img) processFrom() {
	if cnt.manifest.From != "" {
		log.Get().Info("Prepare rootfs from " + cnt.manifest.From)

		app, err := discovery.NewAppFromString(string(cnt.manifest.From))
		if app.Labels["os"] == "" {
			app.Labels["os"] = "linux"
		}
		if app.Labels["arch"] == "" {
			app.Labels["arch"] = "amd64"
		}

		endpoint, _, err := discovery.DiscoverEndpoints(*app, false)
		if err != nil {
			panic(err)
		}

		url := endpoint.ACIEndpoints[0].ACI

		aciPath := config.GetConfig().AciPath + "/" + string(cnt.manifest.From)
		if _, err := os.Stat(aciPath + "/image.aci"); cnt.args.ForceUpdate || os.IsNotExist(err) {
			if err := os.MkdirAll(aciPath, 0755); err != nil {
				log.Get().Panic(err)
			}
			if err = utils.ExecCmd("wget", "-O", aciPath+"/image.aci", url); err != nil {
				os.Remove(aciPath + "/image.aci")
				log.Get().Panic("Cannot download from image", err)
			}
		} else {
			log.Get().Info("Image " + cnt.manifest.From + " Already exists locally, will not be downloaded")
		}

		utils.ExecCmd("tar", "xpf", aciPath+"/image.aci", "-C", cnt.target)

		//		utils.ExecCmd("rkt", "--insecure-skip-verify=true", "fetch", cnt.manifest.From)
		//		utils.ExecCmd("rkt", "image", "export", "--overwrite", cnt.manifest.From, cnt.target + "/from.aci")
		//		utils.ExecCmd("tar", "xf", cnt.target + "/from.aci", "-C", cnt.target)
		//		os.Remove(cnt.target + "/from.aci")
	}
}

func (cnt *Img) copyRunlevelsScripts() {
	if err := os.MkdirAll(cnt.target+RUNLEVELS, 0755); err != nil {
		log.Get().Panic(err)
	}
	utils.CopyDir(cnt.path+RUNLEVELS, cnt.target+RUNLEVELS)
}

func (cnt *Img) runlevelBuildSetup() {
	files, err := ioutil.ReadDir(cnt.path + RUNLEVELS_BUILD_SETUP)
	if err != nil {
		return
	}

	os.Setenv("BASEDIR", cnt.path)
	os.Setenv("TARGET", cnt.target)
	for _, f := range files {
		if !f.IsDir() {
			log.Get().Info("Running Build setup level : ", f.Name())
			if err := utils.ExecCmd(cnt.path + RUNLEVELS_BUILD_SETUP + "/" + f.Name()); err != nil {
				log.Get().Panic(err)
			}
		}
	}
}

func (cnt *Img) tarAci() {
	dir, _ := os.Getwd()
	log.Get().Debug("chdir to", cnt.target)
	os.Chdir(cnt.target)

	args := []string{"manifest", "rootfs/"}

	if _, err := os.Stat(cnt.path + RUNLEVELS_BUILD_INHERIT_EARLY); err == nil {
		args = append(args, strings.TrimPrefix(RUNLEVELS_BUILD_INHERIT_EARLY, "/"))
	}
	if _, err := os.Stat(cnt.path + RUNLEVELS_BUILD_INHERIT_LATE); err == nil {
		args = append(args, strings.TrimPrefix(RUNLEVELS_BUILD_INHERIT_LATE, "/"))
	}

	utils.Tar(cnt.args.Zip, "image.aci", args...)
	log.Get().Debug("chdir to", dir)
	os.Chdir(dir)
}

//func (cnt *Cnt) copyInstallAndCreatePacker() {
//	if _, err := os.Stat(cnt.path + "/install.sh"); err == nil {
//		utils.CopyFile(cnt.path + "/install.sh", cnt.target + "/install.sh")
//		sum, _ := utils.ChecksumFile(cnt.target + "/install.sh")
//		lastSum, err := ioutil.ReadFile(cnt.target + "/install.sh.SUM")
//		if err != nil || !bytes.Equal(lastSum, sum) {
//			utils.WritePackerFiles(cnt.target)
//			ioutil.WriteFile(cnt.target + "/install.sh.SUM", sum, 0755)
//			return
//		}
//	}
//	utils.RemovePackerFiles(cnt.target)
//}

func (cnt *Img) copyRunlevelsPrestart() {
	if err := os.MkdirAll(cnt.rootfs+"/etc/prestart/late-prestart.d", 0755); err != nil {
		log.Get().Panic(err)
	}
	if err := os.MkdirAll(cnt.rootfs+"/etc/prestart/early-prestart.d", 0755); err != nil {
		log.Get().Panic(err)
	}
	utils.CopyDir(cnt.path+RUNLEVELS_PRESTART, cnt.rootfs+"/etc/prestart/early-prestart.d")
	utils.CopyDir(cnt.path+RUNLEVELS_LATESTART, cnt.rootfs+"/etc/prestart/late-prestart.d")
}

func (cnt *Img) copyConfd() {
	if err := os.MkdirAll(cnt.rootfs+"/etc/prestart/", 0755); err != nil {
		log.Get().Panic(err)
	}
	utils.CopyDir(cnt.path+CONFD_CONFIG, cnt.rootfs+"/etc/prestart/conf.d")
	utils.CopyDir(cnt.path+CONFD_TEMPLATE, cnt.rootfs+"/etc/prestart/templates")
}

func (cnt *Img) copyFiles() {
	utils.CopyDir(cnt.path+FILES_PATH, cnt.rootfs)
}

func (cnt *Img) copyAttributes() {
	if err := os.MkdirAll(cnt.rootfs+"/etc/prestart/attributes/"+cnt.manifest.NameAndVersion.ShortName(), 0755); err != nil {
		log.Get().Panic(err)
	}
	utils.CopyDir(cnt.path+ATTRIBUTES, cnt.rootfs+"/etc/prestart/attributes/"+cnt.manifest.NameAndVersion.ShortName())
}

func (cnt *Img) writeImgManifest() {
	log.Get().Debug("Writing aci manifest")
	version := cnt.manifest.NameAndVersion.Version()
	if version == "" {
		version = utils.GenerateVersion()
	}
	utils.WriteImageManifest(&cnt.manifest, cnt.target+"/manifest", cnt.manifest.NameAndVersion.Name(), version)
}
