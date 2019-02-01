package main

import (
	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/shims"
	"io/ioutil"
	"os"
	"path/filepath"
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

	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	v3LayersDir := filepath.Join(string(filepath.Separator), "home", "vcap", "deps")
	err = os.MkdirAll(v3LayersDir, 0777)
	if err != nil {
		return err
	}
	defer os.RemoveAll(v3LayersDir)

	v3AppDir := filepath.Join(string(filepath.Separator), "home", "vcap", "app")

	v3BuildpacksDir := filepath.Join(tempDir, "cnbs")
	err = os.MkdirAll(v3BuildpacksDir, 0777)
	if err != nil {
		return err
	}
	defer os.RemoveAll(v3BuildpacksDir)

	supplier := shims.Supplier{
		V2AppDir:       v2AppDir,
		V3AppDir:       v3AppDir,
		V2DepsDir:      v2DepsDir,
		DepsIndex:      depsIndex,
		V2BuildpackDir: buildpackDir,
	}

	return supplier.Supply()
}
