package main

import (
	"errors"
	"log"
	"os"

	"github.com/cloudfoundry/libbuildpack/shims"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal(errors.New("incorrect number of arguments"))
	}

	buildDir := os.Args[1]

	if err := shims.Release(buildDir, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
