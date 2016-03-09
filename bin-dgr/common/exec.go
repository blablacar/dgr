package common

import (
	"bytes"
	"github.com/n0rad/go-erlog/logs"
	"os"
	"os/exec"
	"strings"
)

func ExecCmdGetStdoutAndStderr(head string, parts ...string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if logs.IsDebugEnabled() {
		logs.WithField("command", strings.Join([]string{head, " ", strings.Join(parts, " ")}, " ")).Debug("Running external command")
	}
	cmd := exec.Command(head, parts...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Start()
	err := cmd.Wait()
	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), err
}

func ExecCmdGetOutput(head string, parts ...string) (string, error) {
	var stdout bytes.Buffer

	if logs.IsDebugEnabled() {
		logs.WithField("command", strings.Join([]string{head, " ", strings.Join(parts, " ")}, " ")).Debug("Running external command")
	}
	cmd := exec.Command(head, parts...)
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	err := cmd.Wait()
	return strings.TrimSpace(stdout.String()), err
}

func ExecCmdGetStderr(head string, parts ...string) (string, error) {
	var stderr bytes.Buffer

	if logs.IsDebugEnabled() {
		logs.WithField("command", strings.Join([]string{head, " ", strings.Join(parts, " ")}, " ")).Debug("Running external command")
	}
	cmd := exec.Command(head, parts...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr
	cmd.Start()
	err := cmd.Wait()
	return strings.TrimSpace(stderr.String()), err
}

func ExecCmd(head string, parts ...string) error {
	if logs.IsDebugEnabled() {
		logs.WithField("command", strings.Join([]string{head, " ", strings.Join(parts, " ")}, " ")).Debug("Running external command")
	}
	cmd := exec.Command(head, parts...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
