package beads

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// installMockBDStderrOnly puts a bd on PATH that writes a line to stderr,
// nothing to stdout, and exits with MOCK_BD_EXIT for every command.
func installMockBDStderrOnly(t *testing.T, exitCode string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("mock bd is a POSIX shell script")
	}
	binDir := t.TempDir()
	script := "#!/bin/sh\necho 'Warning: failed to setup Claude hooks: home directory: $HOME is not defined' >&2\nexit \"$MOCK_BD_EXIT\"\n"
	if err := os.WriteFile(filepath.Join(binDir, "bd"), []byte(script), 0755); err != nil {
		t.Fatalf("write mock bd: %v", err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("MOCK_BD_EXIT", exitCode)
}

// bd init --quiet prints nothing on stdout; warnings it writes to stderr while
// exiting 0 (bd keeps installer failures and its first-run metrics notice on
// stderr under --quiet) must not fail Init.
func TestInit_StderrWarningsWithExitZeroSucceed(t *testing.T) {
	installMockBDStderrOnly(t, "0")
	if err := NewIsolated(t.TempDir()).Init("tt"); err != nil {
		t.Fatalf("Init with exit 0 and stderr warnings: %v", err)
	}
}

func TestInit_NonZeroExitFails(t *testing.T) {
	installMockBDStderrOnly(t, "1")
	err := NewIsolated(t.TempDir()).Init("tt")
	if err == nil {
		t.Fatal("Init with exit 1: want error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to setup Claude hooks") {
		t.Errorf("error should carry bd's stderr, got %v", err)
	}
}

// Other commands keep treating "exit 0, empty stdout, stderr" as a failure.
func TestRun_StderrOnlyExitZeroStillFailsForNonInit(t *testing.T) {
	installMockBDStderrOnly(t, "0")
	_, err := NewIsolated(t.TempDir()).run("show", "tt-1", "--json")
	if err == nil {
		t.Fatal("show with exit 0 and stderr only: want error, got nil")
	}
}
