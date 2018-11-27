package shims

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Finalizer struct {
	AppDir     string
	BinDir     string
	DepsIndex  string
	LaunchDir  string
	ProfileDir string
}

func (f *Finalizer) Finalize() error {
	bash := fmt.Sprintf(`
export PACK_STACK_ID="org.cloudfoundry.stacks.%s"
export PACK_LAUNCH_DIR="$DEPS_DIR/%s"
export PACK_APP_DIR="$HOME"
echo "BEFORE!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
exec $DEPS_DIR/v3-launcher
echo "RETURNED!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
`, os.Getenv("CF_STACK"), f.DepsIndex)

	return ioutil.WriteFile(filepath.Join(f.ProfileDir, "0_shim.sh"), []byte(bash), 0666)
}
