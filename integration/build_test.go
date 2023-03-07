package integration

import (
	"fmt"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/mt-sre/devkube/dev"
	"github.com/stretchr/testify/assert"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

const ()

var (
	projectRoot string
	cacheDir    string
	testDataDir string
)

func init() {
	dir, err := filepath.Abs("..")
	if err != nil {
		panic(err)
	}
	projectRoot = dir
	cacheDir = filepath.Join(projectRoot, ".cache/test-stub")
	testDataDir = filepath.Join(projectRoot, "integration/test-data")
}

func buildBinary() error {
	args := []string{"build", filepath.Join(testDataDir, "test-stub/main.go")}
	cmd := exec.Command("go", args...)
	cmd.Dir = testDataDir
	return cmd.Run()
}

func cleanCacheDir() error {
	if err := os.RemoveAll(cacheDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("deleting cache: %w", err)
	}
	if err := os.Remove(cacheDir + ".tar"); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("deleting image cache: %w", err)
	}
	if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
		return fmt.Errorf("creating cache dir: %w", err)
	}
	return nil
}

func populateCacheDir() error {
	if err := sh.Copy(filepath.Join(cacheDir, "main"),
		filepath.Join(testDataDir, "main")); err != nil {
		return fmt.Errorf("copying binary: %w", err)
	}
	if err := sh.Copy(filepath.Join(cacheDir, "passwd"),
		filepath.Join(testDataDir, "passwd")); err != nil {
		return fmt.Errorf("copying passwd: %w", err)
	}
	if err := sh.Copy(filepath.Join(cacheDir, "test-stub.Containerfile"),
		filepath.Join(testDataDir, "test-stub.Containerfile")); err != nil {
		return fmt.Errorf("copying Containerfile: %w", err)
	}
	return nil
}

func TestBuildImage(t *testing.T) {
	runtime, err := dev.DetectContainerRuntime()
	if err != nil {
		t.Fatal(err)
	}

	deps := []interface{}{
		mg.F(buildBinary),
		mg.F(cleanCacheDir),
		mg.F(populateCacheDir),
	}

	buildInfo := dev.ImageBuildInfo{
		ImageTag:      "test-stub",
		CacheDir:      cacheDir,
		ContainerFile: "test-stub.Containerfile",
		ContextDir:    ".",
		Runtime:       string(runtime),
	}

	err = dev.BuildImage(&buildInfo, deps)
	fmt.Println(err)
	assert.NoError(t, err)

	// TODO: test that image is present and correctly tagged
}
