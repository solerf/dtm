package runner

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

const (
	pathSeparator = string(filepath.Separator)
)

var (
	explodeDirs = []string{".local"}
)

type dotFile struct {
	Key        string `json:"key"`
	SourcePath string `json:"source_path"`
}

func (d *dotFile) isExploded() bool {
	return slices.ContainsFunc(explodeDirs, func(e string) bool {
		return strings.HasSuffix(d.SourcePath, e)
	})
}

func (d *dotFile) install(at string) error {
	target := path.Join(at, d.Key)
	targetDir := path.Dir(target)

	_, errLink := os.Stat(targetDir)
	if errLink != nil && errors.Is(errLink, os.ErrNotExist) {
		if errors.Is(errLink, os.ErrNotExist) {
			if errMkdir := os.MkdirAll(targetDir, fileMode); errMkdir != nil {
				return fmt.Errorf("mkdir all [%v]: %w", targetDir, errMkdir)
			}
		} else {
			return fmt.Errorf("stating link [%v]: %w", target, errLink)
		}
	}

	if _, errLstat := os.Lstat(target); errLstat == nil {
		// check if points to same target that is trying to map
		foundLinkTarget, errRead := os.Readlink(target)
		if errRead != nil {
			return fmt.Errorf("reading link [%v]: %w", target, errRead)
		}

		if foundLinkTarget != d.SourcePath {
			// if different it is removed to then be recreated
			if errRem := os.Remove(target); errRem != nil {
				return fmt.Errorf("remove existent [%v]: %w", target, errRem)
			}
		} else {
			return nil
		}
	}

	if errSymlink := os.Symlink(d.SourcePath, target); errSymlink != nil {
		return fmt.Errorf("linking [%v]: %w", target, errSymlink)
	}

	return nil
}

func (d *dotFile) uninstall(from string) error {
	toRemove := path.Join(from, d.Key)

	ok, foundSysDir := hasPrefixAny(d.Key, sysDirs)
	if ok {
		var keyWithoutSysDir string
		keyWithoutSysDir = strings.ReplaceAll(d.Key, foundSysDir+pathSeparator, "")

		if len(keyWithoutSysDir) > 0 {
			base := strings.Split(keyWithoutSysDir, pathSeparator)[0]
			dir := filepath.Join(from, foundSysDir, base)

			if stat, err := os.Stat(dir); err == nil {
				if stat.IsDir() {
					toRemove = path.Join(from, dir)
				}
			}
		}
	}

	if _, err := os.Stat(toRemove); err == nil {
		if err = os.RemoveAll(toRemove); err != nil {
			return fmt.Errorf("removing [%v]: %w", toRemove, err)
		}
	}
	return nil
}

func newDotFile(sourceFullpath string) dotFile {
	key := relativeToProfile(sourceFullpath)
	return dotFile{Key: key, SourcePath: sourceFullpath}
}

func relativeToProfile(p string) string {
	split := strings.Split(p, string(filepath.Separator))
	profileIndex := slices.IndexFunc(split, func(s string) bool {
		return strings.HasPrefix(s, profilePrefix)
	})
	return path.Join(split[profileIndex+1:]...)
}

func extractPaths(dotfiles []dotFile) []string {
	paths := make([]string, 0, len(dotfiles))
	for i := 0; i < len(dotfiles); i++ {
		paths = append(paths, dotfiles[i].SourcePath)
	}
	return paths
}
