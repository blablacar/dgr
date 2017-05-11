package main

import (
	"os"
	"strings"

	"github.com/blablacar/dgr/dgr/common"
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

	aciName := string(aci.manifest.NameAndVersion.TinyNameId())
	aci.updateDirInTar(pathAttributes, inAciAttributesPath+"/"+aciName)
	aci.updateDirInTar(pathTemplates, inAciTemplatePath)
	aci.updateDirInTar(pathPrestartEarly, inAciPrestartEarlyPath)
	aci.updateDirInTar(pathPrestartLate, inAciPrestartLatePath)

	return nil
}

func (aci *Aci) updateDirInTar(localPath string, inAciPath string) {
	aciPath := aci.path + localPath
	if _, err := os.Stat(aciPath); err == nil {
		common.ExecCmd("tar", "--delete", inAciPath, "-f", aci.target+pathImageAci)
		common.ExecCmd("tar", "--owner=0", "--group=0", "-rf",
			aci.target+pathImageAci,
			"--transform",
			"s,"+strings.TrimPrefix(aciPath, "/")+","+inAciPath+",",
			aciPath)
		common.ExecCmd("tar", "-tvf", aci.target+pathImageAci, inAciPath)
	}
}
