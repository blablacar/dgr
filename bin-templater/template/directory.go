package template

import (
	"github.com/leekchan/gtf"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"io/ioutil"
	"os"
	"strings"
	txttmpl "text/template"
)

type TemplateDir struct {
	continueOnError bool
	fields          data.Fields
	src             string
	dst             string
	Partials        *txttmpl.Template
}

func NewTemplateDir(path string, targetRoot string, continueOnError bool) (*TemplateDir, error) {
	fields := data.WithField("dir", path).WithField("continueOnError", continueOnError)
	logs.WithF(fields).Debug("Reading template dir")
	tmplDir := &TemplateDir{
		fields:          fields,
		src:             path,
		dst:             targetRoot,
		continueOnError: continueOnError,
	}
	return tmplDir, tmplDir.LoadPartial()
}

func (t *TemplateDir) LoadPartial() error {
	partials := []string{}

	directory, err := os.Open(t.src)
	if err != nil {
		return errs.WithEF(err, t.fields, "Failed to open template dir")
	}
	objects, err := directory.Readdir(-1)
	if err != nil {
		return errs.WithEF(err, t.fields, "Failed to read template dir")
	}
	for _, obj := range objects {
		if !obj.IsDir() && strings.HasSuffix(obj.Name(), ".partial") {
			partials = append(partials, t.src+"/"+obj.Name())
		}
	}

	if len(partials) == 0 {
		return nil
	}
	var tmpl *txttmpl.Template
	for _, partial := range partials {
		if tmpl == nil {
			tmpl = txttmpl.New(partial).Funcs(TemplateFunctions).Funcs(map[string]interface{}(gtf.GtfFuncMap))
		} else {
			tmpl = tmpl.New(partial).Funcs(TemplateFunctions).Funcs(map[string]interface{}(gtf.GtfFuncMap))
		}

		content, err := ioutil.ReadFile(partial)
		if err != nil {
			return errs.WithEF(err, t.fields.WithField("partial", partial), "Cannot read partial file")
		}
		tmpl, err = tmpl.Funcs(TemplateFunctions).Parse(CleanupOfTemplate(string(content)))
		if err != nil {
			return errs.WithEF(err, t.fields.WithField("partial", partial), "Failed to parse partial")
		}
	}
	t.Partials = tmpl
	return nil
}

func (t *TemplateDir) Process(attributes map[string]interface{}) error {
	if err := t.processSingleDir(t.src, t.dst, attributes); err != nil {
		return err
	}
	return nil
}

func (t *TemplateDir) processSingleDir(src string, dst string, attributes map[string]interface{}) error {
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
		} else if strings.HasSuffix(obj.Name(), ".tmpl") || (strings.Contains(obj.Name(), ".tmpl.") && (!strings.HasSuffix(obj.Name(), ".cfg"))) {
			if strings.HasSuffix(obj.Name(), ".tmpl") {
				dstObj = dstObj[:len(dstObj)-5]
			} else {
				dstObj = dst + "/" + strings.Replace(obj.Name(), ".tmpl.", ".", 1)
			}
			template, err := NewTemplateFile(t.Partials, srcObj, obj.Mode())
			if err != nil {
				return err
			}
			if err2 := template.runTemplate(dstObj, attributes, !t.continueOnError); err2 != nil {
				if t.continueOnError {
					err = err2
					logs.WithEF(err, t.fields).Error("Templating failed")
				} else {
					return err2
				}
			}
		}
	}
	return err
}
