package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveDownloadURLUsesYZboardNodeRelease(t *testing.T) {
	t.Parallel()

	got := resolveDownloadURL("xboard-node-linux-amd64", "v1.13-yz.2")
	want := "https://github.com/P0me1oo/YZboard-Node/releases/download/v1.13-yz.2/xboard-node-linux-amd64"
	if got != want {
		t.Fatalf("resolveDownloadURL() = %q, want %q", got, want)
	}
}

func TestVerifyReleaseChecksum(t *testing.T) {
	t.Parallel()

	payload := []byte("yzboard-node release artifact")
	artifact := "xboard-node-linux-amd64"
	path := filepath.Join(t.TempDir(), artifact)
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(payload)
	checksums := []byte(fmt.Sprintf("%x  %s\n", sum, artifact))
	if err := verifyReleaseChecksum(path, artifact, checksums); err != nil {
		t.Fatalf("verifyReleaseChecksum() error = %v", err)
	}
}

func TestVerifyReleaseChecksumRejectsMismatch(t *testing.T) {
	t.Parallel()

	artifact := "xbctl-linux-arm64"
	path := filepath.Join(t.TempDir(), artifact)
	if err := os.WriteFile(path, []byte("unexpected"), 0o600); err != nil {
		t.Fatal(err)
	}
	checksums := []byte(fmt.Sprintf("%064x  %s\n", 0, artifact))
	if err := verifyReleaseChecksum(path, artifact, checksums); err == nil {
		t.Fatal("verifyReleaseChecksum() error = nil, want checksum mismatch")
	}
}

func TestInstalledVersion(t *testing.T) {
	t.Parallel()

	report := []byte("xboard-node v1.13-yz.2 (built 2026-07-25T00:00:00Z, commit abc1234)\nother details")
	if got := installedVersion("latest", report); got != "v1.13-yz.2" {
		t.Fatalf("installedVersion(latest) = %q", got)
	}
	if got := installedVersion("v1.13-yz.2", []byte("invalid report")); got != "v1.13-yz.2" {
		t.Fatalf("installedVersion(fixed) = %q", got)
	}
}
