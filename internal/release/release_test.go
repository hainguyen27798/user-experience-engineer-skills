package release_test

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestInstaller(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the release installer targets Linux and macOS")
	}

	script := filepath.Join("..", "..", "scripts", "test-install.sh")
	command := exec.Command("sh", script)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("%s failed: %v\n%s", script, err, output)
	}
}

func TestPublisherResumesFailedUploadWithoutAnotherVersionBump(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the release publisher targets Linux and macOS")
	}

	script := filepath.Join("..", "..", "scripts", "test-release.sh")
	command := exec.Command("sh", script)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("%s failed: %v\n%s", script, err, output)
	}
}
