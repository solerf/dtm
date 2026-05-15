package dotfile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	fileMode         = os.FileMode(0644)
	mappingsFileName = ".dtm"
)

type Mapping struct {
	InstallDir string `json:"install_dir"`
	DotFiles   []Item `json:"dotfiles"`
}

func newMapping(targetDir string, dotfiles []Item) *Mapping {
	return &Mapping{
		InstallDir: targetDir,
		DotFiles:   dotfiles,
	}
}

func readMapping(targetDir string) (*Mapping, error) {
	mappingsPath := filepath.Join(targetDir, mappingsFileName)
	file, err := os.ReadFile(mappingsPath)
	if err != nil {
		return nil, fmt.Errorf("reading mappings: %w", err)
	}

	var m Mapping
	if err = json.Unmarshal(file, &m); err != nil {
		return nil, fmt.Errorf("deserializing mappings: %w", err)
	}
	return &m, nil
}

func (m *Mapping) write() error {
	bMappings, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("generating mappings: %w", err)
	}

	finalFile := filepath.Join(m.InstallDir, mappingsFileName)
	tmpFile := finalFile + ".tmp"

	if err = os.WriteFile(tmpFile, bMappings, fileMode); err != nil {
		return fmt.Errorf("serializing mappings: %w", err)
	}

	if err = os.Rename(tmpFile, finalFile); err != nil {
		return fmt.Errorf("writing mappings: %w", err)
	}
	defer os.Remove(tmpFile)

	return nil
}

func (m *Mapping) Delete() error {
	if err := os.Remove(filepath.Join(m.InstallDir, mappingsFileName)); err != nil {
		return fmt.Errorf("delete mappings: %w", err)
	}
	return nil
}
