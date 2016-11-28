package main

import (
	"bytes"
	"io/ioutil"
	"os"

	"github.com/blablacar/dgr/dgr/common"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
)

func (aci *Aci) Graph() error {
	defer aci.giveBackUserRightsToTarget()
	logs.WithF(aci.fields).Debug("Graphing")

	os.MkdirAll(aci.target, 0777)

	var buffer bytes.Buffer
	buffer.WriteString("digraph {\n")

	for _, dep := range aci.manifest.Aci.Dependencies {
		buffer.WriteString("  ")
		buffer.WriteString("\"")
		buffer.WriteString(dep.TinyNameId())
		buffer.WriteString("\"")
		buffer.WriteString(" -> ")
		buffer.WriteString("\"")
		buffer.WriteString(aci.manifest.NameAndVersion.TinyNameId())
		buffer.WriteString("\"")
		buffer.WriteString("\n")
	}

	for _, dep := range aci.manifest.Builder.Dependencies {
		buffer.WriteString("  ")
		buffer.WriteString("\"")
		buffer.WriteString(dep.TinyNameId())
		buffer.WriteString("\"")
		buffer.WriteString(" -> ")
		buffer.WriteString("\"")
		buffer.WriteString(aci.manifest.NameAndVersion.TinyNameId())
		buffer.WriteString("\"")
		buffer.WriteString("[color=red,penwidth=2.0]")
		buffer.WriteString("\n")
	}

	for _, dep := range aci.manifest.Tester.Builder.Dependencies {
		buffer.WriteString("  ")
		buffer.WriteString("\"")
		buffer.WriteString(dep.TinyNameId())
		buffer.WriteString("\"")
		buffer.WriteString(" -> ")
		buffer.WriteString("\"")
		buffer.WriteString(aci.manifest.NameAndVersion.TinyNameId())
		buffer.WriteString("\"")
		buffer.WriteString("[color=green,penwidth=2.0]")
		buffer.WriteString("\n")
	}

	for _, dep := range aci.manifest.Tester.Aci.Dependencies {
		buffer.WriteString("  ")
		buffer.WriteString("\"")
		buffer.WriteString(dep.TinyNameId())
		buffer.WriteString("\"")
		buffer.WriteString(" -> ")
		buffer.WriteString("\"")
		buffer.WriteString(aci.manifest.NameAndVersion.TinyNameId())
		buffer.WriteString("\"")
		buffer.WriteString("[color=blue,penwidth=2.0]")
		buffer.WriteString("\n")
	}

	buffer.WriteString("}\n")

	if err := ioutil.WriteFile(aci.target+pathGraphDot, buffer.Bytes(), 0644); err != nil {
		return errs.WithEF(err, aci.fields.WithField("file", aci.target+pathGraphDot), "Failed to write file")
	}

	if _, _, err := common.ExecCmdGetStdoutAndStderr("dot", "-V"); err == nil {
		if std, stderr, err := common.ExecCmdGetStdoutAndStderr("dot", "-Tpng", aci.target+pathGraphDot, "-o", aci.target+pathGraphPng); err != nil {
			return errs.WithEF(err, aci.fields.WithField("stdout", std).WithField("stderr", stderr), "Failed to create graph image")
		}
	}

	return nil
}
