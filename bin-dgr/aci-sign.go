package main

import (
	"github.com/blablacar/dgr/bin-dgr/common"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
)

func (aci *Aci) Sign() error {
	logs.WithF(aci.fields).Debug("Signing")
	if err := aci.EnsureBuilt(); err != nil {
		return err
	}

	return aci.signFile(aci.target + pathImageAci)
}

func (aci *Aci) ZipSign() error {
	logs.WithF(aci.fields).Debug("Zip Signing")
	if err := aci.EnsureZip(); err != nil {
		return err
	}

	return aci.signFile(aci.target + pathImageGzAci)
}

func (aci *Aci) signFile(file string) error {
	sign, err := Home.Config.GetSignKeyring(aci.manifest.NameAndVersion.DomainName())
	if err != nil {
		return errs.WithEF(err, aci.fields, "Failed to sign image. Cannot found keyring")
	}
	if sign.Disabled {
		logs.WithF(aci.fields).WithField("domain", aci.manifest.NameAndVersion.DomainName()).Warn("Sign disabled for this aci's domain")
		return nil
	}

	if err := common.ExecCmd("gpg", "--yes", "--no-default-keyring", "--armor",
		"--keyring", sign.Keyring, "--output", file+suffixAsc, "--detach-sig", file); err != nil {
		return errs.WithEF(err, aci.fields, "Failed to sign image")
	}
	return nil
}

// "--batch"
//"--secret-keyring",
//"/rkt.sec",

//err := common.ExecCmd("gpg", "--armor", "--output", aci.target+pathImageGzAciAsc, "--detach-sig", aci.target+pathImageGzAci)
//err := common.ExecCmd("gpg", "--no-default-keyring", "--secret-keyring", "/rkt.sec", "--keyring", "/rkt.pub", "--verify",
//	aci.target+pathImageGzAciAsc, aci.target+pathImageGzAci)
