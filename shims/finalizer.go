package shims

import (
	"os"
	"os/exec"
	"path/filepath"
)

type Finalizer struct {
	AppDir       string
	BinDir       string
	LaunchDir    string
	MetadataPath string
}

func (f *Finalizer) Finalize() error {
	cmd := exec.Command(
		filepath.Join(f.BinDir, "v3-launcher"),
		//"-app", f.AppDir,
		//"-launch", f.LaunchDir,
		//"-metadata", f.MetadataPath,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	//cmd.Env = append(os.Environ(), "PACK_STACK_ID=org.cloudfoundry.stacks."+os.Getenv("CF_STACK"))
	return cmd.Run()
	//files, err := filepath.Glob(filepath.Join(f.DepsDir, f.DepsIndex, "*", "*", "profile.d", "*"))
	//if err != nil {
	//	return err
	//}
	//
	//for _, file := range files {
	//	err := os.Rename(file, filepath.Join(f.ProfileDir, filepath.Base(file)))
	//	if err != nil {
	//		return err
	//	}
	//}
	//
	//BinDirs, err := filepath.Glob(filepath.Join(f.DepsDir, f.DepsIndex, "*", "*", "bin"))
	//if err != nil {
	//	return err
	//}
	//
	//for i, dir := range BinDirs {
	//	BinDirs[i] = strings.Replace(dir, filepath.Clean(f.DepsDir), `$DEPS_DIR`, 1)
	//}
	//
	//script, err := os.OpenFile(filepath.Join(f.ProfileDir, f.DepsIndex+".sh"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	//if err != nil {
	//	return err
	//}
	//defer script.Close()
	//
	//setupPathTemplate, err := template.New("setupPathTemplate").Parse(setupPathContent)
	//if err != nil {
	//	return err
	//}
	//
	//return setupPathTemplate.Execute(script, BinDirs)
}
