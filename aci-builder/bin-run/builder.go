package main

import (
	"encoding/json"
	"fmt"
	"github.com/appc/spec/schema"
	"github.com/appc/spec/schema/types"
	"github.com/blablacar/dgr/bin-dgr/common"
	rktcommon "github.com/coreos/rkt/common"
	"github.com/coreos/rkt/pkg/sys"
	stage1commontypes "github.com/coreos/rkt/stage1/common/types"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Builder struct {
	fields        data.Fields
	stage1Rootfs  string
	stage2Rootfs  string
	aciHomePath   string
	aciTargetPath string
	upperId       string
	pod           *stage1commontypes.Pod
}

func NewBuilder(podRoot string, podUUID *types.UUID) (*Builder, error) {
	pod, err := stage1commontypes.LoadPod(podRoot, podUUID)
	if err != nil {
		logs.WithError(err).Fatal("Failed to load pod")
	}
	if len(pod.Manifest.Apps) != 1 {
		logs.Fatal("dgr builder support only 1 application")
	}

	fields := data.WithField("aci", manifestApp(pod).Name)
	logs.WithF(fields).WithField("path", pod.Root).Info("Loading aci builder")

	aciPath, ok := manifestApp(pod).App.Environment.Get(common.EnvAciPath)
	if !ok || aciPath == "" {
		return nil, errs.WithF(fields, "Builder image require "+common.EnvAciPath+" environment variable")
	}
	aciTarget, ok := manifestApp(pod).App.Environment.Get(common.EnvAciTarget)
	if !ok || aciPath == "" {
		return nil, errs.WithF(fields, "Builder image require "+common.EnvAciTarget+" environment variable")
	}

	return &Builder{
		fields:        fields,
		aciHomePath:   aciPath,
		aciTargetPath: aciTarget,
		pod:           pod,
		stage1Rootfs:  rktcommon.Stage1RootfsPath(pod.Root),
		stage2Rootfs:  filepath.Join(rktcommon.AppPath(pod.Root, manifestApp(pod).Name), "rootfs"),
	}, nil
}

func (b *Builder) Build() error {
	logs.WithF(b.fields).Info("Building aci")

	lfd, err := rktcommon.GetRktLockFD()
	if err != nil {
		return errs.WithEF(err, b.fields, "can't get rkt lock fd")
	}

	if err := sys.CloseOnExec(lfd, true); err != nil {
		return errs.WithEF(err, b.fields, "can't set FD_CLOEXEC on rkt lock")
	}

	if err := b.runBuildSetup(); err != nil {
		return err
	}

	if err := b.runBuild(); err != nil {
		return err
	}

	if err := b.writeManifest(); err != nil {
		return err
	}

	if err := b.tarAci(); err != nil {
		return err
	}

	return nil
}

////////////////////////////////////////////

func (b *Builder) writeManifest() error {
	upperId, err := b.upperTreeStoreId()
	if err != nil {
		return err
	}

	manifestPath := b.stage1Rootfs + PATH_OPT + PATH_STAGE2 + "/" + manifestApp(b.pod).Name.String() + common.PathManifest
	content, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return errs.WithEF(err, b.fields.WithField("file", manifestPath), "Failed to read manifest file")
	}
	im := &schema.ImageManifest{}
	err = im.UnmarshalJSON(content)
	if err != nil {
		return errs.WithEF(err, data.WithField("content", string(content)), "Cannot unmarshall json content")
	}

	im.Name.Set(strings.Replace(im.Name.String(), common.PrefixBuilder, "", 1))
	if _, ok := im.Labels.Get("version"); !ok {
		os.Chdir(b.aciHomePath)
		im.Labels = append(im.Labels, types.Label{Name: "version", Value: common.GenerateVersion(b.aciHomePath)})
	}
	if content, err := json.MarshalIndent(im, "", "  "); err != nil {
		return errs.WithEF(err, b.fields, "Failed to write manifest")
	} else if err := ioutil.WriteFile(b.pod.Root+PATH_OVERLAY+"/"+upperId+PATH_UPPER+common.PathManifest, content, 0644); err != nil {
		return errs.WithEF(err, b.fields, "Failed to write manifest")
	}
	return nil
}

