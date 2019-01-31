package shims

import (
	"fmt"
	"github.com/cloudfoundry/libbuildpack"
	"os"
	"path/filepath"
)

type Supplier struct {
	V2AppDir       string
	V3AppDir        string
	V2DepsDir      string
	DepsIndex      string
	V2BuildpackDir string
}

const (
	ERROR_FILE = "Error V2 Buildpack After V3 Buildpack"
	SENTINEL = "sentinel"
)

func (s *Supplier) Supply() error {
	if err := s.SetUpFirstV3Buildpack(); err != nil {
		return err
	}

	return s.SaveOrderToml()
}

func (s *Supplier) SetUpFirstV3Buildpack() error {
	fi, err := os.Lstat(s.V2AppDir)
	if err != nil {
		return err
	}
	
	if fi.Mode() & os.ModeSymlink == os.ModeSymlink {
		return nil
	}

	if err := os.Rename(s.V2AppDir, s.V3AppDir); err != nil {
		return err
	}

	if err := os.Symlink(ERROR_FILE, s.V2AppDir); err != nil {
		return err
	}

	cfPath := filepath.Join(s.V3AppDir, ".cloudfoundry")
	if err := os.MkdirAll(cfPath, 0777); err != nil {
		fmt.Println("could not open the cloudfoundry dir")
		return err
	}

	if _, err := os.OpenFile(filepath.Join(cfPath, SENTINEL), os.O_RDONLY|os.O_CREATE, 0666); err != nil {
		return err
	}

	return nil
}

func (s *Supplier) SaveOrderToml() error {
	orderDir := filepath.Join(s.V2DepsDir, "order")
	if err := os.MkdirAll(orderDir, 0777); err != nil {
		return err
	}

	return libbuildpack.CopyFile(filepath.Join(s.V2BuildpackDir, "order.toml"), filepath.Join(orderDir, fmt.Sprintf("order%s.toml", s.DepsIndex)))
}