package runner

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

func collectDotFiles(notIgnoreF func(string) bool, profiles ...profile) ([]dotFile, error) {
	uniqueDotFiles := func(paths []string) ([]dotFile, error) {
		uniques, duplicatesQueue, err := collectFromPaths(notIgnoreF, paths...)
		if err != nil {
			return nil, err
		}

		for len(duplicatesQueue) > 0 {
			dupPaths := extractPaths(duplicatesQueue) // pop

			moreUniques, moreDuplicates, _ := collectFromPaths(notIgnoreF, dupPaths...)
			uniques = append(uniques, moreUniques...)

			duplicatesQueue = moreDuplicates // push
		}
		return uniques, nil
	}

	paths := make([]string, 0, len(profiles))
	for _, p := range profiles {
		paths = append(paths, p.Path)
	}

	return uniqueDotFiles(paths)
}

func collectFromPaths(notIgnoreF func(string) bool, paths ...string) ([]dotFile, []dotFile, error) {
	entries := make(map[string][]dotFile, len(paths))
	for _, p := range paths {

		collected, err := collectDirEntries(notIgnoreF, p)
		if err != nil {
			return nil, nil, err
		}

		for i := 0; i < len(collected); i++ {
			fullpath := path.Join(p, collected[i])
			dtf := newDotFile(fullpath)

			if entry, ok := entries[dtf.Key]; ok {
				entries[dtf.Key] = append(entry, dtf)
				continue
			}

			if dtf.isExploded() {
				err = filepath.WalkDir(dtf.SourcePath, func(childPath string, childEntry fs.DirEntry, err error) error {
					if err != nil {
						return fmt.Errorf("in walk [%v]: %w", childPath, err)
					}

					if !childEntry.IsDir() {
						d := newDotFile(childPath)
						entries[d.Key] = []dotFile{d}
					}
					return nil
				})

				if err != nil {
					return nil, nil, fmt.Errorf("walking [%v]: %w", dtf.SourcePath, err)
				}
				continue
			}

			entries[dtf.Key] = []dotFile{dtf}
		}
	}

	uniques := make([]dotFile, 0, len(entries))
	duplicates := make([]dotFile, 0, len(entries))

	for _, p := range entries {
		if len(p) == 1 {
			uniques = append(uniques, p[0])
		} else {
			duplicates = append(duplicates, p...)
		}
	}
	return uniques, duplicates, nil
}

func collectDirEntries(notIgnoreF func(string) bool, path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("collecting dirs [%v]: %v", path, err)
	}

	collected := make([]string, 0, len(entries))
	var name string
	for i := 0; i < len(entries); i++ {
		name = entries[i].Name()
		if notIgnoreF(name) {
			collected = append(collected, name)
		}
	}
	return collected, nil
}
