package dotfile

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/solerf/dtm/profile"
)

var explodeDirs = []string{".local"}

func Collect(ignoreFunc func(string) bool, profiles ...profile.Info) ([]string, error) {
	paths := make([]string, 0, len(profiles))
	for i := 0; i < len(profiles); i++ {
		paths = append(paths, profiles[i].Path)
	}

	collectedPaths, err := collectDistinct(ignoreFunc, paths...)
	if err != nil {
		return nil, fmt.Errorf("collector: %w", err)
	}
	return slices.Clip(collectedPaths), nil
}

func collectDistinct(ignoreFunc func(string) bool, paths ...string) ([]string, error) {
	uniques, duplicatesQueue, err := collectPaths(ignoreFunc, paths...)
	if err != nil {
		return nil, err
	}

	for len(duplicatesQueue) > 0 { // pop
		moreUniques, moreDuplicates, err := collectPaths(ignoreFunc, duplicatesQueue...)
		if err != nil {
			return nil, err
		}

		uniques = append(uniques, moreUniques...)
		duplicatesQueue = moreDuplicates // push
	}
	return uniques, nil
}

func collectPaths(ignoreFunc func(string) bool, paths ...string) ([]string, []string, error) {
	entries := make(map[string][]string, len(paths))

	for _, p := range paths {
		collected, err := collectDirEntries(ignoreFunc, p)
		if err != nil {
			return nil, nil, err
		}

		for i := 0; i < len(collected); i++ {
			fullPath := path.Join(p, collected[i])
			key := profile.RemoveProfile(fullPath)

			if isExplodedDir(fullPath) {
				err = filepath.WalkDir(fullPath, func(childPath string, childEntry fs.DirEntry, err error) error {
					if err != nil {
						return fmt.Errorf("in walk [%v]: %w", childPath, err)
					}

					if !childEntry.IsDir() && !ignoreFunc(childEntry.Name()) {
						entries[key] = append(entries[key], childPath)
					}
					return nil
				})

				if err != nil {
					return nil, nil, fmt.Errorf("walking [%v]: %w", fullPath, err)
				}
				continue
			}
			entries[key] = append(entries[key], fullPath)
		}
	}

	uniques := make([]string, 0, len(paths))
	duplicates := make([]string, 0, len(paths))

	for _, v := range entries {
		if len(v) == 1 {
			uniques = append(uniques, v...)
			continue
		}
		duplicates = append(duplicates, v...)
	}

	return slices.Clip(uniques), slices.Clip(duplicates), nil
}

func isExplodedDir(p string) bool {
	return slices.ContainsFunc(explodeDirs, func(e string) bool {
		return strings.HasSuffix(p, e)
	})
}

func collectDirEntries(ignoreFunc func(string) bool, profilePath string) ([]string, error) {
	entries, err := os.ReadDir(profilePath)
	if err != nil {
		return nil, fmt.Errorf("reading dir [%v]: %v", profilePath, err)
	}

	collected := make([]string, 0, len(entries))
	for i := 0; i < len(entries); i++ {
		name := entries[i].Name()
		if !ignoreFunc(name) {
			collected = append(collected, name)
		}
	}
	return collected, nil
}
