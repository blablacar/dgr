package template

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	txttmpl "text/template"

	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"gopkg.in/yaml.v2"
)

type TemplateFile struct {
	Uid      int    `yaml:"uid"`
	Gid      int    `yaml:"gid"`
	CheckCmd string `yaml:"checkCmd"`

	fields   data.Fields
	Mode     os.FileMode
	template *Templating
}

func NewTemplateFile(partials *txttmpl.Template, src string, mode os.FileMode) (*TemplateFile, error) {
	fields := data.WithField("src", src)

	content, err := ioutil.ReadFile(src)
	if err != nil {
		return nil, errs.WithEF(err, fields, "Cannot read template file")
	}

	template, err := NewTemplating(partials, src, string(content))
	if err != nil {
		return nil, errs.WithEF(err, fields, "Failed to prepare template")
	}

	t := &TemplateFile{
		Uid:      os.Getuid(),
		Gid:      os.Getgid(),
		fields:   fields,
		template: template,
		Mode:     mode,
	}
	err = t.loadTemplateConfig(src)
	logs.WithF(fields).WithField("data", t).Trace("Template loaded")
	return t, err
}

func (t *TemplateFile) loadTemplateConfig(src string) error {
	cfgPath := src + EXT_CFG
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return nil
	}

	source, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal([]byte(source), t)
	if err != nil {
		return errs.WithEF(err, data.WithField("name", src), "Cannot unmarshall cfg")
	}
	return nil
}

func (f *TemplateFile) runTemplate(dst string, attributes map[string]interface{}, failOnNoValue bool) error {
	if logs.IsTraceEnabled() {
		logs.WithF(f.fields).WithField("attributes", attributes).WithField("failOnNoValue", failOnNoValue).Trace("templating with attributes")
	}
	fields := f.fields.WithField("dst", dst)

	logs.WithF(fields).Info("Templating file")

	out, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, f.Mode)
	if err != nil {
		return errs.WithEF(err, fields, "Cannot open destination file")
	}
	defer func() {
		out.Close()
	}()

	buff := bytes.Buffer{}
	writer := bufio.NewWriter(&buff)
	if err := f.template.Execute(writer, attributes); err != nil {
		return errs.WithEF(err, fields, "Templating execution failed")
	}

	if err := writer.Flush(); err != nil {
		return errs.WithEF(err, fields, "Failed to flush buffer")
	}
	buff.WriteByte('\n')

	b := buff.Bytes()

	if logs.IsTraceEnabled() {
		logs.WithF(f.fields).WithField("result", string(b)).Trace("templating done")
	}

	scanner := bufio.NewScanner(bytes.NewReader(b)) // TODO this sux
	scanner.Split(bufio.ScanLines)
	for i := 1; scanner.Scan(); i++ {
		text := scanner.Text()
		if bytes.Contains([]byte(text), []byte("<no value>")) {
			err = errs.WithF(fields.WithField("line", i).WithField("text", text), "Templating result have <no value>")
			if failOnNoValue {
				return err
			} else {
				logs.WithE(err).Error("Templating result have <no value>")
			}
		}
	}

	if length, err := out.Write(b); length != len(b) || err != nil {
		return errs.WithEF(err, fields, "Write to file failed")
	}

	if err = out.Sync(); err != nil {
		return errs.WithEF(err, fields, "Failed to sync output file")
	}
	if err = os.Chmod(dst, f.Mode); err != nil {
		return errs.WithEF(err, fields.WithField("file", dst), "Failed to set mode on file")
	}
	if err = os.Chown(dst, f.Uid, f.Gid); err != nil {
		return errs.WithEF(err, fields.WithField("file", dst), "Failed to set owner of file")
	}

	if f.CheckCmd != "" {
		cmd := exec.Command("/dgr/bin/busybox", "sh", "-c", f.CheckCmd)
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		if err = cmd.Run(); err != nil {
			return errs.WithEF(err, fields.WithField("file", dst), "Check command failed after templating")
		}
	}
	return err
}
