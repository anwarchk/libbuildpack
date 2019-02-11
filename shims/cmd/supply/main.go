package main

import (
	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/shims"
	"log"
	"os"
	"path/filepath"
	"time"
)

var logger = libbuildpack.NewLogger(os.Stdout)

func init() {
	if len(os.Args) != 5 {
		logger.Error("Incorrect number of arguments")
		os.Exit(1)
	}
}

func main() {
	exit(supply(logger))
}

func exit(err error) {
	if err == nil {
		os.Exit(0)
	}
	logger.Error("Failed supply step: %s", err.Error())
	os.Exit(1)
}

func supply(logger *libbuildpack.Logger) error {
	v2AppDir := os.Args[1]
	v2DepsDir := os.Args[3]
	depsIndex := os.Args[4]

	buildpackDir, err := libbuildpack.GetBuildpackDir()
	if err != nil {
		return err
	}

	v3AppDir := filepath.Join(string(filepath.Separator), "home", "vcap", "app")
	if err := os.MkdirAll(v3AppDir, 0777); err != nil {
		return err
	}

	v3BuildpacksDir := filepath.Join(v2DepsDir, "cnbs")
	err = os.MkdirAll(v3BuildpacksDir, 0777)
	if err != nil {
		return err
	}

	manifest, err := libbuildpack.NewManifest(buildpackDir, logger, time.Now())
	if err != nil {
		log.Fatal(err)
	}
	installer := shims.NewCNBInstaller(manifest)

	supplier := shims.Supplier{
		V2AppDir:        v2AppDir,
		V3AppDir:        v3AppDir,
		V2DepsDir:       v2DepsDir,
		DepsIndex:       depsIndex,
		V2BuildpackDir:  buildpackDir,
		V3BuildpacksDir: v3BuildpacksDir,
		Installer:       installer,
	}

	return supplier.Supply()
}
