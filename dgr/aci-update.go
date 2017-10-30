package main

import (
	"os"
	"strings"

	"io/ioutil"

	"github.com/blablacar/dgr/dgr/common"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
)

const pathTemplates = "/templates"
const pathAttributes = "/attributes"
const pathPrestartEarly = "/runlevels/prestart-early"
const pathPrestartLate = "/runlevels/prestart-late"
const pathFiles = "/files"
const pathDgr = "/dgr"

const rootfs = "rootfs"
const rootfsDgr = rootfs + pathDgr
const inAciTemplatePath = rootfsDgr + pathTemplates
const inAciAttributesPath = rootfsDgr + pathAttributes
const inAciPrestartEarlyPath = rootfsDgr + pathPrestartEarly
const inAciPrestartLatePath = rootfsDgr + pathPrestartLate

func (aci *Aci) Update() error {
	aci.EnsureBuilt()
	return aci.RunUpdate()
}

func (aci *Aci) RunUpdate() error {
	logs.WithF(aci.fields).Info("Updating")

	updatedFile := aci.target + pathUpdated
	file, err := os.OpenFile(updatedFile, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return errs.WithEF(err, aci.fields.WithField("path", updatedFile), "Failed to create update file")
	}
	file.Close()

	if err := aci.deleteInTar(); err != nil {
		return err
	}
	if err := aci.addDirToTar(pathAttributes, inAciAttributesPath+"/"+aci.manifest.NameAndVersion.TinyName()); err != nil {
		return err
	}
	if err := aci.addDirToTar(pathTemplates, inAciTemplatePath); err != nil {
		return err
	}
	if err := aci.addDirToTar(pathPrestartEarly, inAciPrestartEarlyPath); err != nil {
		return err
	}
	if err := aci.addDirToTar(pathPrestartLate, inAciPrestartLatePath); err != nil {
		return err
	}
	if err := aci.addDirToTar(pathFiles, rootfs); err != nil {
		return err
	}

	return nil
}

func (aci *Aci) deleteInTar() error {
	var files []string
	if err := getFileList(aci.path+pathFiles, rootfs, &files); err != nil {
		return err
	}

	args := []string{"--delete", "--ignore-failed-read", "-f", aci.target + pathImageAci,
		inAciAttributesPath + "/" + aci.manifest.NameAndVersion.TinyName(),
		inAciTemplatePath,
		inAciPrestartEarlyPath,
		inAciPrestartLatePath,
	}
	args = append(args, files...)
	updateExec(args...) // err is ignored because it may already not exists
	return nil
}

func getFileList(dir string, relativePath string, res *[]string) error {
	files, err := ioutil.ReadDir(dir)
	logs.WithField("dir", dir).Debug("Reading directory")
	if err != nil {
		return errs.WithEF(err, data.WithField("dir", dir), "Cannot read directory")
	}

	for _, f := range files {
		if f.IsDir() {
			if err := getFileList(dir+"/"+f.Name(), relativePath+"/"+f.Name(), res); err != nil {
				return err
			}
		} else {
			*res = append(*res, relativePath+"/"+f.Name())
		}
	}

	return nil
}

func (aci *Aci) addDirToTar(localPath string, inAciPath string) error {
	aciPath := aci.path + localPath
	if _, err := os.Stat(aciPath); err == nil {
		if err := updateExec("--owner=0", "--group=0", "-rf",
			aci.target+pathImageAci,
			"--transform",
			"s,"+strings.TrimPrefix(aciPath, "/")+","+inAciPath+",",
			aciPath); err != nil {
			return errs.WithE(err, "Failed to add path to aci")
		}
	}
	return nil
}

func updateExec(args ...string) error {
	out, stderr, err := common.ExecCmdGetStdoutAndStderr("tar", args...)
	if err != nil {
		return errs.WithEF(err, data.
			WithField("stdout", out).
			WithField("stderr", stderr), "Tar update failed")
	}
	if logs.IsDebugEnabled() {
		println(out)
	}
	return nil
}
