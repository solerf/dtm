package dotfile

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	sysDirs = []string{".config", ".local/bin"}
)

func Uninstall(targetDir string) ([]OperationStatus, error) {
	log.Printf("uninstalling from [%s]", targetDir)
	mapping, err := readMapping(targetDir)
	if err != nil {
		return nil, fmt.Errorf("uninstall: %w", err)
	}

	result := make([]OperationStatus, 0, len(mapping.DotFiles))
	remainingDirs := make(map[string]*Item)

	log.Println("collecting...")
	for _, e := range mapping.DotFiles {
		st, errUninstall := e.uninstall()
		if errUninstall == nil {
			// collect any remaining dir previously installed
			if minSysDir := findSysDir(e.SymLink); minSysDir != "" {
				remainingDirs[minSysDir] = &e
			}
		}
		result = append(result, OperationStatus{
			Dotfile: &e,
			Status:  st,
			Error:   errUninstall,
		})
	}

	for minDir, item := range remainingDirs {
		if err = pruneEmptyDirs(minDir); err != nil {
			result = append(result, OperationStatus{
				Dotfile: item,
				Status:  Failed,
				Error:   fmt.Errorf("remaining dir [%s]: %w", minDir, err),
			})
		}
	}

	if err = mapping.Delete(); err != nil {
		return nil, fmt.Errorf("uninstall: mapping: %w", err)
	}
	return result, nil
}

// pruneEmptyDirs walks root bottom-up and removes any directory that is empty
// after dtm-managed symlinks were removed. Directories containing untracked
// files (or non-empty subtrees) are left intact.
func pruneEmptyDirs(root string) error {
	info, err := os.Lstat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return nil
	}

	dirs := make([]string, 0, 3)
	err = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			dirs = append(dirs, p)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// remove deepest first; os.Remove only succeeds on empty directories,
	// so any dir still holding untracked files is preserved.
	for i := len(dirs) - 1; i >= 0; i-- {
		_ = os.Remove(dirs[i])
	}
	return nil
}

func findSysDir(symlink string) string {
	for _, sd := range sysDirs {
		index := strings.Index(symlink, sd)
		if index != -1 {
			var i int
			sepCount := 0
			for i = index + len(sd); i < len(symlink); i++ {
				if symlink[i] == filepath.Separator {
					sepCount++
				}

				// if we passed the 1st '/' after the sysdir, stop
				if sepCount > 1 {
					break
				}
			}

			minDir := symlink[:i]
			if minDir == symlink {
				return ""
			}
			return minDir
		}
	}
	return ""
}
