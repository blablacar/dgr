package template

import (
	"bufio"
	"bytes"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"io/ioutil"
	"os"
)

type TemplateFile struct {
	fields   data.Fields
	srcMode  os.FileMode
	template *Templating
}

func NewTemplateFile(src string, mode os.FileMode) (*TemplateFile, error) {
	fields := data.WithField("src", src)

	content, err := ioutil.ReadFile(src)
	if err != nil {
		return nil, errs.WithEF(err, fields, "Cannot read template file")
	}

	template := NewTemplating(src, string(content))

	if err := template.ParseWithSuccess(); err != nil {
		return nil, errs.WithEF(err, fields, "Failed to parse template file")
	}

	return &TemplateFile{
		fields:   fields,
		template: template,
		srcMode:  mode,
	}, nil
}

func (f *TemplateFile) runTemplate(dst string, attributes map[string]interface{}) error {
	if logs.IsTraceEnabled() {
		logs.WithF(f.fields).WithField("attributes", attributes).Trace("templating with attributes")
	}
	fields := f.fields.WithField("dst", dst)

	logs.WithF(fields).Info("Templating file")

	out, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, f.srcMode)
	if err != nil {
		return errs.WithEF(err, fields, "Cannot open destination file")
	}
	defer func() { out.Close() }()

	buff := bytes.Buffer{}
	writer := bufio.NewWriter(&buff)
	if err := f.template.Execute(writer, attributes); err != nil {
		return errs.WithEF(err, fields, "Templating execution failed")
	}

	if err := writer.Flush(); err != nil {
		return errs.WithEF(err, fields, "Failed to flush buffer")
	}

	b := buff.Bytes()
	if logs.IsTraceEnabled() {
		logs.WithF(f.fields).WithField("result", string(b)).Trace("templating done")
	}
	if bytes.Contains(b, []byte("<no value>")) {
		return errs.WithF(fields, "Templating result have <no value>")
	}
	out.Write(b)

	if err = out.Sync(); err != nil {
		return errs.WithEF(err, fields, "Failed to sync output file")
	}
	return nil
}
