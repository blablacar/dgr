package merger

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"text/template"

	tpl "github.com/blablacar/dgr/bin-templater/template"
	"github.com/ghodss/yaml"
	"github.com/leekchan/gtf"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"github.com/peterbourgon/mergemap"
)

type AttributesMerger struct {
	dir []string
}

func NewAttributesMerger(rootDir string,attributesDir string) (*AttributesMerger, error) {
	in := newInputs(rootDir + attributesDir)
	// initialize input files list)
	attrDir, err := os.Lstat(in.Directory)
	if err != nil {
		errs.WithEF(err, data.WithField("dir", in.Directory), "Cannot list files of dir")
	}
	err = in.addFiles(attrDir,rootDir)
	if err != nil {
		errs.WithEF(err, data.WithField("dir", in.Directory), "Cannot list files of dir")
	}

	res := []string{}
	for _, file := range in.Files {
		res = append(res,file)
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

func (in *inputs) addFiles(f_info os.FileInfo,cwd string) error{
	switch {
	case f_info == nil:
		return nil
	case f_info.Mode()&os.ModeSymlink == os.ModeSymlink :
		logs.WithField("files", f_info.Mode()).WithField("name",f_info.Name()).Trace("Checking symlink")
		followed_file, err := os.Readlink(cwd + "/" + f_info.Name())
		if err != nil {
		return err
		}
		if followed_file[0] != '/' {
			followed_file = cwd + "/" + followed_file
		} else {
			cwd = "/"
		}
		f_info, err = os.Lstat(followed_file)
		if err != nil {
		return err
		}
		in.addFiles(f_info,cwd + "/")
		logs.WithField("followed_link", f_info.Name()).Trace("Followed Link")
	case f_info.IsDir():
		list_f_info, err := ioutil.ReadDir(cwd + "/" + f_info.Name())
		if err != nil {
			return err
		}
		for _, f_info2 := range list_f_info {
			in.addFiles(f_info2,cwd + "/" + f_info.Name())
		}
	default:
		logs.WithField("files", f_info.Mode()).WithField("name",f_info.Name()).WithField("cwd",cwd).Trace("Adding a file")
		in.Files = append(in.Files, cwd + "/" + f_info.Name())
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
		if data == nil {
			continue
		}
		json := data.(map[string]interface{})
		omap = mergemap.Merge(newMap, json)
	}
	data := ProcessOverride(newMap)
	return data
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
	} else if omap["default"] == nil {
		omap = make(map[string]interface{})
	} else {
		omap = omap["default"].(map[string]interface{})
	}
	return omap
}

func Merge(envName string, files []string) []byte {
	// inputDir string,
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

func ProcessAttributesTemplating(in interface{}, attributes interface{}) (_ interface{}, err error) {
	switch in.(type) {
	case map[string]interface{}:
		o := make(map[string]interface{})
		for k, v := range in.(map[string]interface{}) {
			v, err = ProcessAttributesTemplating(v, attributes)
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
			o[i], err = ProcessAttributesTemplating(in1[i], attributes)
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
		if templated != in.(string) {
			var rr interface{}
			err := yaml.Unmarshal([]byte(templated), &rr)
			if err != nil {
				return nil, err
			}
			return rr, nil
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
