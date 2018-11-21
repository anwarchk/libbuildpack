package shims

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v2"
)

const setupPathContent = "export PATH={{ range $_, $path := . }}{{ $path }}:{{ end }}$PATH"

type Releaser struct {
	BuildDir string
	Writer   io.Writer
}

type inputMetadata struct {
	Processes []struct {
		Type    string
		Command string
	}
}

type defaultProcessTypes struct {
	Web string `yaml:"web"`
}

type outputMetadata struct {
	DefaultProcessTypes defaultProcessTypes `yaml:"default_process_types"`
}

func (i *inputMetadata) findCommand(processType string) (string, error) {
	for _, p := range i.Processes {
		if p.Type == processType {
			return p.Command, nil
		}
	}
	return "", fmt.Errorf("unable to find process with type %s in launch metadata", processType)
}

func (r *Releaser) Release() error {
	metadataFile, input := filepath.Join(r.BuildDir, ".cloudfoundry", "metadata.toml"), inputMetadata{}
	_, err := toml.DecodeFile(metadataFile, &input)

	defer os.Remove(metadataFile)

	webCommand, err := input.findCommand("web")
	if err != nil {
		return err
	}

	output := outputMetadata{DefaultProcessTypes: defaultProcessTypes{Web: webCommand}}
	return yaml.NewEncoder(r.Writer).Encode(output)
}
