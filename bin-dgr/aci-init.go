package main

import (
	"github.com/blablacar/dgr/bin-dgr/common"
)

func (aci *Aci) Init() error {
	err := aci.RunBuilderCommand(common.COMMAND_INIT)
	aci.Clean() // TODO this is a ack because init will leave an empty aci in target. should better be processed by stage1
	return err
}
