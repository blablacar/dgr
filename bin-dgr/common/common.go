package common

import (
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
)

const PathImageAci = "/image.aci"
const PathManifest = "/manifest"
const PathRootfs = "/rootfs"
const PathAciManifest = "/aci-manifest.yml"
const PathManifestYmlTmpl = "/aci-manifest.yml.tmpl"

const EnvDgrVersion = "DGR_VERSION"
const EnvAciPath = "ACI_PATH"
const EnvAciTarget = "ACI_TARGET"
const EnvLogLevel = "LOG_LEVEL"
const EnvTrapOnError = "TRAP_ON_ERROR"
const EnvTrapOnStep = "TRAP_ON_STEP"

const EnvBuilderCommand = "BUILDER_COMMAND"
const PrefixBuilder = "builder/"

const ManifestDrgVersion = "dgr-version"

type BuilderCommand string

const (
	CommandBuild BuilderCommand = "build"
	CommandInit  BuilderCommand = "init"
	CommandTry   BuilderCommand = "try"
)

func (b BuilderCommand) CommandManifestKey() (string, error) {
	switch b {
	case CommandBuild:
		return "dgrtool.com/dgr/stage1/build", nil
	case CommandInit:
		return "dgrtool.com/dgr/stage1/init", nil
	case CommandTry:
		return "dgrtool.com/dgr/stage1/try", nil
	default:
		return "", errs.WithF(data.WithField("command", b), "Unimplemented command manifest key")
	}
}