func (b *Builder) tarAci() error {
	upperId, err := b.upperTreeStoreId()
	if err != nil {
		return err
	}

	upperPath := b.pod.Root + PATH_OVERLAY + "/" + upperId + PATH_UPPER
	upperNamedRootfs := upperPath + "/" + manifestApp(b.pod).Name.String()
	upperRootfs := upperPath + common.PathRootfs

	if err := os.Rename(upperNamedRootfs, upperRootfs); err != nil { // TODO this is dirty and can probably be renamed during tar
		return errs.WithEF(err, b.fields.WithField("path", upperNamedRootfs), "Failed to rename rootfs")
	}
	defer os.Rename(upperRootfs, upperNamedRootfs)

	dir, err := os.Getwd()
	if err != nil {
		return errs.WithEF(err, b.fields, "Failed to get current working directory")
	}
	defer func() {
		if err := os.Chdir(dir); err != nil {
			logs.WithEF(err, b.fields.WithField("path", dir)).Warn("Failed to chdir back")
		}
	}()

	if err := os.Chdir(upperPath); err != nil {
		return errs.WithEF(err, b.fields.WithField("path", upperPath), "Failed to chdir to upper base path")
	}
	if err := common.Tar(b.aciTargetPath+common.PathImageAci, common.PathManifest[1:], common.PathRootfs[1:]+"/"); err != nil {
		return errs.WithEF(err, b.fields, "Failed to tar aci")
	}
	logs.WithField("path", dir).Debug("chdir")
	return nil
}

func (b *Builder) runBuildSetup() error { //TODO REMOVE
	if empty, err := common.IsDirEmpty(b.aciHomePath + PATH_RUNLEVELS + PATH_BUILD_SETUP); empty || err != nil {
		return nil
	}

	logs.WithF(b.fields).Info("Running build setup")

	for _, e := range manifestApp(b.pod).App.Environment {
		logs.WithField("name", e.Name).WithField("value", e.Value).Debug("Adding environment var")
		os.Setenv(e.Name, e.Value)
	}

	logs.WithF(b.fields).Warn("Build setup is deprecated and will be removed. it create unreproductible builds and run as root directly on the host. Please use builder dependencies and builder runlevels instead")
	time.Sleep(5 * time.Second)

	os.Setenv("BASEDIR", b.aciHomePath)
	os.Setenv("TARGET", b.stage2Rootfs+"/..")
	os.Setenv("ROOTFS", b.stage2Rootfs+"/../rootfs")
	os.Setenv(common.EnvLogLevel, logs.GetLevel().String())

	if err := common.ExecCmd(b.stage1Rootfs + PATH_DGR + PATH_BUILDER + "/stage2/build-setup.sh"); err != nil {
		return errs.WithEF(err, b.fields, "Build setup failed")
	}

	return nil
}

func (b *Builder) runBuild() error {
	command, err := b.getCommandPath()
	if err != nil {
		return err
	}

	logs.WithF(b.fields).Debug("Running build command")
	args, env := b.prepareNspawnArgsAndEnv(command)

	if logs.IsDebugEnabled() {
		logs.WithField("command", strings.Join([]string{args[0], " ", strings.Join(args[1:], " ")}, " ")).Debug("Running external command")
	}
	//	var stderr bytes.Buffer
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return errs.WithEF(err, b.fields, "Builder run failed")
	}

	return nil
}

