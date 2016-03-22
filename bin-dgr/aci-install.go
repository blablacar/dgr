package main

import (
	"github.com/n0rad/go-erlog/errs"
)

func (aci *Aci) Install() ([]string, error) {
	hashs := []string{}

	if err := aci.EnsureBuilt(); err != nil {
		return hashs, err
	}
	if aci.args.Test {
		aci.args.Test = false
		if err := aci.Test(); err != nil {
			return hashs, err
		}
	}

	hash, err := Home.Rkt.Fetch(aci.target + PATH_IMAGE_ACI)
	if err != nil {
		return hashs, errs.WithEF(err, aci.fields, "Failed to install aci")
	}
	hashs = append(hashs, hash)
	return hashs, nil
}
