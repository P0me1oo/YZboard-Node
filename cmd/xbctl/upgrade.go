package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type upgradeOperations struct {
	download func(string, string) error
	run      func(string, ...string) ([]byte, error)
	restart  func() error
	rename   func(string, string) error
}

type upgradeFile struct {
	target   string
	staged   string
	backup   string
	hadOld   bool
	replaced bool
}

// 新文件和旧文件的硬链接都位于程序分区，升级最多保存两份二进制内容。
// 失败时只有本次创建的临时目录会被清理；无法回滚的备份会保留供恢复。
func upgradeBinaries(paths installPaths, release string, ops upgradeOperations) (string, error) {
	stage, err := os.MkdirTemp(paths.binDir, ".xboard-node-upgrade-")
	if err != nil {
		return "", fmt.Errorf("create upgrade directory: %w", err)
	}
	keepBackup := false
	defer func() {
		if !keepBackup {
			_ = os.RemoveAll(stage)
		}
	}()

	checksumsPath := filepath.Join(stage, "SHA256SUMS")
	if err := ops.download(resolveDownloadURL("SHA256SUMS", release), checksumsPath); err != nil {
		return "", fmt.Errorf("download release checksums: %w", err)
	}
	checksums, err := os.ReadFile(checksumsPath)
	if err != nil {
		return "", fmt.Errorf("read release checksums: %w", err)
	}
	files := []upgradeFile{
		{target: paths.binary(), staged: filepath.Join(stage, "xboard-node"), backup: filepath.Join(stage, "previous-xboard-node")},
		{target: paths.cli(), staged: filepath.Join(stage, "xbctl"), backup: filepath.Join(stage, "previous-xbctl")},
	}
	var report []byte
	for i := range files {
		name := filepath.Base(files[i].staged)
		artifact := fmt.Sprintf("%s-linux-%s", name, runtime.GOARCH)
		if err := ops.download(resolveDownloadURL(artifact, release), files[i].staged); err != nil {
			return "", fmt.Errorf("download %s: %w", name, err)
		}
		if err := verifyReleaseChecksum(files[i].staged, artifact, checksums); err != nil {
			return "", err
		}
		if err := os.Chmod(files[i].staged, 0o755); err != nil {
			return "", fmt.Errorf("chmod %s: %w", name, err)
		}
		versionArg := "version"
		if i == 0 {
			versionArg = "-v"
		}
		out, err := ops.run(files[i].staged, versionArg)
		if err != nil {
			return "", fmt.Errorf("%s version check failed: %w", name, err)
		}
		if i == 0 {
			report = out
		}
	}
	if paths.binDir != defaultBinDir {
		if _, err := ops.run(files[1].staged, "config", "bin-dir"); err != nil {
			return "", fmt.Errorf("release does not support custom binary directories; use the installer to migrate to %s before downgrading: %w", defaultBinDir, err)
		}
	}

	for i := range files {
		info, err := os.Lstat(files[i].target)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return "", err
		}
		if !info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0 {
			return "", fmt.Errorf("binary path is not a regular file or symlink: %s", files[i].target)
		}
		if err := os.Link(files[i].target, files[i].backup); err != nil {
			return "", fmt.Errorf("create rollback link for %s: %w", files[i].target, err)
		}
		files[i].hadOld = true
	}

	rollback := func(cause error, restart bool) error {
		var restoreErrors []error
		for i := len(files) - 1; i >= 0; i-- {
			if !files[i].replaced {
				continue
			}
			var err error
			if files[i].hadOld {
				err = ops.rename(files[i].backup, files[i].target)
			} else {
				err = os.Remove(files[i].target)
			}
			if err != nil {
				restoreErrors = append(restoreErrors, fmt.Errorf("restore %s: %w", files[i].target, err))
			}
		}
		if len(restoreErrors) > 0 {
			keepBackup = true
			return errors.Join(cause, fmt.Errorf("rollback incomplete; recovery files retained in %s: %w", stage, errors.Join(restoreErrors...)))
		}
		if restart {
			if err := ops.restart(); err != nil {
				return errors.Join(cause, fmt.Errorf("original binaries restored but service restart failed: %w", err))
			}
		}
		return fmt.Errorf("%w; original binaries restored", cause)
	}

	for i := range files {
		if err := ops.rename(files[i].staged, files[i].target); err != nil {
			return "", rollback(fmt.Errorf("replace %s: %w", files[i].target, err), false)
		}
		files[i].replaced = true
	}
	if err := ops.restart(); err != nil {
		return "", rollback(fmt.Errorf("upgrade restart failed: %w", err), true)
	}
	return installedVersion(release, report), nil
}