func (b *Builder) prepareNspawnArgsAndEnv(commandPath string) ([]string, []string) {
	var args []string
	env := os.Environ()

	args = append(args, b.stage1Rootfs+"/dgr/usr/lib/ld-linux-x86-64.so.2")
	args = append(args, b.stage1Rootfs+"/dgr/usr/bin/systemd-nspawn")
	if context := os.Getenv(rktcommon.EnvSELinuxContext); context != "" {
		args = append(args, fmt.Sprintf("-Z%s", context))
	}
	args = append(args, "--register=no")
	args = append(args, "-q")
	args = append(args, "--link-journal=auto")
	env = append(env, "LD_LIBRARY_PATH="+b.stage1Rootfs+"/dgr/usr/lib")
	if !logs.IsDebugEnabled() {
		args = append(args, "--quiet")
	}
	lvl := "info"
	switch logs.GetLevel() {
	case logs.FATAL:
		lvl = "crit"
	case logs.PANIC:
		lvl = "alert"
	case logs.ERROR:
		lvl = "err"
	case logs.WARN:
		lvl = "warning"
	case logs.INFO:
		lvl = "info"
	case logs.DEBUG | logs.TRACE:
		lvl = "debug"
	}
	args = append(args, "--uuid="+b.pod.UUID.String())
	args = append(args, "--machine=dgr"+b.pod.UUID.String())
	env = append(env, "SYSTEMD_LOG_LEVEL="+lvl)

	for _, e := range manifestApp(b.pod).App.Environment {
		if e.Name != common.EnvBuilderCommand && e.Name != common.EnvAciTarget {
			args = append(args, "--setenv="+e.Name+"="+e.Value)
		}
	}

	trap, _ := manifestApp(b.pod).App.Environment.Get(common.EnvTrapOnError)
	args = append(args, "--setenv="+common.EnvTrapOnError+"="+string(trap))

	version, ok := manifestApp(b.pod).Image.Labels.Get("version")
	if ok {
		args = append(args, "--setenv=ACI_VERSION="+version)
	}
	args = append(args, "--setenv=ACI_NAME="+manifestApp(b.pod).Name.String())
	args = append(args, "--setenv=ACI_EXEC="+"'"+strings.Join(manifestApp(b.pod).App.Exec, "' '")+"'")
	args = append(args, "--setenv=ROOTFS="+PATH_OPT+PATH_STAGE2+"/"+manifestApp(b.pod).Name.String()+common.PathRootfs)

	args = append(args, "--capability=all")
	args = append(args, "--directory="+b.stage1Rootfs)
	args = append(args, "--bind="+b.aciHomePath+"/:/dgr/aci-home")
	args = append(args, "--bind="+b.aciTargetPath+"/:/dgr/aci-target")
	args = append(args, commandPath)

	return args, env
}

func (b *Builder) getCommandPath() (string, error) {
	command, ok := manifestApp(b.pod).App.Environment.Get(common.EnvBuilderCommand)
	if !ok {
		return string(common.CommandBuild), errs.WithF(b.fields.WithField("env_name", common.EnvBuilderCommand), "No command sent to builder using environment var")
	}

	key, err := common.BuilderCommand(command).CommandManifestKey()
	if err != nil {
		return "", errs.WithEF(err, b.fields, "Unknown command")
	}

	stage1Manifest, err := b.getStage1Manifest()
	if err != nil {
		return "", err
	}

	commandPath, ok := stage1Manifest.Annotations.Get(key)
	if !ok {
		return "", errs.WithEF(err, b.fields.WithField("key", key), "Stage1 image manifest does not have command annotation")
	}

	return string(commandPath), nil
}

func (b *Builder) getStage1Manifest() (*schema.ImageManifest, error) {
	content, err := ioutil.ReadFile(rktcommon.Stage1ManifestPath(b.pod.Root))
	if err != nil {
		return nil, errs.WithEF(err, b.fields.WithField("file", rktcommon.Stage1ManifestPath(b.pod.Root)), "Failed to read stage1 manifest")
	}

	im := &schema.ImageManifest{}
	err = im.UnmarshalJSON(content)
	if err != nil {
		return nil, errs.WithEF(err, b.fields.WithField("content", string(content)).
			WithField("file", rktcommon.Stage1ManifestPath(b.pod.Root)), "Cannot unmarshall json content from file")
	}
	return im, nil
}

func (b *Builder) upperTreeStoreId() (string, error) {
	if b.upperId == "" {
		treeStoreIDFilePath := rktcommon.AppTreeStoreIDPath(b.pod.Root, manifestApp(b.pod).Name)
		treeStoreID, err := ioutil.ReadFile(treeStoreIDFilePath)
		if err != nil {
			return "", errs.WithEF(err, b.fields.WithField("path", treeStoreIDFilePath), "Failed to read treeStoreID from file")
		}
		b.upperId = string(treeStoreID)
	}
	return b.upperId, nil
}

/////////////////////////////////////////

func manifestApp(pod *stage1commontypes.Pod) schema.RuntimeApp {
	return pod.Manifest.Apps[0]
}
