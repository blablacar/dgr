package main

import (
	"github.com/blablacar/dgr/bin-dgr/common"
	"github.com/coreos/rkt/rkt/config"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"net/http"
	"strings"
)

func (aci *Aci) Push() error {
	defer aci.giveBackUserRightsToTarget()

	if err := aci.EnsureBuilt(); err != nil {
		return err
	}

	if aci.args.Test {
		aci.args.Test = false
		if err := aci.Test(); err != nil {
			return err
		}
	}

	if err := aci.EnsureZipSign(); err != nil {
		return err
	}

	im, err := common.ExtractManifestFromAci(aci.target + pathImageAci)
	if err != nil {
		return errs.WithEF(err, aci.fields.WithField("file", pathImageAci), "Failed to extract manifest from aci file")
	}

	return aci.upload(common.ExtractNameVersionFromManifest(im))
}

func (aci *Aci) upload(name *common.ACFullname) error {
	if Home.Config.Push.Type == "maven" && name.DomainName() == "aci.blbl.cr" { // TODO this definitely need to be removed
		logs.WithF(aci.fields).Info("Uploading aci")
		if err := common.ExecCmd("curl", "-f", "-i",
			"-F", "r=releases",
			"-F", "hasPom=false",
			"-F", "e=aci",
			"-F", "g=com.blablacar.aci.linux.amd64",
			"-F", "p=aci",
			"-F", "v="+name.Version(),
			"-F", "a="+strings.Split(string(name.Name()), "/")[1],
			"-F", "file=@"+aci.target+pathImageGzAci,
			"-u", Home.Config.Push.Username+":"+Home.Config.Push.Password,
			Home.Config.Push.Url+"/service/local/artifact/maven/content"); err != nil {
			return errs.WithEF(err, aci.fields, "Failed to push aci")
		}
	} else {
		systemConf := Home.Config.Rkt.SystemConfig
		if systemConf == "" {
			systemConf = "/usr/lib/rkt"
		}
		localConf := Home.Config.Rkt.LocalConfig
		if localConf == "" {
			localConf = "/etc/rkt"
		}

		conf, err := config.GetConfigFrom(systemConf, localConf)
		if err != nil {
			return errs.WithEF(err, aci.fields, "Failed to get rkt configuration")
		}

		upload := Uploader{
			Acipath: aci.target + pathImageGzAci,
			Ascpath: aci.target + pathImageGzAciAsc,
			Uri:     aci.manifest.NameAndVersion.String(),
			Debug:   false,
			SetHTTPHeaders: func(r *http.Request) {
				if r.URL == nil {
					return
				}
				headerer, ok := conf.AuthPerHost[r.URL.Host]
				if !ok {
					logs.WithFields(aci.fields).WithField("domain", r.URL.Host).
						Warn("No auth credential found in rkt configuration for this domain")
					return
				}
				header := headerer.Header()
				for k, v := range header {
					r.Header[k] = append(r.Header[k], v...)
				}
			},
		}
		err = upload.Upload()
		if err != nil {
			return errs.WithEF(err, aci.fields, "Failed to upload aci")
		}

	}
	return nil
}
