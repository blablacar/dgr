package main

import (
	"github.com/blablacar/dgr/bin-dgr/common"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"strings"
)

func (aci *Aci) Push() error {
	defer aci.giveBackUserRightsToTarget()
	if Home.Config.Push.Type == "" {
		return errs.WithF(aci.fields, "Cannot push, push is not configured in dgr global configuration file")
	}

	if err := aci.EnsureBuilt(); err != nil {
		return err
	}

	if aci.args.Test {
		aci.args.Test = false
		aci.Test()
	}

	logs.WithF(aci.fields).Info("Gzipping aci before upload")
	aci.tarAci(aci.target, true)

	im, err := common.ExtractManifestFromAci(aci.target + PATH_IMAGE_ACI_ZIP)
	if err != nil {
		return errs.WithEF(err, aci.fields.WithField("file", PATH_IMAGE_ACI_ZIP), "Failed to extract manifest from aci file")
	}
	val, ok := im.Labels.Get("version")
	if !ok {
		return errs.WithEF(err, aci.fields.WithField("file", PATH_IMAGE_ACI_ZIP), "Failed to get version from aci manifest")
	}

	logs.WithF(aci.fields).Info("Uploading aci")
	if err := common.ExecCmd("curl", "-f", "-i",
		"-F", "r=releases",
		"-F", "hasPom=false",
		"-F", "e=aci",
		"-F", "g=com.blablacar.aci.linux.amd64",
		"-F", "p=aci",
		"-F", "v="+val,
		"-F", "a="+strings.Split(string(im.Name), "/")[1],
		"-F", "file=@"+aci.target+PATH_IMAGE_ACI_ZIP,
		"-u", Home.Config.Push.Username+":"+Home.Config.Push.Password,
		Home.Config.Push.Url+"/service/local/artifact/maven/content"); err != nil {
		return errs.WithEF(err, aci.fields, "Failed to push aci")
	}
	return nil
}
