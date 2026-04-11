package dotfile

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/solerf/dtm/common"
	"github.com/solerf/dtm/profile"
)

func Install(sourceDir, targetDir string, profileNames ...string) error {
	if exists := common.MustPathExists(sourceDir); !exists {
		return fmt.Errorf("install: missing %v", sourceDir)
	}

	if exists := common.MustPathExists(targetDir); !exists {
		if errMkdir := os.MkdirAll(targetDir, common.DirMode); errMkdir != nil {
			return fmt.Errorf("install: mkdir: %w", errMkdir)
		}
	}

	profiles := profile.Transform(sourceDir, profileNames)

	ignoreFilter, err := IgnoreFilter(sourceDir)
	if err != nil {
		return fmt.Errorf("install: %w", err)
	}

	filePaths, err := Collect(ignoreFilter, profiles...)
	if err != nil {
		return fmt.Errorf("install: %w", err)
	}

	removeStaleEntries(targetDir)

	dotfiles := batch(targetDir, filePaths...)
	msgTempl := "installed [%v] to [%v]: %v"
	for i := 0; i < len(dotfiles); i++ {
		e := dotfiles[i]
		if err = e.Install(); err != nil {
			log.Printf("[ERROR] "+msgTempl, e.Key, e.TargetPath, err)
			continue
		}
		log.Printf(msgTempl, e.Key, e.TargetPath, "OK")
	}

	if err = newMapping(targetDir, profiles, dotfiles).write(); err != nil {
		return fmt.Errorf("install: %w", err)
	}
	return nil
}

func removeStaleEntries(targetDir string) {
	if common.MustPathExists(targetDir) {
		mapping, err := readMapping(targetDir)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				log.Printf("[ERROR] trying to remove stale entries: %v", err)
			}
			return
		}

		for i := 0; i < len(mapping.DotFiles); i++ {
			dt := mapping.DotFiles[i]

			// it will follow the symlink, if missing returns false
			// means that the linked file does not exist anymore
			// so is safe to remove the link
			if !common.MustPathExists(dt.TargetPath) {
				if err = os.Remove(dt.TargetPath); err != nil {
					log.Printf("[ERROR] trying to remove stale link [%v]: %v", dt.TargetPath, err)
					continue
				}
				log.Printf("removed stale link [%v]", dt.TargetPath)
			}
		}
	}
}
