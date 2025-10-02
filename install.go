package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"maps"
	"os"
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
	return err
}

func processDotFiles(dotfiles []string, sharedPath string, profilePath string, target string) error {
	spWithSlash := sharedPath + "/"
	ppWithSlash := profilePath + "/"

	if _, errTarget := os.Stat(target); errors.Is(errTarget, os.ErrNotExist) {
		if errMkdirTarget := os.MkdirAll(target, fileMode); errMkdirTarget != nil {
			return errMkdirTarget
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
				return "", errMkdir
			}
		}

		log.Printf("linking [%v] => [%v]", dt, targetLink)
		if errLink := os.Symlink(dt, targetLink); errLink != nil {
			return "", errLink
		}
		return targetLink, nil
	}

	mappings := Mappings{
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

func writeMappings(mappings Mappings) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed setting dtm mappings: %w", err)
	}

	mappingsBytes, err := json.Marshal(&mappings)
	if err != nil {
		return fmt.Errorf("failed generating dtm mappings: %w", err)
	}

	err = os.WriteFile(filepath.Join(homeDir, dtmMappingsFileName), mappingsBytes, fileMode)
	if err != nil {
		return fmt.Errorf("failed writing dtm mappings: %w", err)
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
	maps.Insert(result, maps.All(this))
	maps.Insert(result, maps.All(other))
	for _, ig := range ignoreDuplicates {
		delete(result, ig)
	}
	return slices.Collect(maps.Values(result))
}

func collectDirEntries(path string, collectFilters ...func(entry os.DirEntry) bool) map[string]string {
	collected := make(map[string]string)
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Printf("error reading [%v]: %v", path, err)
		return collected
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

	for _, entry := range entries {
		if runFilters(entry) {
			collected[entry.Name()] = filepath.Join(path, entry.Name())
		}
	}
	return collected
}
