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

	logs.WithF(aci.fields).Info("Gzipping aci before upload")

	im, err := common.ExtractManifestFromAci(aci.target + pathImageAci)
	if err != nil {
		return errs.WithEF(err, aci.fields.WithField("file", pathImageAci), "Failed to extract manifest from aci file")
	}
	val, ok := im.Labels.Get("version")
	if !ok {
		return errs.WithEF(err, aci.fields.WithField("file", pathImageAci), "Failed to get version from aci manifest")
	}

	if err := aci.zipAci(); err != nil {
		return errs.WithEF(err, aci.fields, "Failed to zip aci")
	}

	if Home.Config.Push.Type == "maven" {
		logs.WithF(aci.fields).Info("Uploading aci")
		if err := common.ExecCmd("curl", "-f", "-i",
			"-F", "r=releases",
			"-F", "hasPom=false",
			"-F", "e=aci",
			"-F", "g=com.blablacar.aci.linux.amd64",
			"-F", "p=aci",
			"-F", "v="+val,
			"-F", "a="+strings.Split(string(im.Name), "/")[1],
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

		err = Uploader{
			Acipath:  aci.target + pathImageGzAci,
			Ascpath:  aci.target + pathImageGzAci,
			Uri:      aci.manifest.NameAndVersion.String(),
			Insecure: true,
			Debug:    false,
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
		}.Upload()
		if err != nil {
			return errs.WithEF(err, aci.fields, "Failed to upload aci")
		}

	}
	return nil
}
