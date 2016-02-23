package yop

import (
	"github.com/blablacar/attributes-merger/attributes"
	"os"
	"strings"
)

func AttributeFiles(path string) ([]string, error) {
	res := []string{}
	if _, err := os.Stat(path); err != nil {
		return res, nil
	}

	in := attributes.NewInputs(path) // TODO remove attribute merger
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
