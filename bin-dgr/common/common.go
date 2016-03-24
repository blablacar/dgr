package common

import (
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
)

const PATH_IMAGE_ACI = "/image.aci"
const PATH_MANIFEST = "/manifest"
const PATH_ROOTFS = "/rootfs"

const ENV_ACI_PATH = "ACI_PATH"
const ENV_ACI_TARGET = "ACI_TARGET"
const ENV_LOG_LEVEL = "LOG_LEVEL"
const ENV_TRAP_ON_ERROR = "TRAP_ON_ERROR"

const ENV_BUILDER_COMMAND = "BUILDER_COMMAND"
const PREFIX_BUILDER = "builder/"

type BuilderCommand string

const (
	COMMAND_BUILD BuilderCommand = "build"
	COMMAND_INIT  BuilderCommand = "init"
	COMMAND_TRY   BuilderCommand = "try"
)

func (b BuilderCommand) CommandManifestKey() (string, error) {
	switch b {
	case COMMAND_BUILD:
		return "dgrtool.com/dgr/stage1/build", nil
	case COMMAND_INIT:
		return "dgrtool.com/dgr/stage1/init", nil
	case COMMAND_TRY:
		return "dgrtool.com/dgr/stage1/try", nil
	default:
		return "", errs.WithF(data.WithField("command", b), "Unimplemented command manifest key")
	}
}
