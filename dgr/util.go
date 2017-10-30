package main

import (
	"github.com/blablacar/dgr/dgr/common"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
)

func tarExec(args ...string) error {
	out, stderr, err := common.ExecCmdGetStdoutAndStderr("tar", args...)
	if err != nil {
		return errs.WithEF(err, data.
			WithField("stdout", out).
			WithField("stderr", stderr), "Tar update failed")
	}
	if logs.IsDebugEnabled() {
		println(out)
	}
	return nil
}
