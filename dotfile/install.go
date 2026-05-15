package dotfile

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/solerf/dtm/ignore"
	"github.com/solerf/dtm/profile"
)

func Install(sourceDir, targetDir string, profileNames ...string) ([]OperationStatus, error) {
	log.Printf("installing, %s", strings.Join(profileNames, ", "))
	if sourceDir == targetDir {
		return nil, errors.New("source and target directory cannot be the same")
	}

	if _, err := os.Stat(sourceDir); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("install: %v: %w", sourceDir, err)
		}
		return nil, fmt.Errorf("install: missing %v", sourceDir)
	}

	if err := createIfMissing(targetDir); err != nil {
		return nil, fmt.Errorf("install: %w", err)
	}

	profiles := profile.Transform(sourceDir, profileNames)

	ignoreFilter, err := ignore.FilterFunc(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("install: %w", err)
	}

	log.Println("collecting files...")
	dotfiles, err := collectItemPaths(targetDir, ignoreFilter, profiles...)
	if err != nil {
		return nil, fmt.Errorf("install: %w", err)
	}

	removeStaleEntries(targetDir)

	result := make([]OperationStatus, 0, len(dotfiles))
	dtFilesToMapping := make([]Item, 0, len(dotfiles))
	for _, e := range dotfiles {
		var st Status
		st, err = e.install()
		result = append(result, OperationStatus{Dotfile: &e, Status: st, Error: err})
		if err == nil {
			// discard failed installations
			dtFilesToMapping = append(dtFilesToMapping, e)
		}
	}

	if err = newMapping(targetDir, dtFilesToMapping).write(); err != nil {
		return nil, fmt.Errorf("install: %w", err)
	}
	return result, nil
}

func removeStaleEntries(targetDir string) {
	// if before installing a previous link, which is not pointing to anything is found
	// it is removed as its target probably has been previously removed
	mapping, err := readMapping(targetDir)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Printf("[ERROR] trying to read stale entries: %v", err)
			return
		}
		log.Printf("[WARN] mapping not found, not checking stale entries: %v", err)
		return
	}

	for _, dt := range mapping.DotFiles {
		linkExists, err := dt.linkExists()
		if err != nil {
			log.Printf("[WARN] stale entries, skipping [%s]: %v", dt.Key, err)
			continue
		}

		if linkExists {
			if sourceExist, err := dt.sourceExists(); !sourceExist && err == nil {
				if err = os.Remove(dt.SymLink); err != nil {
					log.Printf("[ERROR] trying to remove stale link [%v]: %v", dt.SymLink, err)
					continue
				}
				log.Printf("removed stale link [%v]", dt.SymLink)
			}
		}
	}
}
