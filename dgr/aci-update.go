package main

import (
	"strings"

	"os"

	"github.com/blablacar/dgr/dgr/common"
	"github.com/n0rad/go-erlog/logs"
)

func (aci *Aci) Update() error {
	aci.EnsureBuilt()
	return aci.RunUpdate()
}

func (aci *Aci) RunUpdate() error {
	logs.WithField("target", aci.target).Info("Experimental features. Prefer a clean rebuild before push.")
	if _, err := os.Stat(aci.path + "/attributes"); err == nil {
		common.ExecCmd("tar", "--delete", "rootfs/dgr/attributes/"+string(aci.manifest.NameAndVersion.ShortName()), "-f", aci.target+"/image.aci")
		common.ExecCmd("tar", "--owner=0", "--group=0", "-rf", aci.target+"/image.aci", "--transform", "s,"+strings.TrimPrefix(aci.path, "/")+"/attributes,rootfs/dgr/attributes/"+string(aci.manifest.NameAndVersion.ShortName())+",", aci.path+"/attributes")
		common.ExecCmd("tar", "-tvf", aci.target+"/image.aci", "rootfs/dgr/attributes/")
	}
	if _, err := os.Stat(aci.path + "/templates"); err == nil {
		common.ExecCmd("tar", "--delete", "rootfs/dgr/templates/", "-f", aci.target+"/image.aci")
		common.ExecCmd("tar", "--owner=0", "--group=0", "-rf", aci.target+"/image.aci", "--transform", "s,"+strings.TrimPrefix(aci.path, "/")+"/templates,rootfs/dgr/templates/,", aci.path+"/templates")
		common.ExecCmd("tar", "-tvf", aci.target+"/image.aci", "rootfs/dgr/templates/")
	}
	logs.Info("Update Done. Prefer a clean rebuild before push.")

	return nil
}
