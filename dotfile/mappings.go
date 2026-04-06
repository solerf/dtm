package dotfile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/solerf/dtm/common"
	"github.com/solerf/dtm/profile"
)

const mappingsFileName = ".dtm"

type Mapping struct {
	InstallDir string         `json:"install_dir"`
	Profiles   []profile.Info `json:"profiles"`
	DotFiles   []entry        `json:"dotfiles"`
}

func newMapping(targetDir string, profiles profile.Profiles, dotfiles []entry) *Mapping {
	return &Mapping{
		InstallDir: targetDir,
		Profiles:   profiles,
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

	m.DotFiles = slices.Clip(m.DotFiles)
	return &m, err
}

func (m *Mapping) write() error {
	bMappings, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("generating mappings: %w", err)
	}

	if err = os.WriteFile(filepath.Join(m.InstallDir, mappingsFileName), bMappings, common.FileMode); err != nil {
		return fmt.Errorf("serializing mappings: %w", err)
	}
	return nil
}

func (m *Mapping) Delete() error {
	if err := os.RemoveAll(filepath.Join(m.InstallDir, mappingsFileName)); err != nil {
		return fmt.Errorf("delete mappings: %w", err)
	}

	return nil
}
