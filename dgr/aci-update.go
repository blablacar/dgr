package main

import (
	"os"
	"strings"

	"github.com/blablacar/dgr/dgr/common"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
)

const pathTemplates = "/templates"
const pathAttributes = "/attributes"
const pathPrestartEarly = "/runlevels/prestart-early"
const pathPrestartLate = "/runlevels/prestart-late"
const rootfsDgr = "rootfs/dgr"
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

	aciName := string(aci.manifest.NameAndVersion.TinyName())
	if err := aci.updateDirInTar(pathAttributes, inAciAttributesPath+"/"+aciName); err != nil {
		return err
	}
	if err := aci.updateDirInTar(pathTemplates, inAciTemplatePath); err != nil {
		return err
	}
	if err := aci.updateDirInTar(pathPrestartEarly, inAciPrestartEarlyPath); err != nil {
		return err
	}
	if err := aci.updateDirInTar(pathPrestartLate, inAciPrestartLatePath); err != nil {
		return err
	}

	return nil
}

func (aci *Aci) updateDirInTar(localPath string, inAciPath string) error {
	aciPath := aci.path + localPath
	if _, err := os.Stat(aciPath); err == nil {

		if err := updateExec(inAciPath, "--delete", inAciPath, "-f", aci.target+pathImageAci); err != nil {
			return errs.WithE(err, "Failed to delete path in aci")
		}

		if err := updateExec(inAciPath, "--owner=0", "--group=0", "-rf",
			aci.target+pathImageAci,
			"--transform",
			"s,"+strings.TrimPrefix(aciPath, "/")+","+inAciPath+",",
			aciPath); err != nil {
			return errs.WithE(err, "Failed to transform path in aci")
		}

		if err := updateExec(inAciPath, "-tvf", aci.target+pathImageAci, inAciPath); err != nil {
			return errs.WithE(err, "Failed to add path to aci")
		}
	}
	return nil
}

func updateExec(inAciPath string, args ...string) error {
	out, stderr, err := common.ExecCmdGetStdoutAndStderr("tar", args...)
	if err != nil {
		return errs.WithEF(err, data.
			WithField("path", inAciPath).
			WithField("stdout", out).
			WithField("stderr", stderr), "Tar update failed")
	}
	if logs.IsDebugEnabled() {
		println(out)
	}
	return nil
}
