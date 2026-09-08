package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	defaultBinDir     = "/usr/local/bin"
	defaultBinDirFile = defaultInstallRoot + "/bin-dir"
	installLockName   = ".install-lock"
)

type installPaths struct {
	binDir string
}

func (p installPaths) binary() string { return path.Join(p.binDir, "xboard-node") }
func (p installPaths) cli() string    { return path.Join(p.binDir, "xbctl") }

// 程序目录会写入 shell 和 systemd 服务文件，只接受无需额外转义的绝对路径。
func validateBinDir(dir string) (string, error) {
	if !strings.HasPrefix(dir, "/") || strings.HasPrefix(dir, "//") {
		return "", errors.New("binary directory must be an absolute Linux path")
	}
	for _, c := range dir {
		if c != '/' && c != '.' && c != '_' && c != '-' &&
			(c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') {
			return "", errors.New("binary directory supports only letters, numbers, /, ., _, and -")
		}
	}
	clean := path.Clean(dir)
	if clean == "/" || clean != strings.TrimRight(dir, "/") {
		return "", errors.New("binary directory must not be / or contain repeated slashes, . or .. components")
	}
	return clean, nil
}

func loadInstallPaths(file string) (installPaths, error) {
	data, err := os.ReadFile(file)
	if errors.Is(err, os.ErrNotExist) {
		return installPaths{binDir: defaultBinDir}, nil
	}
	if err != nil {
		return installPaths{}, fmt.Errorf("read binary directory: %w", err)
	}
	dir, err := validateBinDir(strings.TrimSuffix(string(data), "\n"))
	if err != nil {
		return installPaths{}, fmt.Errorf("invalid %s: %w", file, err)
	}
	return installPaths{binDir: dir}, nil
}

func runConfigBinDir(args []string) error {
	file := defaultBinDirFile
	if len(args) != 0 {
		if len(args) != 2 || args[0] != "--path-file" {
			return errors.New("usage: xbctl config bin-dir [--path-file PATH]")
		}
		file = args[1]
	}
	paths, err := loadInstallPaths(file)
	if err != nil {
		return err
	}
	fmt.Println(paths.binDir)
	return nil
}

// 与安装脚本共用锁目录，避免两个升级过程互相覆盖备份或安装位置。
func lockInstallation(root string) (func(), error) {
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, err
	}
	lock := filepath.Join(root, installLockName)
	if err := os.Mkdir(lock, 0o700); err != nil {
		return nil, fmt.Errorf("cannot lock installation at %s; check for another installer or an interrupted operation: %w", lock, err)
	}
	return func() { _ = os.Remove(lock) }, nil
}
