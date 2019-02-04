package shims

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/cloudfoundry/libbuildpack"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

type Detector interface {
	Detect() error
}

type Finalizer struct {
	V2AppDir        string
	V3AppDir        string
	V2DepsDir       string
	V3LayersDir     string
	V3BuildpacksDir string
	DepsIndex       string
	PlanMetadata    string
	GroupMetadata   string
	ProfileDir      string
	BinDir          string
	Detector        Detector
}

func (f *Finalizer) Finalize() error {
	if err := os.RemoveAll(f.V2AppDir); err != nil {
		return err
	}
	if err := f.IncludePreviousV2Buildpacks(); err != nil {
		return err
	}

	if err := f.MergeOrderTOMLs(); err != nil {
		return err
	}

	//
	//if err := f.Installer.InstallCNBS(f.OrderMetadata, f.V3BuildpacksDir); err != nil {
	//	return err
	//}
	//
	// if err := f.RunV3Detect(); err != nil {
	// 	return err
	// }
	//
	// if err := f.Installer.InstallOnlyVersion("v3-builder", f.BinDir); err != nil {
	// 	return err
	// }
	//
	// if err := f.RunLifeycleBuild(); err != nil {
	// 	return err
	// }
	//
	// if err := f.MoveV3Layers(); err != nil {
	// 	return err
	// }
	//
	//
	// if err := f.Installer.InstallOnlyVersion("v3-launcher", f.V2DepsDir); err != nil {
	// 	return err
	// }
	if err := os.Rename(f.V3AppDir, f.V2AppDir); err != nil {
		return err
	}

	profileContents := fmt.Sprintf(
		`export PACK_STACK_ID="org.cloudfoundry.stacks.%s"
export PACK_LAYERS_DIR="$DEPS_DIR"
export PACK_APP_DIR="$HOME"
exec $DEPS_DIR/v3-launcher "$2"
`,
		os.Getenv("CF_STACK"))

	return ioutil.WriteFile(filepath.Join(f.ProfileDir, "0_shim.sh"), []byte(profileContents), 0666)
}

func (f *Finalizer) MergeOrderTOMLs() error {
	// TOML --> Golang structs
	var tomls []order
	orderFiles, err := ioutil.ReadDir(filepath.Join(f.V2DepsDir, "order"))
	if err != nil {
		return err
	}

	for _, file := range orderFiles {
		orderTOML, err := parseOrderTOML(filepath.Join(f.V2DepsDir, "order", file.Name()))
		if err != nil {
			return err
		}

		tomls = append(tomls, orderTOML)
	}

	// Combine TOMLs
	// TODO: what to do with labels? currently only use the first toml's label
	finalToml := tomls[0]
	finalBuildpacks := &finalToml.Groups[0].Buildpacks

	for i := 1; i < len(tomls); i++ {
		curToml := tomls[i]
		curBuildpacks := curToml.Groups[0].Buildpacks
		*finalBuildpacks = append(*finalBuildpacks, curBuildpacks...)
	}

	// Filter duplicate buildpacks
	for i := range *finalBuildpacks {
		for j := i + 1; j < len(*finalBuildpacks); {
			if (*finalBuildpacks)[i].ID == (*finalBuildpacks)[j].ID {
				*finalBuildpacks = append((*finalBuildpacks)[:j], (*finalBuildpacks)[j+1:]...)
			} else {
				j++
			}
		}
	}

	// Golang structs --> TOML
	finalTomlPath := filepath.Join(f.V2DepsDir, "order.toml")

	return encodeTOML(finalTomlPath, finalToml)
}

func parseOrderTOML(path string) (order, error) {
	var order order
	if _, err := toml.DecodeFile(path, &order); err != nil {
		return order, err
	}
	return order, nil
}

func (f *Finalizer) RunV3Detect() error {
	_, groupErr := os.Stat(f.GroupMetadata)
	_, planErr := os.Stat(f.PlanMetadata)

	if os.IsNotExist(groupErr) || os.IsNotExist(planErr) {
		return f.Detector.Detect()
	}
	return nil
}

