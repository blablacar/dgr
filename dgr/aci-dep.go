package main

import (
	"strconv"

	"github.com/appc/spec/schema"
	"github.com/blablacar/dgr/dgr/common"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
)

func GetDependencyDgrVersion(im schema.ImageManifest) (int, error) {
	version, ok := im.Annotations.Get(common.ManifestDrgVersion)
	if !ok {
		return 0, nil
	}

	val, err := strconv.Atoi(version)
	if err != nil {
		return 0, errs.WithEF(err, data.WithField("version", version), "Failed to parse "+common.ManifestDrgVersion+" from manifest")
	}
	return val, nil
}

func CheckLatestVersion(deps []common.ACFullname, warnText string) {
	for _, dep := range deps {
		if dep.Version() == "" {
			logs.WithField("dependency", dep.String()).Debug("Skip version check since no version set")
			continue
		}

		go func(dep common.ACFullname) {
			version, _ := dep.LatestVersion()
			logs.WithField("dep", dep.String()).Debug("version found : " + version)
			if version != "" && common.Version(dep.Version()).LessThan(common.Version(version)) {
				logs.WithField("newer", dep.Name()+":"+version).
					WithField("current", dep.String()).
					Warn("Newer " + warnText + " version")
			}
		}(dep)
	}
}

func (aci *Aci) prepareDependency(dep common.ACFullname) (*common.ACFullname, error) {
	depFields := aci.fields.WithField("dependency", dep.String())

	logs.WithF(aci.fields).WithFields(depFields).Info("Fetching dependency")

	pullPolicy := common.PullPolicyNew
	if dep.Version() == "" {
		pullPolicy = common.PullPolicyUpdate
	}

	hash, err := Home.Rkt.Fetch(dep.String(), pullPolicy)
	if err != nil {
		return nil, errs.WithEF(err, depFields, "Failed to fetch dependency")
	}

	manifest, err := Home.Rkt.GetManifest(hash)
	if err != nil {
		return nil, errs.WithEF(err, depFields, "Failed to cat manifest on aci just fetched")
	}

	checkCompatibility(manifest, depFields)

	version, ok := manifest.GetLabel("version")
	if !ok {
		logs.WithF(depFields).Warn("No version set on this dependency")
		return &dep, nil
	}
	return common.NewACFullnameWithVersion(dep, version), nil
}

func (aci *Aci) prepareAllDependencies() error {
	if err := aci.prepareDependencies(&aci.manifest.Aci.Dependencies, "dependency"); err != nil {
		return err
	}
	if err := aci.prepareDependencies(&aci.manifest.Builder.Dependencies, "builder dependency"); err != nil {
		return err
	}
	if err := aci.prepareDependencies(&aci.manifest.Tester.Builder.Dependencies, "tester builder dependency"); err != nil {
		return err
	}
	if err := aci.prepareDependencies(&aci.manifest.Tester.Aci.Dependencies, "tester dependency"); err != nil {
		return err
	}
	return nil
}

func (aci *Aci) prepareDependencies(deps *[]common.ACFullname, depTypeForLogs string) error {
	CheckLatestVersion(*deps, depTypeForLogs)

	for index, dep := range *deps {
		prepared, err := aci.prepareDependency(dep)
		if err != nil {
			return errs.WithEF(err, aci.fields, "Failed to prepare dependencies")
		}

		if string(*prepared) != string(dep) {
			logs.WithField("dependency", prepared.String()).Info("Setting dependency version")
			(*deps)[index] = *prepared
		}
	}

	return nil
}

func checkCompatibility(manifest schema.ImageManifest, fields data.Fields) {
	version, err := GetDependencyDgrVersion(manifest)
	if err != nil {
		logs.WithEF(err, fields).Error("Failed to check compatibility version of dependency")
	} else {
		if version < 55 {
			logs.WithF(fields).
				WithField("require", ">=55").
				Error("dependency was not build with a compatible version of dgr")
		}
	}
}
