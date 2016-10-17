package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/blablacar/dgr/bin-templater/merger"
	"github.com/blablacar/dgr/bin-templater/template"
	"github.com/n0rad/go-erlog"
	"github.com/n0rad/go-erlog/logs"
	_ "github.com/n0rad/go-erlog/register"
	"github.com/peterbourgon/mergemap"
	"os"
)

const pathAttributes = "/attributes"
const pathTemplates = "/templates"
const usage = `Usage: templater [-c] [-e overrideEnvVarName] [-t target] templaterDir

  -c continue on error
  -o overrideEnvVarName,  varname of json object that will override attributes files
  -t target,  directory for start of templating instead of /`

func main() {
	logs.GetDefaultLog().(*erlog.ErlogLogger).Appenders[0].(*erlog.ErlogWriterAppender).Out = os.Stdout

	overrideEnvVarName := ""
	target := "/"
	var templateDir string
	logLvl := "INFO"
	continueOnError := false

	processArgs(&overrideEnvVarName, &target, &templateDir, &logLvl, &continueOnError)

	lvl, err := logs.ParseLevel(logLvl)
	if err != nil {
		fmt.Println("Wrong log level : " + logLvl)
		os.Exit(1)
	}
	logs.SetLevel(lvl)

	Run(overrideEnvVarName, target, templateDir, continueOnError)
}

func processArgs(overrideEnvVarName *string, target *string, templaterDir *string, logLevel *string, continueOnError *bool) {
	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--help":
		case "-h":
			fmt.Println(usage)
			os.Exit(1)
		case "-c":
			*continueOnError = true
		case "-o":
			*overrideEnvVarName = os.Args[i+1]
			i++
		case "-L":
			*logLevel = os.Args[i+1]
			i++
		case "-t":
			*target = os.Args[i+1]
			i++
		default:
			*templaterDir = os.Args[i]
		}
	}

	if *templaterDir == "" {
		fmt.Println("templaterDir is mandatory")
		fmt.Println()
		fmt.Println(usage)
		os.Exit(1)
	}
}

func Run(overrideEnvVarName string, target string, templaterDir string, continueOnError bool) {
	attrMerger, err := merger.NewAttributesMerger(templaterDir + pathAttributes)
	if err != nil {
		logs.WithE(err).Warn("Failed to prepare attributes")
	}
	attributes := attrMerger.Merge()
	attributes = overrideWithJsonIfNeeded(overrideEnvVarName, attributes)
	tt, err := merger.ProcessAttributesTemplating(attributes, attributes)
	attributes = tt.(map[string]interface{})
	if err != nil {
		logs.WithField("dir", templaterDir+pathTemplates).Fatal("Failed to template attributes")
	}
	logs.WithField("content", attributes).Debug("Final attributes resolution")

	info, _ := os.Stat(templaterDir + pathTemplates)
	if info == nil {
		logs.WithField("dir", templaterDir+pathTemplates).Debug("Template dir is empty. Nothing to template")
		return
	}
	tmpl, err := template.NewTemplateDir(templaterDir+pathTemplates, target, continueOnError)
	if err != nil {
		logs.WithE(err).WithField("dir", templaterDir+pathTemplates).Fatal("Failed to load template dir")
	}
	err = tmpl.Process(attributes)
	if err != nil {
		logs.WithE(err).WithField("dir", templaterDir+pathTemplates).Fatal("Failed to process template dir")
	}
}

func overrideWithJsonIfNeeded(overrideEnvVarName string, attributes map[string]interface{}) map[string]interface{} {
	if overrideEnvVarName != "" {
		if envjson := os.Getenv(overrideEnvVarName); envjson != "" {
			if len(envjson) > 8 && envjson[0:7] == "base64," {
				logs.WithField("EnvVar", overrideEnvVarName).Debug("Environment variable is base64 encoded")
				b64EnvJson := envjson[7:len(envjson)]
				envjsonBase64Decoded, err := base64.StdEncoding.DecodeString(b64EnvJson)
				if err != nil {
					logs.WithE(err).WithField("base64", b64EnvJson).Fatal("Failed to base64 decode")
				}
				envjson = string(envjsonBase64Decoded)
			}
			logs.WithField("content", envjson).Debug("Override var content")
			var envattr map[string]interface{}
			err := json.Unmarshal([]byte(envjson), &envattr)
			if err != nil {
				logs.WithE(err).
					WithField("varName", overrideEnvVarName).
					WithField("content", envjson).
					Fatal("Invalid format for environment override content")
			}
			attributes = mergemap.Merge(attributes, envattr)
		}
	}
	return attributes
}
