package main

import (
	"io/ioutil"
	"strconv"
	"strings"

	"os"

	"github.com/blablacar/dgr/dgr/common"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
)

var pathCp = "/cp"

func (aci *Aci) Cp(args []string) error {
	logs.WithF(aci.fields).Debug("Copy")
	if len(args) != 2 {
		return errs.With("cp requires 2 arguments")
	}
	if args[0][0] == '/' {
		args[0] = args[0][1:]
	}

	cpDir := aci.target + pathCp
	if err := os.MkdirAll(cpDir, 0755); err != nil {
		return errs.WithEF(err, data.WithField("dir", cpDir), "Failed to create directory")
	}
	defer func(dir string) {
		os.RemoveAll(dir)
	}(cpDir)

	stashes := strings.Count(args[0], "/")
	if err := tarExec(
		"xf",
		aci.target+pathImageAci,
		"--strip-components="+strconv.Itoa(stashes+1),
		"-C",
		cpDir,
		"rootfs/"+args[0],
	); err != nil {
		return errs.WithE(err, "Failed to extract files from aci")
	}

	giveBackUserRights(cpDir)

	files, err := ioutil.ReadDir(cpDir)
	if err != nil {
		return errs.WithEF(err, data.WithField("dir", cpDir), "Cannot read directory")
	}

	for _, f := range files {
		from := cpDir + "/" + f.Name()
		to := args[1] + "/" + f.Name()
		if err := common.ExecCmd("cp", "-R", "--preserve=all", from, to); err != nil {
			return errs.WithEF(err, data.WithField("from", from).
				WithField("to", to),
				"Failed to move file")
		}
	}

	return nil
}
