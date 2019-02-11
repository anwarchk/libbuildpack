package main

import (
	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/shims"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

func main() {
	logger := libbuildpack.NewLogger(os.Stderr)

	if len(os.Args) != 2 {
		logger.Error("Incorrect number of arguments")
		os.Exit(1)
	}

	v2AppDir := os.Args[1]

	tempDir, err := ioutil.TempDir("","temp")
	if err != nil {
		logger.Error("Unable to create temp dir: %s", err.Error())
		os.Exit(1)
	}

	v2BuildpackDir, err := libbuildpack.GetBuildpackDir()
	if err != nil {
		logger.Error("Unable to find buildpack directory: %s", err.Error())
		os.Exit(1)
	}

	//TODO: IS THIS THE RIGHT DIR TO PUT THIS?
	metadataDir, err := filepath.Abs(filepath.Join(v2AppDir, ".."))
	if err != nil {
		logger.Error("Unable to find workspace directory: %s", err.Error())
		os.Exit(1)
	}

	manifest, err := libbuildpack.NewManifest(v2BuildpackDir, logger, time.Now())
	if err != nil {
		logger.Error("Unable to load buildpack manifest: %s", err.Error())
		os.Exit(1)
	}

	detector := shims.DefaultDetector{
		V3LifecycleDir: tempDir,

		AppDir: v2AppDir,

		V3BuildpacksDir: tempDir,

		OrderMetadata: filepath.Join(v2BuildpackDir, "order.toml"),
		GroupMetadata: filepath.Join(metadataDir, "group.toml"),
		PlanMetadata:  filepath.Join(metadataDir, "plan.toml"),

		Installer: shims.NewCNBInstaller(manifest),
	}

	err = detector.Detect()
	if err != nil {
		logger.Error("Failed detection step: %s", err.Error())
		os.Exit(1)
	}
}
