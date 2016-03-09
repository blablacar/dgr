package main

import "github.com/blablacar/dgr/bin-dgr/common"

func (aci *Aci) Init() error {
	return aci.RunBuilderCommand(common.COMMAND_INIT)
}
