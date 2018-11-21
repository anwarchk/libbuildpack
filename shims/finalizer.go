package shims

import (
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Finalizer struct {
	DepsDir    string
	DepsIndex  string
	ProfileDir string
}

func (f *Finalizer) Finalize() error {
	files, err := filepath.Glob(filepath.Join(f.DepsDir, f.DepsIndex, "*", "*", "profile.d", "*"))
	if err != nil {
		return err
	}

	for _, file := range files {
		err := os.Rename(file, filepath.Join(f.ProfileDir, filepath.Base(file)))
		if err != nil {
			return err
		}
	}

	BinDirs, err := filepath.Glob(filepath.Join(f.DepsDir, f.DepsIndex, "*", "*", "bin"))
	if err != nil {
		return err
	}

	for i, dir := range BinDirs {
		BinDirs[i] = strings.Replace(dir, filepath.Clean(f.DepsDir), `$DEPS_DIR`, 1)
	}

	script, err := os.OpenFile(filepath.Join(f.ProfileDir, f.DepsIndex+".sh"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	defer script.Close()

	setupPathTemplate, err := template.New("setupPathTemplate").Parse(setupPathContent)
	if err != nil {
		return err
	}

	return setupPathTemplate.Execute(script, BinDirs)
}
