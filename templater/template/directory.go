package template

import (
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"os"
	"strings"
)

type templateDir struct {
	fields data.Fields
	src    string
	dst    string
}

func NewTemplateDir(path string, targetRoot string) *templateDir {
	fields := data.WithField("dir", path)
	logs.WithF(fields).Info("Reading template dir")
	return &templateDir{
		fields: fields,
		src:    path,
		dst:    targetRoot,
	}
}

func (t *templateDir) Process(attributes map[string]interface{}) error {
	if err := t.processSingleDir(t.src, t.dst, attributes); err != nil {
		return errs.WithEF(err, t.fields, "Failed to process templating of directory")
	}
	return nil
}

func (t *templateDir) processSingleDir(src string, dst string, attributes map[string]interface{}) error {
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return errs.WithEF(err, t.fields, "Cannot read source dir stat")
	}

	if err := os.MkdirAll(dst, sourceInfo.Mode()); err != nil {
		return errs.WithEF(err, t.fields, "Cannot create target directory templates")
	}

	directory, err := os.Open(src)
	if err != nil {
		return errs.WithEF(err, t.fields, "Cannot open directory")
	}
	objects, err := directory.Readdir(-1)
	if err != nil {
		return errs.WithEF(err, t.fields, "Cannot list files in directory")
	}
	for _, obj := range objects {
		srcObj := src + "/" + obj.Name()
		dstObj := dst + "/" + obj.Name()
		if obj.IsDir() {
			if err := t.processSingleDir(srcObj, dstObj, attributes); err != nil {
				return err
			}
		} else if strings.HasSuffix(obj.Name(), ".tmpl") {
			dstObj := dstObj[:len(dstObj)-5]
			template, err := NewTemplateFile(srcObj, obj.Mode())
			if err != nil {
				return err
			}
			if err := template.runTemplate(dstObj, attributes); err != nil {
				return errs.WithEF(err, t.fields, "File templating failed")
			}
		}
	}
	return nil
}
