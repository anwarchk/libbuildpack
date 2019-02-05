package main

import (
	"errors"
	"github.com/cloudfoundry/libbuildpack"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/cloudfoundry/libbuildpack/shims"
)

func main() {
	logger := libbuildpack.NewLogger(os.Stderr)

	if len(os.Args) != 6 {
		log.Fatal(errors.New("incorrect number of arguments"))
	}

	v2DepsDir := os.Args[3]

	// TODO: do we need to make sure all of Finalizer's fields are initialized?
	detector := shims.DefaultDetector{

	}

	buildpackDir, err := libbuildpack.GetBuildpackDir()
	if err != nil {
		log.Fatal(err) // TODO refactor this
	}


	manifest, err := libbuildpack.NewManifest(buildpackDir, logger, time.Now())
	if err != nil {
		log.Fatal(err)
	}
	installer := shims.NewCNBInstaller(manifest)

	finalizer := shims.Finalizer{
		V2AppDir:   os.Args[1],
		V3AppDir:   filepath.Join(string(filepath.Separator), "home", "vcap", "app"),
		V2DepsDir:  v2DepsDir,
		DepsIndex:  os.Args[4],
		ProfileDir: os.Args[5],
		OrderMetadata: filepath.Join(buildpackDir, "order.toml"),
		Detector:	detector,
		Installer: installer,
}
	if err := finalizer.Finalize(); err != nil {
		log.Fatal(err)
	}
}
