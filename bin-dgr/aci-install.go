package main

import (
	"github.com/blablacar/dgr/bin-dgr/common"
	"github.com/n0rad/go-erlog/errs"
)

func (aci *Aci) Install() error {
	if err := aci.EnsureBuilt(); err != nil {
		return err
	}
	if aci.args.Test {
		aci.args.Test = false
		if err := aci.Test(); err != nil {
			return err
		}
	}
	_, stderr, err := common.ExecCmdGetStdoutAndStderr("rkt", "--insecure-options=image", "fetch", aci.target+PATH_IMAGE_ACI)
	if err != nil {
		return errs.WithEF(err, aci.fields.WithField("stderr", stderr), "Failed to install aci")
	}
	return nil
}
