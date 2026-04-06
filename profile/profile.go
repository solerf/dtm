package profile

import (
	"fmt"
	"log"
	"path"
	"slices"
	"strings"

	"github.com/solerf/dtm/common"
)

const (
	prefix = "_"
	shared = "shared"
)

type Profiles []Info

func (p Profiles) Paths() []string {
	paths := make([]string, 0, len(p))
	for i := 0; i < len(p); i++ {
		paths = append(paths, p[i].Path)
	}
	return paths
}

type Info struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func New(source string, name string) Info {
	return Info{
		Name: name,
		Path: fmt.Sprintf("%s/%s%s", source, prefix, name),
	}
}

func Transform(source string, names []string) (Profiles, error) {
	profiles := make([]Info, len(names)+1)
	profiles[0] = New(source, shared)

	for i := 0; i < len(names); i++ {
		p := New(source, names[i])
		if exists := common.MustPathExists(p.Path); !exists {
			log.Printf("[WARN] profile [%v] not found at [%v]", p.Name, p.Path)
			continue
		}
		profiles[i+1] = p
	}
	return profiles, nil
}

func RemoveProfile(p string) string {
	split := strings.Split(p, common.PathSeparator)
	profileIndex := slices.IndexFunc(split, func(s string) bool {
		return strings.HasPrefix(s, prefix)
	})
	return path.Join(split[profileIndex+1:]...)
}
