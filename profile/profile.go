package profile

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
)

const shared = "shared"

type Info struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func New(source string, name string) (Info, error) {
	profilePath := filepath.Join(source, name)
	if _, err := os.Stat(profilePath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return Info{}, fmt.Errorf("failed checking profile exists [%v]: %w", profilePath, err)
		}
		return Info{}, fmt.Errorf("profile [%v] not found at [%v]", name, profilePath)
	}

	return Info{
		Name: name,
		Path: profilePath,
	}, nil
}

func Transform(source string, names []string) []Info {
	pNames := make([]string, 0, len(names))
	if slices.Index(names, shared) == -1 {
		pNames = append(pNames, shared)
	}
	pNames = append(pNames, names...)

	profiles := make([]Info, 0, len(pNames))
	for _, p := range pNames {
		prof, err := New(source, p)
		if err != nil {
			log.Printf("[WARN] %v, ignored: %v", p, err)
			continue
		}
		profiles = append(profiles, prof)
	}
	return profiles
}
