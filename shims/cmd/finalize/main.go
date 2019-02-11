package main

import (
	"errors"
	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/shims"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	logger := libbuildpack.NewLogger(os.Stderr)

	if len(os.Args) != 6 {
		log.Fatal(errors.New("incorrect number of arguments"))
	}

	v2DepsDir := os.Args[3]

	buildpackDir, err := libbuildpack.GetBuildpackDir()
	if err != nil {
		log.Fatal(err)
	}

	manifest, err := libbuildpack.NewManifest(buildpackDir, logger, time.Now())
	if err != nil {
		log.Fatal(err)
	}

	installer := shims.NewCNBInstaller(manifest)

	v3AppDir := filepath.Join(string(filepath.Separator), "home", "vcap", "app")
	v3LayersDir := filepath.Join(string(filepath.Separator), "home", "vcap", "deps")
	if err := os.MkdirAll(v3LayersDir, 0777); err != nil {
		log.Fatal(err)
	}

	v3BuildpacksDir := filepath.Join(v2DepsDir, "cnbs")
	orderMetadata := filepath.Join(v2DepsDir, "order.toml")
	groupMetadata := filepath.Join(v2DepsDir, "group.toml")
	planMetadata := filepath.Join(v2DepsDir, "plan.toml")

	detector := shims.DefaultDetector{
		V3LifecycleDir:  filepath.Join(v2DepsDir, "lifecycle"),
		AppDir:          v3AppDir,
		V3BuildpacksDir: v3BuildpacksDir,
		OrderMetadata:   orderMetadata,
		GroupMetadata:   groupMetadata,
		PlanMetadata:    planMetadata,
		Installer:       installer,
	}

	finalizer := shims.Finalizer{
		V3LifecycleDir:   filepath.Join(v2DepsDir, "lifecycle"),
		V2AppDir:         os.Args[1],
		V3AppDir:         v3AppDir,
		V3BuildpacksDir: v3BuildpacksDir,
		V3LayersDir:      v3LayersDir,
		V2DepsDir:        v2DepsDir,
		DepsIndex:        os.Args[4],
		ProfileDir:       os.Args[5],
		OrderMetadataDir: orderMetadata,
		GroupMetadata:    groupMetadata,
		PlanMetadata:     planMetadata,
		Detector:         detector,
		Installer:        installer,
	}
	if err := finalizer.Finalize(); err != nil {
		log.Fatal(err)
	}
}
