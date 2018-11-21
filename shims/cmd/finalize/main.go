package main

import (
	"errors"
	"log"
	"os"

	"github.com/cloudfoundry/libbuildpack/shims"
)

func main() {
	if len(os.Args) != 6 {
		log.Fatal(errors.New("incorrect number of arguments"))
	}

	depsDir := os.Args[3]
	depsIndex := os.Args[4]
	profileDir := os.Args[5]

	err := shims.Finalize(depsDir, depsIndex, profileDir)
	if err != nil {
		log.Fatal(err)
	}
}
