package common

const PATH_IMAGE_ACI = "/image.aci"
const PATH_MANIFEST = "/manifest"
const PATH_ROOTFS = "/rootfs"

const ENV_ACI_PATH = "ACI_PATH"
const ENV_ACI_TARGET = "ACI_TARGET"
const ENV_LOG_LEVEL = "LOG_LEVEL"

const ENV_BUILDER_COMMAND = "BUILDER_COMMAND"
const PREFIX_BUILDER = "builder/"

type BuilderCommand string

const (
	COMMAND_BUILD BuilderCommand = "build"
	COMMAND_INIT  BuilderCommand = "init"
)
