package main

import (
	"github.com/blablacar/dgr/bin-dgr/common"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
)

func (aci *Aci) Sign() error {
	logs.WithF(aci.fields).Debug("Signing")

	if err := aci.EnsureZip(); err != nil {
		return errs.WithEF(err, aci.fields, "Failed to prepare image for signature")
	}

	err := common.ExecCmd("gpg", "--armor", "--output", aci.target+PATH_IMAGE_GZ_ACI_ASC, "--detach-sig", aci.target+PATH_IMAGE_GZ_ACI)
	if err != nil {
		return errs.WithEF(err, aci.fields, "Failed to sign image")
	}
	return nil
}
