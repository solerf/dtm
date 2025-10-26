package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const fileMode = os.FileMode(0766)

var mappingsFileName = ".dtm_mappings"

type mapping struct {
	InstallDir string    `json:"install_dir"`
	Profiles   []profile `json:"profiles"`
	DotFiles   []dotFile `json:"dotfiles"`
}

func writeMappings(homeDir string, installDir string, profiles []profile, dotfiles []dotFile) error {
	mappings := &mapping{
		InstallDir: installDir,
		Profiles:   profiles,
		DotFiles:   dotfiles,
	}

	bMappings, err := json.Marshal(&mappings)
	if err != nil {
		return fmt.Errorf("generating mappings: %w", err)
	}

	if err = os.WriteFile(filepath.Join(homeDir, mappingsFileName), bMappings, fileMode); err != nil {
		return fmt.Errorf("writing mappings: %w", err)
	}
	return nil
}

func readMappings(homeDir string) (*mapping, error) {
	mappingsPath := filepath.Join(homeDir, mappingsFileName)
	file, err := os.ReadFile(mappingsPath)
	if err != nil {
		return nil, fmt.Errorf("reading mappings: %w", err)
	}

	var m mapping
	if err = json.Unmarshal(file, &m); err != nil {
		return nil, fmt.Errorf("deserializing mappings: %w", err)
	}
	return &m, err
}

func deleteMappings(homeDir string) error {
	if err := os.RemoveAll(filepath.Join(homeDir, mappingsFileName)); err != nil {
		return fmt.Errorf("delete mappings: %w", err)
	}
	return nil
}
