package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/appc/spec/schema"
	"github.com/appc/spec/schema/types"
	"github.com/blablacar/dgr/bin-templater/merger"
	"github.com/blablacar/dgr/dgr/common"
	rktcommon "github.com/coreos/rkt/common"
	"github.com/coreos/rkt/pkg/sys"
	stage1commontypes "github.com/coreos/rkt/stage1/common/types"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
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

	attrMerger, err := merger.NewAttributesMerger(b.stage1Rootfs+PATH_DGR+PATH_BUILDER, PATH_ATTRIBUTES)
	if err != nil {
		logs.WithE(err).Warn("Failed to prepare attributes")
	}
	attributes := attrMerger.Merge()
	logs.WithFields(b.fields).WithField("attributes", attributes).Debug("Merged attributes for manifest templating")

	content, err := ioutil.ReadFile(b.aciTargetPath + common.PathManifestYmlTmpl)
	if err != nil {
		return errs.WithEF(err, b.fields.WithField("file", b.aciTargetPath+common.PathManifestYmlTmpl), "Failed to read manifest template")
	}

	aciManifest, err := common.ProcessManifestTemplate(string(content), attributes, true)
	if err != nil {
		return errs.WithEF(err, b.fields.WithField("content", string(content)), "Failed to process manifest template")
	}
	target := b.pod.Root + PATH_OVERLAY + "/" + upperId + PATH_UPPER + common.PathManifest

	dgrVersion, ok := manifestApp(b.pod).App.Environment.Get(common.EnvDgrVersion)
	if !ok {
		return errs.WithF(b.fields, "Cannot find dgr version")
	}

	if aciManifest.NameAndVersion.Version() == "" {
		aciManifest.NameAndVersion = *common.NewACFullName(aciManifest.NameAndVersion.Name() + ":" + common.GenerateVersion(b.aciTargetPath))
	}

	if err := common.WriteAciManifest(aciManifest, target, aciManifest.NameAndVersion.Name(), dgrVersion); err != nil {
		return errs.WithEF(err, b.fields.WithField("file", target), "Failed to write manifest")
	}
	return nil
}

func (b *Builder) tarAci() error {
	upperId, err := b.upperTreeStoreId()
	if err != nil {
		return err
	}

	upperPath := b.pod.Root + PATH_OVERLAY + "/" + upperId + PATH_UPPER
	rootfsAlias := manifestApp(b.pod).Name.String()
	destination := b.aciTargetPath + common.PathImageAci // absolute dir, outside upperPath (think: /tmp/…)

	params := []string{"--sort=name", "--numeric-owner", "--exclude", rootfsAlias + PATH_TMP + "/*"}
	params = append(params, "-C", upperPath, "--transform", "s@^"+rootfsAlias+"@rootfs@")
	params = append(params, "-cf", destination, common.PathManifest[1:], rootfsAlias)

	logs.WithF(b.fields).Debug("Calling tar to collect all files")
	if err := common.ExecCmd("tar", params...); err != nil {
		return errs.WithEF(err, b.fields, "Failed to tar aci")
	}
	// common.ExecCmd sometimes silently fails, hence the redundant check.
	if _, err := os.Stat(destination); os.IsNotExist(err) {
		return errs.WithEF(err, b.fields, "Expected aci has not been created")
	}
	return nil
}

func (b *Builder) runBuild() error {
	command, err := b.getCommandPath()
	if err != nil {
		return err
	}

	logs.WithF(b.fields).Debug("Running build command")
	args, env, err := b.prepareNspawnArgsAndEnv(command)
	if err != nil {
		return err
	}

	os.Remove(b.stage1Rootfs + "/etc/machine-id")

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

func (b *Builder) prepareNspawnArgsAndEnv(commandPath string) ([]string, []string, error) {
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

	catchError, _ := manifestApp(b.pod).App.Environment.Get(common.EnvCatchOnError)
	catchStep, _ := manifestApp(b.pod).App.Environment.Get(common.EnvCatchOnStep)
	args = append(args, "--setenv="+common.EnvCatchOnError+"="+string(catchError))
	args = append(args, "--setenv="+common.EnvCatchOnStep+"="+string(catchStep))

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

	//
	content, err := ioutil.ReadFile(b.aciTargetPath + common.PathManifestYmlTmpl)
	if err != nil {
		return args, env, errs.WithEF(err, b.fields.WithField("file", b.aciTargetPath+common.PathManifestYmlTmpl), "Failed to read manifest template")
	}

	aciManifest, err := common.ProcessManifestTemplate(string(content), nil, false)
	if err != nil {
		return args, env, errs.WithEF(err, b.fields.WithField("content", string(content)), "Failed to process manifest template")
	}
	for _, mount := range aciManifest.Builder.MountPoints {
		if strings.HasPrefix(mount.From, "~/") {
			user, err := user.Current()
			if err != nil {
				return args, env, errs.WithEF(err, b.fields, "Cannot found current user")
			}
			mount.From = user.HomeDir + mount.From[1:]
		}
		from := mount.From
		if from[0] != '/' {
			from = b.aciHomePath + "/" + from
		}

		if _, err := os.Stat(from); err != nil {
			os.MkdirAll(from, 0755)
		}
		args = append(args, "--bind="+from+":"+mount.To)
	}

	for _, mount := range aciManifest.Build.MountPoints {
		if strings.HasPrefix(mount.From, "~/") {
			user, err := user.Current()
			if err != nil {
				return args, env, errs.WithEF(err, b.fields, "Cannot found current user")
			}
			mount.From = user.HomeDir + mount.From[1:]
		}
		from := mount.From
		if from[0] != '/' {
			from = b.aciHomePath + "/" + from
		}

		if _, err := os.Stat(from); err != nil {
			os.MkdirAll(from, 0755)
		}
		args = append(args, "--bind="+from+":"+PATH_OPT+PATH_STAGE2+"/"+manifestApp(b.pod).Name.String()+common.PathRootfs+"/"+mount.To)
	}

	args = append(args, commandPath)

	return args, env, nil
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
