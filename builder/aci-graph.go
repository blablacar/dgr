package builder

import (
	"bytes"
	"github.com/n0rad/go-erlog/logs"
	"io/ioutil"
	"os"
	"os/exec"
)

func (aci *Aci) Graph() {
	logs.WithF(aci.fields).Info("Graphing")

	os.MkdirAll(aci.target, 0777)

	var buffer bytes.Buffer
	buffer.WriteString("digraph {\n")

	froms, err := aci.manifest.GetFroms()
	if err != nil {
		logs.WithEF(err, aci.fields).Fatal("Cannot render from part of grapth")
	}
	for _, from := range froms {
		if from == "" {
			continue
		}
		buffer.WriteString("  ")
		buffer.WriteString("\"")
		buffer.WriteString(from.ShortNameId())
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
	cmd := "dot -V"
	if err := exec.Command(cmd).Run(); err != nil {
		_, err := os.Stat(aci.target + "/graph.dot")
		if os.IsNotExist(err) {
			logs.WithF(aci.fields).Error("No such file : " + aci.target + "/graph.dot")
			return
		} else {
			cmd := exec.Command("dot", "-Tpng", aci.target+"/graph.dot", "-o", aci.target+"/graph.png")
			cmd.Run()
		}
	}
}
