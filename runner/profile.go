package runner

import (
	"errors"
	"fmt"
	"os"
)

const (
	profilePrefix = "_"
	sharedProfile = "shared"
)

type profile struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func toProfiles(source string, names []string) ([]profile, error) {
	profiles := make([]profile, len(names)+1)
	profiles[0] = newProfile(source, sharedProfile)

	for i := 0; i < len(names); i++ {
		p := newProfile(source, names[i])
		if err := tryStat(p.Path); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return nil, err
			}
			continue
		}
		profiles[i+1] = p
	}
	return profiles, nil
}

func newProfile(source string, name string) profile {
	return profile{
		Name: name,
		Path: fmt.Sprintf("%s/%s%s", source, profilePrefix, name),
	}
}
