package main

import (
	"encoding/json"
	"github.com/appc/spec/schema"
	"github.com/appc/spec/schema/types"
	"github.com/blablacar/dgr/bin-dgr/common"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

func (aci *Aci) prepareRktRunArguments(command common.BuilderCommand, builderHash string, stage1Hash string) []string {
	var args []string

	if logs.IsDebugEnabled() {
		args = append(args, "--debug")
	}
	args = append(args, "--set-env="+common.EnvLogLevel+"="+logs.GetLevel().String())
	args = append(args, "--set-env="+common.EnvAciPath+"="+aci.path)
	args = append(args, "--set-env="+common.EnvAciTarget+"="+aci.target)
	args = append(args, "--set-env="+common.EnvBuilderCommand+"="+string(command))
	args = append(args, "--set-env="+common.EnvTrapOnError+"="+strconv.FormatBool(aci.args.TrapOnError))
	args = append(args, "--net=host")
	args = append(args, "--insecure-options=image")
	args = append(args, "--uuid-file-save="+aci.target+pathBuilderUuid)
	args = append(args, "--interactive")
	if stage1Hash != "" {
		args = append(args, "--stage1-hash="+stage1Hash)
	} else {
		args = append(args, "--stage1-name="+aci.manifest.Builder.Image.String())
	}

	for _, v := range aci.args.SetEnv.Strings() {
		args = append(args, "--set-env="+v)
	}
	args = append(args, builderHash)
	return args
}

func (aci *Aci) RunBuilderCommand(command common.BuilderCommand) error {
	defer aci.giveBackUserRightsToTarget()
	aci.Clean()

	logs.WithF(aci.fields).Info("Building")
	if err := os.MkdirAll(aci.target, 0777); err != nil {
		return errs.WithEF(err, aci.fields, "Cannot create target directory")
	}

	stage1Hash, err := aci.prepareStage1aci()
	if err != nil {
		return errs.WithEF(err, aci.fields, "Failed to prepare stage1 image")
	}

	builderHash, err := aci.prepareBuildAci()
	if err != nil {
		return errs.WithEF(err, aci.fields, "Failed to prepare build image")
	}

	defer aci.cleanupRun(builderHash, stage1Hash)
	if err := Home.Rkt.Run(aci.prepareRktRunArguments(command, builderHash, stage1Hash)); err != nil {
		return errs.WithEF(err, aci.fields, "Builder container return with failed status")
	}

	if content, err := common.ExtractManifestContentFromAci(aci.target + pathImageAci); err != nil {
		logs.WithEF(err, aci.fields).Warn("Failed to write manifest.json")
	} else if err := ioutil.WriteFile(aci.target+pathManifestJson, content, 0644); err != nil {
		logs.WithEF(err, aci.fields).Warn("Failed to write manifest.json")
	}

	return nil
}

func (aci *Aci) cleanupRun(builderHash string, stage1Hash string) {
	if !Args.KeepBuilder {
		if _, _, err := Home.Rkt.RmFromFile(aci.target + pathBuilderUuid); err != nil {
			logs.WithEF(err, aci.fields).Warn("Failed to remove build container")
		}
	}

	if err := Home.Rkt.ImageRm(builderHash); err != nil {
		logs.WithEF(err, aci.fields.WithField("hash", builderHash)).Warn("Failed to remove build container image")
	}

	if stage1Hash != "" {
		if err := Home.Rkt.ImageRm(stage1Hash); err != nil {
			logs.WithEF(err, aci.fields.WithField("hash", stage1Hash)).Warn("Failed to remove stage1 container image")
		}
	}
}

func (aci *Aci) CleanAndBuild() error {
	return aci.RunBuilderCommand(common.CommandBuild)
}

func (aci *Aci) prepareStage1aci() (string, error) {
	ImportInternalBuilderIfNeeded(aci.manifest)
	if len(aci.manifest.Builder.Dependencies) == 0 {
		return "", nil
	}

	logs.WithFields(aci.fields).Debug("Preparing stage1")

	if err := os.MkdirAll(aci.target+pathStage1+common.PathRootfs, 0777); err != nil {
		return "", errs.WithEF(err, aci.fields.WithField("path", aci.target+pathBuilder), "Failed to create stage1 aci path")
	}

	manifestStr, err := Home.Rkt.CatManifest(aci.manifest.Builder.Image.String())
	if err != nil {
		return "", errs.WithEF(err, aci.fields, "Failed to read stage1 image manifest")
	}

	manifest := schema.ImageManifest{}
	if err := json.Unmarshal([]byte(manifestStr), &manifest); err != nil {
		return "", errs.WithEF(err, aci.fields.WithField("content", manifestStr), "Failed to unmarshal stage1 manifest received from rkt")
	}

	manifest.Dependencies = types.Dependencies{}
	stage1Image, err := toAppcDependencies([]common.ACFullname{aci.manifest.Builder.Image})
	if err != nil {
		return "", errs.WithEF(err, aci.fields, "Invalid image on stage1 for rkt")
	}
	manifest.Dependencies = append(manifest.Dependencies, stage1Image...)

	dep, err := toAppcDependencies(aci.manifest.Builder.Dependencies)
	if err != nil {
		return "", errs.WithEF(err, aci.fields, "Invalid dependency on stage1 for rkt")
	}
	manifest.Dependencies = append(manifest.Dependencies, dep...)

	name, err := types.NewACIdentifier(prefixBuilderStage1 + aci.manifest.NameAndVersion.Name())
	if err != nil {
		return "", errs.WithEF(err, aci.fields.WithField("name", prefixBuilderStage1+aci.manifest.NameAndVersion.Name()),
			"aci name is not a valid identifier for rkt")
	}
	manifest.Name = *name

	content, err := json.Marshal(&manifest)
	if err != nil {
		return "", errs.WithEF(err, aci.fields, "Failed to marshal builder's stage1 manifest")
	}

	if err := ioutil.WriteFile(aci.target+pathStage1+common.PathManifest, content, 0644); err != nil {
		return "", errs.WithEF(err, aci.fields.WithField("path", aci.target+pathStage1+common.PathManifest),
			"Failed to write builder's stage1 manifest to file")
	}

	if err := aci.tarAci(aci.target + pathStage1); err != nil {
		return "", err
	}

	logs.WithF(aci.fields.WithField("path", aci.target+pathStage1+pathImageAci)).Info("Importing builder's stage1")
	hash, err := Home.Rkt.Fetch(aci.target + pathStage1 + pathImageAci)
	if err != nil {
		return "", errs.WithEF(err, aci.fields, "fetch of builder's stage1 aci failed")
	}
	return hash, nil
}

func (aci *Aci) prepareBuildAci() (string, error) {
	logs.WithFields(aci.fields).Debug("Preparing builder")

	if err := os.MkdirAll(aci.target+pathBuilder+common.PathRootfs, 0777); err != nil {
		return "", errs.WithEF(err, aci.fields.WithField("path", aci.target+pathBuilder), "Failed to create builder aci path")
	}

	if err := aci.WriteImageManifest(aci.manifest, aci.target+pathBuilder+common.PathManifest, common.PrefixBuilder+aci.manifest.NameAndVersion.Name()); err != nil {
		return "", err
	}
	if err := aci.tarAci(aci.target + pathBuilder); err != nil {
		return "", err
	}

	logs.WithF(aci.fields.WithField("path", aci.target+pathBuilder+pathImageAci)).Info("Importing builder")
	hash, err := Home.Rkt.Fetch(aci.target + pathBuilder + pathImageAci)
	if err != nil {
		return "", errs.WithEF(err, aci.fields, "fetch of builder aci failed")
	}
	return hash, nil
}

func (aci *Aci) EnsureBuilt() error {
	if _, err := os.Stat(aci.target + pathImageAci); os.IsNotExist(err) {
		if err := aci.CleanAndBuild(); err != nil {
			return err
		}
	}
	return nil
}

func (aci *Aci) EnsureZip() error {
	if _, err := os.Stat(aci.target + pathImageGzAci); os.IsNotExist(err) {
		if err := aci.EnsureBuilt(); err != nil {
			return err
		}

		if err := aci.zipAci(); err != nil {
			return err
		}
	}
	return nil
}

func (aci *Aci) WriteImageManifest(m *AciManifest, targetFile string, projectName string) error {
	name, err := types.NewACIdentifier(projectName)
	if err != nil {
		return errs.WithEF(err, aci.fields.WithField("name", projectName), "aci name is not a valid identifier for rkt")
	}

	labels := types.Labels{}
	if m.NameAndVersion.Version() != "" {
		labels = append(labels, types.Label{Name: "version", Value: m.NameAndVersion.Version()})
	}
	labels = append(labels, types.Label{Name: "os", Value: "linux"})
	labels = append(labels, types.Label{Name: "arch", Value: "amd64"})

	if m.Aci.App.User == "" {
		m.Aci.App.User = "0"
	}
	if m.Aci.App.Group == "" {
		m.Aci.App.Group = "0"
	}

	im := schema.BlankImageManifest()
	im.Annotations = m.Aci.Annotations

	dgrBuilderIdentifier, _ := types.NewACIdentifier(manifestDrgBuilder)
	dgrVersionIdentifier, _ := types.NewACIdentifier(manifestDrgVersion)
	buildDateIdentifier, _ := types.NewACIdentifier("build-date")
	im.Annotations.Set(*dgrVersionIdentifier, DgrVersion)
	im.Annotations.Set(*dgrBuilderIdentifier, m.Builder.Image.String())
	im.Annotations.Set(*buildDateIdentifier, time.Now().Format(time.RFC3339))
	im.Dependencies, err = toAppcDependencies(m.Aci.Dependencies)
	if err != nil {
		return errs.WithEF(err, aci.fields, "Failed to prepare dependencies for manifest")
	}
	im.Name = *name
	im.Labels = labels

	if len(m.Aci.App.Exec) == 0 {
		m.Aci.App.Exec = []string{"/dgr/bin/busybox", "sh"}
	}

	im.App = &types.App{
		Exec:             m.Aci.App.Exec,
		EventHandlers:    []types.EventHandler{{Name: "pre-start", Exec: []string{"/dgr/bin/prestart"}}},
		User:             m.Aci.App.User,
		Group:            m.Aci.App.Group,
		WorkingDirectory: m.Aci.App.WorkingDirectory,
		Environment:      m.Aci.App.Environment,
		MountPoints:      m.Aci.App.MountPoints,
		Ports:            m.Aci.App.Ports,
		Isolators:        m.Aci.App.Isolators,
	}

	buff, err := im.MarshalJSON()
	if err != nil {
		return errs.WithEF(err, aci.fields.WithField("object", im), "Failed to marshal manifest")
	}
	err = ioutil.WriteFile(targetFile, buff, 0644)
	if err != nil {
		return errs.WithEF(err, aci.fields.WithField("file", targetFile), "Failed to write manifest file")
	}
	return nil
}

func toAppcDependencies(dependencies []common.ACFullname) (types.Dependencies, error) {
	appcDependencies := types.Dependencies{}
	for _, dep := range dependencies {
		id, err := types.NewACIdentifier(dep.Name())
		if err != nil {
			return nil, errs.WithEF(err, data.WithField("name", dep.Name()), "invalid identifer name for rkt")
		}
		t := types.Dependency{ImageName: *id}
		if dep.Version() != "" {
			t.Labels = types.Labels{}
			t.Labels = append(t.Labels, types.Label{Name: "version", Value: dep.Version()})
		}

		appcDependencies = append(appcDependencies, t)
	}
	return appcDependencies, nil
}
