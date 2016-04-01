package main

import "github.com/blablacar/dgr/bin-dgr/common"

func (aci *Aci) CleanAndTry() error {
	aci.Clean()
	return aci.RunBuilderCommand(common.CommandTry)
}
