package main

import (
	"net/http"
	"strings"

	"net/url"
	"os"
	"path"

	"bytes"
	"io"
	"mime/multipart"
	"path/filepath"

	"github.com/blablacar/dgr/dgr/common"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"github.com/rkt/rkt/rkt/config"
)

func (aci *Aci) Push() error {
	defer aci.giveBackUserRightsToTarget()
	logs.WithF(aci.fields).Info("Pushing")

	if aci.isUpdated() {
		logs.WithFields(aci.fields).Warn("You cannot push an updated aci, rebuilding")
		if err := aci.CleanAndBuild(); err != nil {
			return err
		}
	}

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

func (aci *Aci) nexusRegistryCheck(urlReq url.URL, params map[string]string) (ok bool, err error) {
	logs.WithField("url", urlReq.String()).Debug("Checking remote registry")
	client := &http.Client{}
	queryCheck := urlReq.Query()
	for key, val := range params {
		queryCheck.Set(key, val)
	}
	urlReq.RawQuery = queryCheck.Encode()

	/* Authenticate */
	req, err := http.NewRequest("HEAD", urlReq.String(), nil)
	req.SetBasicAuth(Home.Config.Push.Username, Home.Config.Push.Password)
	resp, err := client.Do(req)
	if err != nil {
		logs.WithE(err).Info("Failed to check version")
		return false, err
	}
	ok = false
	switch {
	case resp.StatusCode >= 200 && resp.StatusCode <= 300:
		logs.WithField("statusCode", resp.StatusCode).
			WithField("status", resp.Status).
			Info("Aci already uploaded.")
	case 404 == resp.StatusCode:
		logs.WithField("statusCode", resp.StatusCode).
			WithField("status", resp.Status).
			Info("Ready to upload")
		ok = true
	case 401 == resp.StatusCode || 403 == resp.StatusCode:
		logs.WithField("statusCode", resp.StatusCode).
			WithField("status", resp.Status).
			Info("Failed to login to the registry")
	case resp.StatusCode >= 500:
		logs.WithField("statusCode", resp.StatusCode).
			WithField("status", resp.Status).
			Info("Registry server error")
	}
	return
}

func (aci *Aci) nexusRegistryPush(urlReq url.URL, aciPath string, params map[string]string) (err error) {
	file, err := os.Open(aciPath)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	part, err := writer.CreateFormFile("file", filepath.Base(aciPath))
	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)

	err = writer.Close()
	if err != nil {
		return err
	}

	client := &http.Client{}

	/* Authenticate */
	req, err := http.NewRequest("POST", urlReq.String(), body)
	req.SetBasicAuth(Home.Config.Push.Username, Home.Config.Push.Password)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := client.Do(req)
	if err != nil {
		return errs.WithE(err, "Failed to execute request")
	}
	defer resp.Body.Close()
	logs.WithField("statusCode", resp.StatusCode).
		WithField("status", resp.Status).
		WithField("aciPath", aciPath).
		Debug("Upload logs")
	if resp.StatusCode != 201 {
		return errs.WithE(err, resp.Status)
	}
	return
}

func (aci *Aci) upload(name *common.ACFullname) error {
	if Home.Config.Push.Type == "maven" && name.DomainName() == "aci.blbl.cr" { // TODO this definitely need to be removed
		urlCheck, err := url.Parse(Home.Config.Push.Url)
		if err != nil {
			return errs.WithEF(err, data.Fields{
				"url": Home.Config.Push.Url,
			}, "Failed to parse url in dgr config")
		}
		formArgs := map[string]string{
			"r":         "releases",
			"hasPom":    "false",
			"e":         "aci",
			"p":         "aci",
			"extension": "aci",
			"packaging": "aci",
			"g":         "com.blablacar.aci.linux.amd64",
			"v":         name.Version(),
			"a":         strings.Split(string(name.Name()), "/")[1],
		}
		urlCheck.Path = urlCheck.Path + "/service/local/artifact/maven/content"
		isNotInRegistry, err := aci.nexusRegistryCheck(*urlCheck, formArgs)
		if err != nil {
			return errs.WithEF(err, aci.fields, "Failed to check aci in remote registry")
		}
		aciPath := path.Join(aci.target, pathImageGzAci)
		if isNotInRegistry {
			logs.WithF(aci.fields).Info("Uploading aci")
			if err = aci.nexusRegistryPush(*urlCheck, aciPath, formArgs); err != nil {
				return errs.WithEF(err, aci.fields, "Failed to push aci")
			}
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
			Uri:     name.String(),
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
				header := headerer.GetHeader()
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
