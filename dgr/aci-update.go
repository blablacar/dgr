package main

import (
	"strings"

	"os"

	"github.com/blablacar/dgr/dgr/common"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
)

const inAciTemplatePath = "rootfs/dgr/templates/"
const inAciAttributesPath = "rootfs/dgr/attributes/"
const pathTemplates = "/templates"
const pathAttributes = "/attributes"

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

	aciName := string(aci.manifest.NameAndVersion.ShortName())
	if _, err := os.Stat(aci.path + pathAttributes); err == nil {
		common.ExecCmd("tar", "--delete", inAciAttributesPath+aciName, "-f", aci.target+pathImageAci)
		common.ExecCmd("tar", "--owner=0", "--group=0", "-rf",
			aci.target+pathImageAci,
			"--transform",
			"s,"+strings.TrimPrefix(aci.path, "/")+pathAttributes+","+inAciAttributesPath+aciName+",",
			aci.path+pathAttributes)
		common.ExecCmd("tar", "-tvf", aci.target+pathImageAci, inAciAttributesPath)
	}
	if _, err := os.Stat(aci.path + pathTemplates); err == nil {
		common.ExecCmd("tar", "--delete", inAciTemplatePath, "-f", aci.target+pathImageAci)
		common.ExecCmd("tar", "--owner=0", "--group=0", "-rf",
			aci.target+pathImageAci,
			"--transform",
			"s,"+strings.TrimPrefix(aci.path, "/")+pathTemplates+","+inAciTemplatePath+",",
			aci.path+pathTemplates)
		common.ExecCmd("tar", "-tvf", aci.target+pathImageAci, inAciTemplatePath)
	}
	return nil
}
