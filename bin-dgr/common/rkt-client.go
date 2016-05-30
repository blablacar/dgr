package common

import (
	"bufio"
	"fmt"
	"github.com/appc/spec/discovery"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"strings"
)

const rktSupportedVersion Version = "1.4.0"

type InsecuOptions []string

func (i InsecuOptions) ToDiscoveryInsecureOption() discovery.InsecureOption {
	val := discovery.InsecureNone
	for _, option := range i {
		switch strings.ToLower(option) {
		case "tls":
			val |= discovery.InsecureTLS
		case "http":
			val |= discovery.InsecureHTTP
		}
	}
	return val
}

func (i InsecuOptions) HasImage() bool {
	for _, option := range i {
		if strings.ToLower(option) == "image" {
			return true
		}
	}
	return false
}

type RktConfig struct {
	Path               string        `yaml:"path"`
	InsecureOptions    InsecuOptions `yaml:"insecureOptions"`
	dir                string        `yaml:"dir"`
	LocalConfig        string        `yaml:"localConfig"`
	SystemConfig       string        `yaml:"systemConfig"`
	UserConfig         string        `yaml:"userConfig"`
	TrustKeysFromHttps bool          `yaml:"trustKeysFromHttps"`
	NoStore            bool          `yaml:"noStore"`
	StoreOnly          bool          `yaml:"storeOnly"`
}

type RktClient struct {
	config     RktConfig
	globalArgs []string
	fields     data.Fields
}

func NewRktClient(config RktConfig) (*RktClient, error) {
	if len(config.InsecureOptions) == 0 {
		config.InsecureOptions = []string{"ondisk", "image"}
	}

	rkt := &RktClient{
		fields:     data.WithField("config", config),
		config:     config,
		globalArgs: config.prepareGlobalArgs(config.InsecureOptions),
	}

	v, err := rkt.Version()
	if err != nil {
		return nil, err
	}
	if v.LessThan(rktSupportedVersion) {
		return nil, errs.WithF(rkt.fields.WithField("current", v).WithField("required", ">="+rktSupportedVersion), "Unsupported version of rkt")
	}

	logs.WithField("version", v).WithField("args", rkt.globalArgs).Debug("New rkt client")
	return rkt, nil
}

func (rktCfg *RktConfig) prepareGlobalArgs(insecureOptions []string) []string {
	args := []string{}

	cmd := "rkt"
	if rktCfg.Path != "" {
		cmd = rktCfg.Path
	}
	args = append(args, cmd)

	if logs.IsDebugEnabled() {
		args = append(args, "--debug")
	}
	if rktCfg.TrustKeysFromHttps {
		args = append(args, "--trust-keys-from-https")
	}
	if rktCfg.UserConfig != "" {
		args = append(args, "--user-config="+rktCfg.UserConfig)
	}
	if rktCfg.LocalConfig != "" {
		args = append(args, "--local-config="+rktCfg.LocalConfig)
	}
	if rktCfg.SystemConfig != "" {
		args = append(args, "--system-config="+rktCfg.SystemConfig)
	}
	if rktCfg.dir != "" {
		args = append(args, "--rkt.dir="+rktCfg.dir)
	}
	args = append(args, "--insecure-options="+strings.Join(insecureOptions, ","))
	return args
}

func (rkt *RktClient) argsStore(cmd []string, globalArgs []string, cmdArgs ...string) []string {
	args := globalArgs[1:]
	args = append(args, cmd...)
	if rkt.config.NoStore {
		args = append(args, "--no-store")
	}
	if rkt.config.StoreOnly {
		args = append(args, "--store-only")
	}
	args = append(args, cmdArgs...)
	return args
}

func (rkt *RktClient) GetPath() (string, error) {
	if rkt.config.Path != "" {
		return rkt.config.Path, nil
	}
	return ExecCmdGetOutput("/bin/bash", "-c", "command -v rkt")
}

func (rkt *RktClient) Version() (Version, error) {
	output, err := ExecCmdGetOutput(rkt.globalArgs[0], "version")
	if err != nil {
		return "", errs.WithEF(err, rkt.fields, "Failed to get rkt Version")
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "rkt Version:") {
			var versionString string
			if read, err := fmt.Sscanf(line, "rkt Version: %s", &versionString); read != 1 || err != nil {
				return "", errs.WithEF(err, rkt.fields.WithField("cpntent", line), "Failed to read rkt version")
			}
			version := Version(versionString)
			return version, nil
		}
	}
	return "", errs.WithF(rkt.fields.WithField("content", output), "Cannot found rkt version from rkt call")
}

func (rkt *RktClient) Fetch(image string) (string, error) {
	hash, err := ExecCmdGetOutput(rkt.globalArgs[0], rkt.argsStore([]string{"fetch"}, rkt.globalArgs, "--full", image)...)
	if err != nil {
		return "", errs.WithEF(err, rkt.fields.WithField("image", image), "Failed to fetch image")
	}
	return hash, err
}

func (rkt *RktClient) FetchInsecure(image string) (string, error) {
	globalArgs := rkt.globalArgs
	if !rkt.config.InsecureOptions.HasImage() {
		globalArgs = rkt.config.prepareGlobalArgs(append(rkt.config.InsecureOptions, "image"))
	}
	hash, err := ExecCmdGetOutput(rkt.globalArgs[0], rkt.argsStore([]string{"fetch"}, globalArgs, "--full", image)...)
	if err != nil {
		return "", errs.WithEF(err, rkt.fields.WithField("image", image), "Failed to fetch image")
	}
	return hash, err
}

func (rkt *RktClient) CatManifest(image string) (string, error) {
	content, err := ExecCmdGetOutput(rkt.globalArgs[0], append(rkt.globalArgs[1:], "image", "cat-manifest", image)...)
	if err != nil {
		return "", errs.WithEF(err, rkt.fields.WithField("image", image), "Failed to cat manifest")
	}
	return content, err
}

func (rkt *RktClient) ImageRm(images string) error {
	stdout, stderr, err := ExecCmdGetStdoutAndStderr(rkt.globalArgs[0], append(rkt.globalArgs[1:], "image", "rm", images)...)
	if err != nil {
		return errs.WithEF(err, rkt.fields.WithField("images", images).WithField("stdout", stdout).WithField("stderr", stderr), "Failed to cat manifest")
	}
	return err
}

func (rkt *RktClient) RmFromFile(path string) (string, string, error) {
	out, stderr, err := ExecCmdGetStdoutAndStderr(rkt.globalArgs[0], append(rkt.globalArgs[1:], "rm", "--uuid-file", path)...)
	if err != nil {
		return "", "", errs.WithEF(err, rkt.fields.WithField("path", path).
			WithField("stdout", out).
			WithField("stderr", stderr), "Failed to remove containers")
	}
	return out, stderr, err
}

func (rkt *RktClient) Rm(uuids string) (string, string, error) {
	out, stderr, err := ExecCmdGetStdoutAndStderr(rkt.globalArgs[0], append(rkt.globalArgs[1:], "rm", uuids)...)
	if err != nil {
		return "", "", errs.WithEF(err, rkt.fields.WithField("uuids", uuids).
			WithField("stdout", out).
			WithField("stderr", stderr), "Failed to remove containers")
	}
	return out, stderr, err
}

func (rkt *RktClient) Run(args []string) error {
	if err := ExecCmd(rkt.globalArgs[0], append(append(rkt.globalArgs[1:], "run"), args...)...); err != nil {
		return errs.WithEF(err, rkt.fields, "Run failed")
	}
	return nil
}
