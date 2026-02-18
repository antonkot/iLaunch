package system

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

const MinNodeMajor = 18

type CheckResult struct {
	NodePath    string
	NodeVersion string
	PackageMgr  string
}

type Commander interface {
	LookPath(file string) (string, error)
	Output(ctx context.Context, name string, args ...string) ([]byte, error)
}

type ExecCommander struct{}

func (ExecCommander) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func (ExecCommander) Output(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.Output()
}

func CheckEnvironment(ctx context.Context, commander Commander) (CheckResult, error) {
	nodePath, err := commander.LookPath("node")
	if err != nil {
		return CheckResult{}, fmt.Errorf("check node binary: %w", err)
	}

	pkgMgr := ""
	if _, err = commander.LookPath("pnpm"); err == nil {
		pkgMgr = "pnpm"
	} else if _, err = commander.LookPath("npm"); err == nil {
		pkgMgr = "npm"
	} else {
		return CheckResult{}, fmt.Errorf("check package manager: neither pnpm nor npm found")
	}

	args := []string{"--version"}
	if runtime.GOOS == "windows" {
		args = []string{"-v"}
	}
	out, err := commander.Output(ctx, "node", args...)
	if err != nil {
		return CheckResult{}, fmt.Errorf("read node version: %w", err)
	}
	version := strings.TrimSpace(string(out))
	if err = validateNodeVersion(version); err != nil {
		return CheckResult{}, fmt.Errorf("validate node version: %w", err)
	}

	return CheckResult{NodePath: nodePath, NodeVersion: version, PackageMgr: pkgMgr}, nil
}

func validateNodeVersion(version string) error {
	trimmed := strings.TrimPrefix(version, "v")
	parts := strings.Split(trimmed, ".")
	if len(parts) == 0 {
		return fmt.Errorf("invalid node version: %s", version)
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("parse major version: %w", err)
	}
	if major < MinNodeMajor {
		return fmt.Errorf("node version %s is lower than required %d", version, MinNodeMajor)
	}
	return nil
}
