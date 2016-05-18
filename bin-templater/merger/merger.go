package merger

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"github.com/peterbourgon/mergemap"
	"io/ioutil"
	"os"
	"strconv"
	"text/template"
	tpl "github.com/blablacar/dgr/bin-templater/template"
	"github.com/leekchan/gtf"
)

type AttributesMerger struct {
	dir []string
}

func NewAttributesMerger(path string) (*AttributesMerger, error) {
	in := newInputs(path)
	// initialize input files list
	err := in.listFiles()
	if err != nil {
		errs.WithEF(err, data.WithField("dir", path), "Cannot list files of dir")
	}

	res := []string{}
	for _, file := range in.Files {
		res = append(res, in.Directory+file)
	}
	return &AttributesMerger{dir: res}, nil
}

func (a AttributesMerger) Merge() map[string]interface{} {
	return MergeAttributesFiles(a.dir)
}

type inputs struct {
	Directory string
	Files     []string
}

// constructor
func newInputs(d string) *inputs {
	in := new(inputs)
	in.Directory = d + "/"
	return in
}

// list input files
func (in *inputs) listFiles() error {
	list_l1, err := ioutil.ReadDir(in.Directory)
	if err != nil {
		return err
	}
	for _, f_l1 := range list_l1 {
		if f_l1.IsDir() {
			list_l2, err := ioutil.ReadDir(in.Directory + "/" + f_l1.Name())
			if err != nil {
				return err
			}
			for _, f_l2 := range list_l2 {
				in.Files = append(in.Files, f_l1.Name()+"/"+f_l2.Name())
			}
		} else {
			in.Files = append(in.Files, f_l1.Name())
		}
	}
	return nil
}

func MergeAttributesFilesForMap(omap map[string]interface{}, files []string) map[string]interface{} {

	newMap := make(map[string]interface{})
	newMap["default"] = omap

	// loop over attributes files
	// merge override files to default files
	for _, file := range files {
		var data interface{}
		yml, err := ioutil.ReadFile(file)
		if err != nil {
			panic(err)
		}
		// yaml to data
		err = yaml.Unmarshal(yml, &data)
		if err != nil {
			panic(err)
		}
		data, err = transform(data)
		if err != nil {
			panic(err)
		}
		// data to map
		json := data.(map[string]interface{})
		omap = mergemap.Merge(newMap, json)
	}
	data := ProcessOverride(newMap)

	data2, err := processAttributesTemplating(data, data)
	if err != nil {
		panic(err)
	}
	return data2.(map[string]interface{})
}

func MergeAttributesFiles(files []string) map[string]interface{} {
	omap := make(map[string]interface{})
	return MergeAttributesFilesForMap(omap, files)
}

func ProcessOverride(omap map[string]interface{}) map[string]interface{} {
	// merge override to default inside the file
	_, okd := omap["default"]
	if okd == false {
		omap["default"] = make(map[string]interface{}) //init if default doesn't exist
	}
	_, oko := omap["override"]
	if oko == true {
		omap = mergemap.Merge(omap["default"].(map[string]interface{}), omap["override"].(map[string]interface{}))
	} else {
		omap = omap["default"].(map[string]interface{})
	}
	return omap
}

func Merge(envName string, files []string) []byte { // inputDir string,
	// "out map" to store merged yamls
	omap := MergeAttributesFiles(files)

	envjson := os.Getenv(envName)
	if envjson != "" {
		var envattr map[string]interface{}
		err := json.Unmarshal([]byte(envjson), &envattr)
		if err != nil {
			panic(err)
		}
		mergemap.Merge(omap, envattr)
	}

	// map to json
	out, err := json.Marshal(omap)
	if err != nil {
		panic(err)
	}

	return out
}

func processAttributesTemplating(in interface{}, attributes interface{}) (_ interface{}, err error) {
	switch in.(type) {
	case map[string]interface{}:
		o := make(map[string]interface{})
		for k, v := range in.(map[string]interface{}) {
			v, err = processAttributesTemplating(v, attributes)
			if err != nil {
				return nil, err
			}
			o[k] = v
		}
		return o, nil
	case []interface{}:
		in1 := in.([]interface{})
		len1 := len(in1)
		o := make([]interface{}, len1)
		for i := 0; i < len1; i++ {
			o[i], err = processAttributesTemplating(in1[i], attributes)
			if err != nil {
				return nil, err
			}
		}
		return o, nil
	case string:
		templated, err := templateAttribute(in.(string), attributes)
		if err != nil {
			return nil, err
		}
		return templated, nil
	default:
		return in, nil
	}
}

func templateAttribute(text string, attributes interface{}) (string, error) {
	tmpl, err := template.New("").Funcs(tpl.TemplateFunctions).Funcs(map[string]interface{}(gtf.GtfFuncMap)).Parse(text)
	if err != nil {
		return "", errs.WithEF(err, data.WithField("attribute", text), "Failed to parse template for attribute")
	}
	var b bytes.Buffer
	if err := tmpl.Execute(&b, attributes); err != nil {
		return "", errs.WithEF(err, data.WithField("attribute", text), "Failed to template attribute")
	}
	res := b.String()
	if logs.IsDebugEnabled() {
		logs.WithField("from", text).WithField("to", res).Debug("attribute templated")
	}
	return res, nil
}

// transform YAML to JSON
func transform(in interface{}) (_ interface{}, err error) {
	switch in.(type) {
	case map[interface{}]interface{}:
		o := make(map[string]interface{})
		for k, v := range in.(map[interface{}]interface{}) {
			sk := ""
			switch k.(type) {
			case string:
				sk = k.(string)
			case int:
				sk = strconv.Itoa(k.(int))
			default:
				return nil, errors.New(
					fmt.Sprintf("type not match: expect map key string or int get: %T", k))
			}
			v, err = transform(v)
			if err != nil {
				return nil, err
			}
			o[sk] = v
		}
		return o, nil
	case []interface{}:
		in1 := in.([]interface{})
		len1 := len(in1)
		o := make([]interface{}, len1)
		for i := 0; i < len1; i++ {
			o[i], err = transform(in1[i])
			if err != nil {
				return nil, err
			}
		}
		return o, nil
	default:
		return in, nil
	}
}
