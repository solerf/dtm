package dotfile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/solerf/dtm/profile"
)

const dirMode = os.FileMode(0755)

type Item struct {
	Profile *profile.Info `json:"profile"`
	//the key serves as to map the real installation path at provided 'targetDir'
	Key string `json:"key"`
	// to where link points
	SourcePath string `json:"source_path"`
	// effective symlink
	SymLink string `json:"symlink"`
}

func (e *Item) install() (Status, error) {
	lStat, errStat := os.Lstat(e.SymLink)

	if errStat != nil && !errors.Is(errStat, os.ErrNotExist) {
		return Failed, fmt.Errorf("reading link [%v]: %w", e.SymLink, errStat)
	}

	if errStat == nil {
		if !isSymlinkMode(lStat.Mode()) {
			return Failed, fmt.Errorf("path already exists [%v] and is not a symlink", e.SymLink)
		}

		foundLinkTarget, err := os.Readlink(e.SymLink)
		if err != nil {
			return Failed, fmt.Errorf("reading link [%v]: %w", e.SymLink, err)
		}

		if foundLinkTarget == e.SourcePath {
			return Skipped, nil
		}
		return Failed, fmt.Errorf("link already exists [%v] and points to another source [%s]", e.SymLink, foundLinkTarget)
	}

	symlinkDir := filepath.Dir(e.SymLink)
	if err := createIfMissing(symlinkDir); err != nil {
		return Failed, err
	}

	//install
	if err := os.Symlink(e.SourcePath, e.SymLink); err != nil {
		return Failed, fmt.Errorf("linking [%v]: %w", e.SymLink, err)
	}
	return Ok, nil
}

func (e *Item) uninstall() (Status, error) {
	// check if link still exists before removing
	exist, err := e.linkExists()

	if err != nil {
		return Failed, fmt.Errorf("failed checking path [%v] to remove: %w", e.SymLink, err)
	}

	if !exist {
		return Skipped, nil
	}

	if err := os.RemoveAll(e.SymLink); err != nil {
		return Failed, fmt.Errorf("removing [%v]: %w", e.SymLink, err)
	}
	return Ok, nil
}

func (e *Item) linkExists() (bool, error) {
	lStat, err := os.Lstat(e.SymLink)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return isSymlinkMode(lStat.Mode()), nil
}

func (e *Item) sourceExists() (bool, error) {
	if _, err := os.Stat(e.SourcePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func createIfMissing(dir string) error {
	if err := os.MkdirAll(dir, dirMode); err != nil {
		return fmt.Errorf("mkdir all [%v]: %w", dir, err)
	}
	return nil
}

func newEntry(targetDir string, prof *profile.Info, key, itemPath string) (Item, error) {
	item := Item{
		Profile:    prof,
		Key:        key,
		SourcePath: itemPath,
		SymLink:    filepath.Join(targetDir, key),
	}

	exists, err := item.sourceExists()
	if err != nil {
		return Item{}, fmt.Errorf("new entry stat [%v]: %w", itemPath, err)
	}

	if !exists {
		return Item{}, fmt.Errorf("new entry missing [%v]", itemPath)
	}
	return item, nil
}

func isSymlinkMode(mode os.FileMode) bool {
	return mode&os.ModeSymlink != 0
}
