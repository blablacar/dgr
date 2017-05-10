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
const pathRootfsDgr = "rootfs/dgr"
const inAciTemplatePath = pathRootfsDgr + pathTemplates
const inAciAttributesPath = pathRootfsDgr + pathAttributes
const inAciPrestartEarlyPath = pathRootfsDgr + pathPrestartEarly
const inAciPrestartLatePath = pathRootfsDgr + pathPrestartLate

func (aci *Aci) Update() error {
	aci.EnsureBuilt()
	return aci.RunUpdate()
}

func (aci *Aci) RunUpdate() error {
	logs.WithF(aci.fields).Debug("Updating")

	updatedFile := aci.target + pathUpdated
	file, err := os.OpenFile(updatedFile, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return errs.WithEF(err, aci.fields.WithField("path", updatedFile), "failed to create update file")
	}
	file.Close()

	aciName := string(aci.manifest.NameAndVersion.TinyNameId())
	logs.WithField("lol", aciName).Debug("Updating")
	aci.updateFileInTar(pathAttributes, inAciAttributesPath+"/"+aciName)
	aci.updateFileInTar(pathTemplates, inAciTemplatePath)
	aci.updateFileInTar(pathPrestartEarly, inAciPrestartEarlyPath)
	aci.updateFileInTar(pathPrestartLate, inAciPrestartLatePath)

	return nil
}

func (aci *Aci) updateFileInTar(localPath string, inAciPath string) {
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
