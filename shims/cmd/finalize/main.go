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

	v2DepsDir := os.Args[3]

	// TODO: do we need to make sure all of Finalizer's fields are initialized?
	finalizer := shims.Finalizer{
		V2AppDir:   os.Args[1],
		V3AppDir:   filepath.Join(string(filepath.Separator), "home", "vcap", "app"),
		V2DepsDir:  v2DepsDir,
		DepsIndex:  os.Args[4],
		ProfileDir: os.Args[5],
	}
	if err := finalizer.Finalize(); err != nil {
		log.Fatal(err)
	}
}
