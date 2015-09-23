package builder

import (
	"github.com/blablacar/cnt/dist"
	"github.com/blablacar/cnt/log"
	"github.com/blablacar/cnt/spec"
	"github.com/blablacar/cnt/utils"
	"io/ioutil"
	"os"
)

func (cnt *Img) Test() {
	cnt.checkInstalled()
	log.Get().Info("Testing " + cnt.manifest.NameAndVersion)

	cnt.importAciBats()

	testAci, err := cnt.prepareTestAci()
	if err != nil {
		log.Get().Panic(err)
	}
	testAci.Install()

	// cleanup
	// report ?

	//TODO add flag to clean rkt after test

	if err := utils.ExecCmd("rkt", "--insecure-skip-verify=true", "run", cnt.target+PATH_TEST+PATH_TARGET+"/image.aci"); err != nil {
		log.Get().Panic(err)
	}

	// remove test image from rkt

	//	if err := utils.ExecCmd("systemd-nspawn", "--directory=" + cnt.rootfs, "--capability=all",
	//		"--bind=" + cnt.target + "/:/target", "--share-system", "target/build.sh"); err != nil {
	//		log.Get().Panic("Build step did not succeed", err)
	//
	//
	//		utils.ExecCmd("rkt", "--insecure-skip-verify=true", "run", cnt.target + "/image.aci")
	//	}
}

func (cnt *Img) importAciBats() {
	if err := utils.ExecCmd("bash", "-c", "rkt image list --fields name --no-legend | grep aci-bats"); err != nil {
		content, _ := dist.Asset("dist/bindata/aci-bats.aci")
		if err := ioutil.WriteFile("/tmp/aci-bats.aci", content, 0644); err != nil {
			log.Get().Panic(err)
		}
		utils.ExecCmd("rkt", "--insecure-skip-verify=true", "fetch", "/tmp/aci-bats.aci")
		os.Remove("/tmp/aci-bats.aci")
	}
}

func (cnt *Img) prepareTestAci() (*Img, error) {

	files, err := ioutil.ReadDir(cnt.path + PATH_TEST)
	if err != nil {
		return nil, err
	}

	os.MkdirAll(cnt.target+PATH_TEST+PATH_FILES+PATH_TEST, 0777)
	for _, f := range files {
		if !f.IsDir() {
			if err := utils.CopyFile(cnt.path+PATH_TEST+"/"+f.Name(), cnt.target+PATH_TEST+PATH_FILES+PATH_TEST+"/"+f.Name()); err != nil {
				log.Get().Panic(err)
			}
		}
	}

	fullname, err := spec.NewACFullName(cnt.manifest.NameAndVersion.Name() + "_test:" + cnt.manifest.NameAndVersion.Version())
	if err != nil {
		log.Get().Panic(err)
	}
	testAci, err := NewAciWithManifest(cnt.target+PATH_TEST, cnt.args, spec.AciManifest{
		Aci: spec.AciDefinition{
			App: &spec.CntApp{
				Exec: []string{"/test.sh"},
			},
			Dependencies: []spec.ACFullname{cnt.manifest.NameAndVersion, "aci.blablacar.com/aci-bats:1"},
		},
		NameAndVersion: *fullname,
	})
	if err != nil {
		log.Get().Panic("Cannot build test aci", err)
	}
	return testAci, nil
}
