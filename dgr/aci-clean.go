package main

import (
	"os"

	"github.com/n0rad/go-erlog/logs"
)

func (aci *Aci) Clean() {
	logs.WithF(aci.fields).Debug("Cleaning")

	if err := os.RemoveAll(aci.target + "/"); err != nil {
		logs.WithEF(err, aci.fields).WithField("dir", aci.target).Warn("Cannot remove directory")
	}
}
