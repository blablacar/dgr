package builder

import (
	"bytes"
	"github.com/n0rad/go-erlog/logs"
	"io/ioutil"
	"os"
	"os/exec"
)

func (aci *Aci) Graph() {
	logs.WithF(aci.fields).Debug("Graphing")

	os.MkdirAll(aci.target, 0777)

	var buffer bytes.Buffer
	buffer.WriteString("digraph {\n")

	if aci.manifest.From != "" {
		buffer.WriteString("  ")
		buffer.WriteString("\"")
		buffer.WriteString(aci.manifest.From.ShortNameId())
		buffer.WriteString("\"")
		buffer.WriteString(" -> ")
		buffer.WriteString("\"")
		buffer.WriteString(aci.manifest.NameAndVersion.ShortNameId())
		buffer.WriteString("\"")
		buffer.WriteString("[color=red,penwidth=2.0]")
		buffer.WriteString("\n")
	}

	for _, dep := range aci.manifest.Aci.Dependencies {
		buffer.WriteString("  ")
		buffer.WriteString("\"")
		buffer.WriteString(dep.ShortNameId())
		buffer.WriteString("\"")
		buffer.WriteString(" -> ")
		buffer.WriteString("\"")
		buffer.WriteString(aci.manifest.NameAndVersion.ShortNameId())
		buffer.WriteString("\"")
		buffer.WriteString("\n")
	}

	buffer.WriteString("}\n")

	ioutil.WriteFile(aci.target+"/graph.dot", buffer.Bytes(), 0644)
	cmd := exec.Command("dot", "-Tpng", aci.target+"/graph.dot", "-o", aci.target+"/graph.png")
	cmd.Run()
}
