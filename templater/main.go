package main

import (
	"encoding/json"
	"fmt"
	"github.com/blablacar/cnt/templater/merger"
	"github.com/blablacar/cnt/templater/template"
	"github.com/n0rad/go-erlog"
	"github.com/n0rad/go-erlog/logs"
	_ "github.com/n0rad/go-erlog/register"
	"github.com/peterbourgon/mergemap"
	"os"
)

func main() {
	logs.GetDefaultLog().(*erlog.ErlogLogger).Appenders[0].(*erlog.ErlogWriterAppender).Out = os.Stdout
	logs.SetLevel(logs.TRACE)

	overrideEnvVarName := ""
	target := "/"
	var templateDir string

	processArgs(&overrideEnvVarName, &target, &templateDir)

	Run(overrideEnvVarName, target, templateDir)
}

const USAGE = `Usage: templater [-e overrideEnvVarName] [-t target] templaterDir

  -o overrideEnvVarName,  varname of json object that will override attributes files
  -t target,  directory for start of templating instead of /`

func processArgs(overrideEnvVarName *string, target *string, templaterDir *string) {
	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--help":
		case "-h":
			fmt.Println(USAGE)
			os.Exit(1)
		case "-o":
			*overrideEnvVarName = os.Args[i+1]
			i++
		case "-t":
			*target = os.Args[i+1]
			i++
		default:
			*templaterDir = os.Args[i]
		}
	}

	if *templaterDir == "" {
		fmt.Println("templaterDir is mandatory\n")
		fmt.Println(USAGE)
		os.Exit(1)
	}
}

const PATH_ATTRIBUTES = "/attributes"
const PATH_TEMPLATES = "/templates"

func Run(overrideEnvVarName string, target string, templaterDir string) {
	attrMerger, err := merger.NewAttributesMerger(templaterDir + PATH_ATTRIBUTES)
	if err != nil {
		logs.WithE(err).Warn("Failed to prepare attributes")
	}
	attributes := attrMerger.Merge()
	attributes = overrideWithJsonIfNeeded(overrideEnvVarName, attributes)

	err = template.NewTemplateDir(templaterDir+PATH_TEMPLATES, target).Process(attributes)
	if err != nil {
		logs.WithE(err).WithField("dir", templaterDir+PATH_TEMPLATES).Fatal("Failed to process template dir")
	}

	//		out := attrMerger.Merge(fgs.confd_env, res)
	//	err = ioutil.WriteFile("attributes.json", []byte(out), 0777)
	//	if err != nil {
	//		panic(err)
	//	}

	// read files from dir
	// run templates
}

func overrideWithJsonIfNeeded(overrideEnvVarName string, attributes map[string]interface{}) map[string]interface{} {
	if overrideEnvVarName != "" {
		if envjson := os.Getenv(overrideEnvVarName); envjson != "" {
			var envattr map[string]interface{}
			err := json.Unmarshal([]byte(envjson), &envattr)
			if err != nil {
				panic(err)
			}
			attributes = mergemap.Merge(attributes, envattr)
		}
	}
	return attributes
}
