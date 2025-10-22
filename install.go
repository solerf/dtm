package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

var (
	sharedProfile = profilePrefix + "shared"

	fileMode = os.FileMode(0766)

	explodeDirs = []string{".local"}
	isExploded  = func(path string) bool {
		return slices.ContainsFunc(explodeDirs, func(e string) bool {
			return strings.HasSuffix(path, e)
		})
	}

	ignores        = []string{".gitignore"}
	ignorePatterns = []string{"-", "_IGNORE"}

	notIgnoreFilter = func(e os.DirEntry) bool {
		noPatterns := func(name string) bool {
			for _, ip := range ignorePatterns {
				if strings.HasPrefix(name, ip) || strings.HasSuffix(name, ip) {
					return false
				}
			}
			return true
		}
		return !slices.Contains(ignores, e.Name()) && noPatterns(e.Name())
	}
)

func Install(profile string, source string, target string) error {
	sharedPath := fmt.Sprintf("%s/%s", source, sharedProfile)
	profilePath := fmt.Sprintf("%s/%s%s", source, profilePrefix, profile)
	dotfiles := collectDotFiles(sharedPath, profilePath)
	err := processDotFiles(dotfiles, sharedPath, profilePath, target)
	if err != nil {
		err = fmt.Errorf("install: %w", err)
	}
	return err
}

func processDotFiles(dotfiles []string, sharedPath string, profilePath string, target string) error {
	spWithSlash := sharedPath + "/"
	ppWithSlash := profilePath + "/"

	if _, errTarget := os.Stat(target); errors.Is(errTarget, os.ErrNotExist) {
		if errMkdirTarget := os.MkdirAll(target, fileMode); errMkdirTarget != nil {
			return fmt.Errorf("mkdir target: %w", errMkdirTarget)
		}
	}

	makeLink := func(dt string) (string, error) {
		relative := strings.ReplaceAll(strings.ReplaceAll(dt, spWithSlash, ""), ppWithSlash, "")
		relativeDirs := filepath.Dir(relative)
		relativeBase := filepath.Base(relative)
		targetLinkDir := filepath.Join(target, relativeDirs)
		targetLink := filepath.Join(targetLinkDir, relativeBase)

		if _, e := os.Stat(targetLink); e == nil {
			log.Printf("already exists, skipping [%v] => [%v]", dt, targetLink)
			return targetLink, nil
		}

		if _, errLink := os.Stat(targetLinkDir); os.IsNotExist(errLink) {
			log.Printf("making dir [%v]", targetLinkDir)
			if errMkdir := os.MkdirAll(targetLinkDir, fileMode); errMkdir != nil {
				return "", fmt.Errorf("mkdir all: %w", errMkdir)
			}
		}

		log.Printf("linking [%v] => [%v]", dt, targetLink)

		// if link already exists
		if _, errLstat := os.Lstat(targetLink); errLstat == nil {
			// check if points to same target that is trying to map
			if foundLinkTarget, errRead := os.Readlink(targetLink); errRead == nil && foundLinkTarget != dt {
				// if different it is removed to then be recreated
				if errRem := os.Remove(targetLink); errRem != nil {
					return "", fmt.Errorf("remove existent: %w", errRem)
				}
			}
		}

		if errLink := os.Symlink(dt, targetLink); errLink != nil {
			return "", fmt.Errorf("linking: %w", errLink)
		}
		return targetLink, nil
	}

	mappings := LocalMappings{
		Profile:    strings.ReplaceAll(path.Base(profilePath), "_", ""),
		SourcesDir: []string{sharedPath, profilePath},
		InstallDir: target,
		Entries:    make(map[string]string, len(dotfiles)),
	}

	for _, d := range dotfiles {
		link, err := makeLink(d)
		if err != nil {
			return err
		}
		mappings.Entries[strings.ReplaceAll(link, target+"/", "")] = d
	}
	return writeMappings(mappings)
}

func writeMappings(mappings LocalMappings) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get homedir: %w", err)
	}

	mappingsBytes, err := json.Marshal(&mappings)
	if err != nil {
		return fmt.Errorf("generating mappings: %w", err)
	}

	err = os.WriteFile(filepath.Join(homeDir, dtmMappingsFileName), mappingsBytes, fileMode)
	if err != nil {
		return fmt.Errorf("writing mappings: %w", err)
	}
	return nil
}

func collectDotFiles(sharedPath string, profilePath string) []string {
	collectFiles := func(thisPath string, thatPath string) ([]string, []string) {
		thisEntries := collectDirEntries(thisPath, notIgnoreFilter)
		thatEntries := collectDirEntries(thatPath, notIgnoreFilter)
		duplicates := intersectKeys(thisEntries, thatEntries)
		uniquesToAdd := mergeToSlice(thisEntries, thatEntries, duplicates)

		for i, u := range uniquesToAdd {
			if isExploded(u) {
				_ = filepath.WalkDir(u, func(path string, d fs.DirEntry, err error) error {
					if !d.IsDir() {
						uniquesToAdd = append(uniquesToAdd, path)
					}
					return nil
				})
				uniquesToAdd = slices.Delete(uniquesToAdd, i, i+1)
			}
		}
		return uniquesToAdd, duplicates
	}

	uniques, duplicatesQueue := collectFiles(sharedPath, profilePath)
	for len(duplicatesQueue) > 0 {
		duplicate := duplicatesQueue[0] // pop

		uniquesToAdd, duplicatesFound := collectFiles(filepath.Join(sharedPath, duplicate), filepath.Join(profilePath, duplicate))
		uniques = append(uniques, uniquesToAdd...)

		for i := 0; i < len(duplicatesFound); i++ {
			duplicatesFound[i] = filepath.Join(duplicate, duplicatesFound[i])
		}

		// push
		duplicatesQueue = append(duplicatesQueue[1:], duplicatesFound...)
	}
	return uniques
}

func intersectKeys(this map[string]string, other map[string]string) []string {
	equals := make([]string, 0, len(this)+len(other))
	for entry := range this {
		if _, ok := other[entry]; ok {
			equals = append(equals, entry)
		}
	}
	return equals
}

func mergeToSlice(this map[string]string, other map[string]string, ignoreDuplicates []string) []string {
	result := make(map[string]string, len(this)+len(other))
	addFiltered := func(m map[string]string) {
		for k, v := range m {
			if !slices.Contains(ignoreDuplicates, k) {
				result[k] = v
			}
		}
	}

	addFiltered(this)
	addFiltered(other)
	return slices.Collect(maps.Values(result))
}

func collectDirEntries(path string, collectFilters ...func(entry os.DirEntry) bool) map[string]string {
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Printf("error reading [%v]: %v", path, err)
		return nil
	}

	runFilters := func(entry os.DirEntry) bool {
		if len(collectFilters) != 0 {
			for _, f := range collectFilters {
				if !f(entry) {
					return false
				}
			}
		}
		return true
	}

	collected := make(map[string]string, 10)
	for _, entry := range entries {
		if runFilters(entry) {
			collected[entry.Name()] = filepath.Join(path, entry.Name())
		}
	}
	return collected
}
