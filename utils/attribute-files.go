package utils

import (
	"github.com/blablacar/attributes-merger/attributes"
	"github.com/google/cadvisor/utils"
	"strings"
)

func AttributeFiles(path string) ([]string, error) {
	res := []string{}
	if !utils.FileExists(path) {
		return res, nil
	}

	in := attributes.NewInputs(path)
	// initialize input files list
	err := in.ListFiles()
	if err != nil {
		return nil, err
	}

	for _, file := range in.Files {
		if strings.HasSuffix(file, ".yml") || strings.HasSuffix(file, ".yaml") {
			res = append(res, in.Directory+file)
		}
	}
	return res, nil
}
