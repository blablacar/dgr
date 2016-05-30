package main

import (
	"bytes"
	"github.com/blablacar/dgr/bin-dgr/common"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"io/ioutil"
	"os"
)

func (p *Pod) Graph() error {
	logs.WithF(p.fields).Info("Graphing")
	os.MkdirAll(p.target, 0777)

	var buffer bytes.Buffer
	buffer.WriteString("digraph {\n")
	buffer.WriteString("  {\n")
	buffer.WriteString("  ")
	buffer.WriteString("\"")
	buffer.WriteString(p.manifest.Name.TinyNameId())
	buffer.WriteString("\"")
	buffer.WriteString(" [style=filled, fillcolor=yellow, shape=box]\n")
	buffer.WriteString("  }\n")

	for _, e := range p.manifest.Pod.Apps {
		for _, d := range e.Dependencies {
			buffer.WriteString("  ")
			buffer.WriteString("\"")
			buffer.WriteString(d.TinyNameId())
			buffer.WriteString("\"")
			buffer.WriteString(" -> ")
			buffer.WriteString("\"")
			buffer.WriteString(p.manifest.Name.TinyNameId())
			buffer.WriteString("\"")
			buffer.WriteString("\n")
		}
	}

	buffer.WriteString("}\n")

	if err := ioutil.WriteFile(p.target+pathGraphDot, buffer.Bytes(), 0644); err != nil {
		return errs.WithEF(err, p.fields.WithField("file", p.target+pathGraphDot), "Failed to write file")
	}

	if _, _, err := common.ExecCmdGetStdoutAndStderr("dot", "-V"); err == nil {
		if std, stderr, err := common.ExecCmdGetStdoutAndStderr("dot", "-Tpng", p.target+pathGraphDot, "-o", p.target+pathGraphPng); err != nil {
			return errs.WithEF(err, p.fields.WithField("stdout", std).WithField("stderr", stderr), "Failed to create graph image")
		}
	}

	return nil

}
