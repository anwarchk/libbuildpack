package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/shims"
)

func main() {
	if len(os.Args) != 6 {
		log.Fatal(errors.New("incorrect number of arguments"))
	}

	buildpackDir, err := filepath.Abs(filepath.Join(os.Args[0], "..", ".."))
	if err != nil {
		log.Fatal(err)
	}

	appDir := os.Args[1]
	depsDir := os.Args[3]
	depsIndex := os.Args[4]
	profileDir := os.Args[5]
	launchDir := filepath.Join(depsDir, depsIndex)

	finalizer := shims.Finalizer{
		BinDir:     filepath.Join(buildpackDir, "bin"),
		AppDir:     appDir,
		DepsIndex:  depsIndex,
		LaunchDir:  launchDir,
		ProfileDir: profileDir,
	}
	err = finalizer.Finalize()
	if err != nil {
		log.Fatal(err)
	}
}
