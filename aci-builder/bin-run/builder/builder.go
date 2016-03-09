package builder

import (
	"bytes"
	"encoding/json"
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

	aciPath, ok := manifestApp(pod).App.Environment.Get(common.ENV_ACI_PATH)
	if !ok || aciPath == "" {
		return nil, errs.WithF(fields, "Builder image require "+common.ENV_ACI_PATH+" environment variable")
	}
	aciTarget, ok := manifestApp(pod).App.Environment.Get(common.ENV_ACI_TARGET)
	if !ok || aciPath == "" {
		return nil, errs.WithF(fields, "Builder image require "+common.ENV_ACI_TARGET+" environment variable")
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
	defer b.chownTargetFiles()

	lfd, err := rktcommon.GetRktLockFD()
	if err != nil {
		return errs.WithEF(err, b.fields, "can't get rkt lock fd")
	}

	if err := sys.CloseOnExec(lfd, true); err != nil {
		return errs.WithEF(err, b.fields, "can't set FD_CLOEXEC on rkt lock")
	}

	if err := b.runBuildSetup(); err != nil { // TODO run as non-root ???
		return err // TODO DO NOT EVEN RUN HERE
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

	manifestPath := b.stage1Rootfs + PATH_OPT + PATH_STAGE2 + "/" + manifestApp(b.pod).Name.String() + common.PATH_MANIFEST
	content, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return errs.WithEF(err, b.fields.WithField("file", manifestPath), "Failed to read manifest file")
	}
	im := &schema.ImageManifest{}
	err = im.UnmarshalJSON(content)
	if err != nil {
		return errs.WithEF(err, data.WithField("content", string(content)), "Cannot unmarshall json content")
	}

	im.Name.Set(strings.Replace(im.Name.String(), "builder/", "", 1)) //TODO dirty and will not work for inherited builds
	if content, err := json.MarshalIndent(im, "", "  "); err != nil {
		return errs.WithEF(err, b.fields, "Failed to write manifest")
	} else if err := ioutil.WriteFile(b.pod.Root+PATH_OVERLAY+"/"+upperId+PATH_UPPER+common.PATH_MANIFEST, content, 0644); err != nil {
		return errs.WithEF(err, b.fields, "Failed to write manifest")
	}
	return nil
}

func (b *Builder) chownTargetFiles() {
	if os.Getenv(SUDO_UID) != "" {
		logs.WithF(b.fields).Debug("Give back ownership of target directory")
		if err := common.ExecCmd("chown", os.Getenv(SUDO_UID)+":"+os.Getenv(SUDO_GID), "-R", b.aciTargetPath); err != nil { // TODO path target may not be there
			logs.WithEF(err, b.fields).WithField("uid", os.Getenv(SUDO_UID)).WithField("gid", os.Getenv(SUDO_GID)).
				Warn("Cannot give back ownership of target directory")
		}
	}
}

func (b *Builder) tarAci() error {
	upperId, err := b.upperTreeStoreId()
	if err != nil {
		return err
	}

	upperPath := b.pod.Root + PATH_OVERLAY + "/" + upperId + PATH_UPPER
	upperNamedRootfs := upperPath + "/" + manifestApp(b.pod).Name.String()
	upperRootfs := upperPath + common.PATH_ROOTFS

	if err := os.Rename(upperNamedRootfs, upperRootfs); err != nil { // TODO this is dirty and can probably be renamed during tar
		return errs.WithEF(err, b.fields.WithField("path", upperNamedRootfs), "Failed to rename rootfs")
	}
	defer os.Rename(upperRootfs, upperNamedRootfs)

	//
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
	if err := common.Tar(false, b.aciTargetPath+common.PATH_IMAGE_ACI, common.PATH_MANIFEST[1:], common.PATH_ROOTFS[1:]+"/"); err != nil {
		return errs.WithEF(err, b.fields, "Failed to tar aci")
	}
	logs.WithField("path", dir).Debug("chdir")
	return nil
}

func (b *Builder) runBuildSetup() error { // TODO do not run as root ??
	if empty, err := common.IsDirEmpty(b.aciHomePath + PATH_RUNLEVELS + PATH_BUILD_SETUP); empty || err != nil {
		return nil
	}

	logs.WithF(b.fields).Info("Running build setup")

	os.Setenv("BASEDIR", b.aciHomePath)
	os.Setenv("TARGET", b.stage2Rootfs+"/..") //TODO
	os.Setenv(common.ENV_LOG_LEVEL, logs.GetLevel().String())

	if err := common.ExecCmd(b.stage1Rootfs + PATH_DGR + PATH_BUILDER + "/stage2/build-setup.sh"); err != nil { // TODO this sux
		return errs.WithEF(err, b.fields, "Build setup failed")
	}

	return nil
}

func (b *Builder) runBuild() error {
	if empty, err := common.IsDirEmpty(b.aciHomePath + PATH_RUNLEVELS + PATH_BUILD); empty || err != nil {
		return nil
	}

	logs.WithF(b.fields).Debug("Running build")

	args, env := b.prepareNspawnArgsAndEnv()

	if logs.IsDebugEnabled() {
		logs.WithField("command", strings.Join([]string{args[0], " ", strings.Join(args[1:], " ")}, " ")).Debug("Running external command")
	}
	var stderr bytes.Buffer
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return errs.WithEF(err, b.fields.WithField("stderr", stderr), "Builder run failed")
	}

	return nil
}

func (b *Builder) prepareNspawnArgsAndEnv() ([]string, []string) {
	var args []string
	env := os.Environ()

	args = append(args, b.stage1Rootfs+"/usr/lib/ld-linux-x86-64.so.2")
	args = append(args, b.stage1Rootfs+"/usr/bin/systemd-nspawn")
	//	if context := os.Getenv(common.EnvSELinuxContext); context != "" {
	//		args = append(args, fmt.Sprintf("-Z%s", context))
	//	}
	args = append(args, "--register=no")
	args = append(args, "--link-journal=auto")
	env = append(env, "LD_LIBRARY_PATH="+b.stage1Rootfs+"/usr/lib")
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

	args = append(args, "--setenv=LOG_LEVEL="+logs.GetLevel().String())
	args = append(args, "--setenv=ACI_NAME="+manifestApp(b.pod).Name.String())
	args = append(args, "--setenv=ACI_EXEC="+"'"+strings.Join(manifestApp(b.pod).App.Exec, "' '")+"'") // TODO hum
	args = append(args, "--capability=all")
	args = append(args, "--directory="+b.stage1Rootfs)
	args = append(args, "--bind="+b.aciTargetPath+"/:/opt/aci-target")
	args = append(args, "--bind="+b.aciHomePath+"/:/opt/aci-home")
	args = append(args, "/build")

	return args, env
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
