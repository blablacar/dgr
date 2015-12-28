package attributes

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/peterbourgon/mergemap"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strconv"
)

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
	return ProcessOverride(newMap)
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
		omap = mergemap.Merge(omap, envattr)
	}

	// map to json
	out, err := json.Marshal(omap)
	if err != nil {
		panic(err)
	}

	return out
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
	return in, nil
}