func (f *Finalizer) IncludePreviousV2Buildpacks() error {
	myIDx, err := strconv.Atoi(f.DepsIndex)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(filepath.Join(f.V2DepsDir, f.DepsIndex)); err != nil {
		return err
	}
	for supplyDepsIndex := myIDx - 1; supplyDepsIndex >= 0; supplyDepsIndex-- {
		v2Layer := filepath.Join(f.V2DepsDir, strconv.Itoa(supplyDepsIndex))
		buildpackID := fmt.Sprintf("buildpack.%d", supplyDepsIndex)
		v3Layer := filepath.Join(f.V3LayersDir, buildpackID, "layer")
		if err := f.MoveV2Layers(v2Layer, v3Layer); err != nil {
			return err
		}
		if err := f.RenameEnvDir(v3Layer); err != nil {
			return err
		}
		if err := f.UpdateGroupTOML(buildpackID); err != nil {
			return err
		}
		if err := f.AddFakeCNBBuildpack(buildpackID); err != nil {
			return err
		}
	}
	return nil
}

func (f *Finalizer) MoveV3Layers() error {
	bpLayers, err := filepath.Glob(filepath.Join(f.V3LayersDir, "*"))
	if err != nil {
		return err
	}

	for _, bpLayer := range bpLayers {
		if filepath.Base(bpLayer) == "config" {
			if err := os.Rename(filepath.Join(f.V3LayersDir, "config"), filepath.Join(f.V2DepsDir, "config")); err != nil {
				return err
			}

			if err := os.MkdirAll(filepath.Join(f.V2AppDir, ".cloudfoundry"), 0777); err != nil {
				return err
			}

			if err := libbuildpack.CopyFile(filepath.Join(f.V2DepsDir, "config", "metadata.toml"), filepath.Join(f.V2AppDir, ".cloudfoundry", "metadata.toml")); err != nil {
				return err
			}
		} else if err := os.Rename(bpLayer, filepath.Join(f.V2DepsDir, filepath.Base(bpLayer))); err != nil {
			return err
		}
	}

	return nil
}

func (f *Finalizer) RunLifeycleBuild() error {
	cmd := exec.Command(
		filepath.Join(f.BinDir, "v3-builder"),
		"-app", f.V3AppDir,
		"-buildpacks", f.V3BuildpacksDir,
		"-group", f.GroupMetadata,
		"-layers", f.V3LayersDir,
		"-plan", f.PlanMetadata,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "PACK_STACK_ID=org.cloudfoundry.stacks."+os.Getenv("CF_STACK"))

	return cmd.Run()
}

func (f *Finalizer) MoveV2Layers(src, dst string) error {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0777); err != nil {
		return err
	}

	return os.Rename(src, dst)
}

func (f *Finalizer) RenameEnvDir(dst string) error {
	if err := os.Rename(filepath.Join(dst, "env"), filepath.Join(dst, "env.build")); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (f *Finalizer) UpdateGroupTOML(buildpackID string) error {
	var groupMetadata group

	// TODO f.GroupMetadata is an empty string! - crashes here
	// might need to worry about initializing GroupMetadata for finalizer in finalize/main.go?
	if _, err := toml.DecodeFile(f.GroupMetadata, &groupMetadata); err != nil {
		return err
	}

	groupMetadata.Buildpacks = append([]buildpack{{ID: buildpackID}}, groupMetadata.Buildpacks...)

	return encodeTOML(f.GroupMetadata, groupMetadata)
}

func (f *Finalizer) AddFakeCNBBuildpack(buildpackID string) error {
	buildpackPath := filepath.Join(f.V3BuildpacksDir, buildpackID, "latest")
	if err := os.MkdirAll(buildpackPath, 0777); err != nil {
		return err
	}

	buildpackMetadataFile, err := os.Create(filepath.Join(buildpackPath, "buildpack.toml"))
	if err != nil {
		return err
	}
	defer buildpackMetadataFile.Close()

	type buildpack struct {
		ID      string `toml:"id"`
		Name    string `toml:"name"`
		Version string `toml:"version"`
	}
	type stack struct {
		ID string `toml:"id"`
	}

	if err = encodeTOML(filepath.Join(buildpackPath, "buildpack.toml"), struct {
		Buildpack buildpack `toml:"buildpack"`
		Stacks    []stack   `toml:"stacks"`
	}{
		Buildpack: buildpack{
			ID:      buildpackID,
			Name:    buildpackID,
			Version: "latest",
		},
		Stacks: []stack{{
			ID: "org.cloudfoundry.stacks." + os.Getenv("CF_STACK"),
		}},
	}); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(buildpackPath, "bin"), 0777); err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(buildpackPath, "bin", "build"), []byte(`#!/bin/bash`), 0777)
}

func encodeTOML(dest string, data interface{}) error {
	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	return toml.NewEncoder(destFile).Encode(data)
}
