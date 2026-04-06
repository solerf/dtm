package dotfile

import (
	"bufio"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
)

var (
	ignoreGlobs = map[string]struct{}{
		".gitignore": {},
		"-*":         {},
		"*_IGNORE":   {},
	}
)

func IgnoreFilter(source string) func(string) bool {
	globList, _ := ignoreList(source)
	return func(s string) bool {
		for i := 0; i < len(globList); i++ {
			if globList[i].Match(s) {
				return true
			}
		}
		return false
	}
}

func ignoreList(source string) ([]glob.Glob, error) {
	ignores := make(map[string]struct{}, len(ignoreGlobs))
	maps.Insert(ignores, maps.All(ignoreGlobs))

	err := filepath.WalkDir(source, walkDir(ignores))
	if err != nil {
		return nil, fmt.Errorf("ignore: %w", err)
	}

	globs := make([]glob.Glob, 0, len(ignores))
	for ig := range maps.Keys(ignores) {
		globs = append(globs, glob.MustCompile(ig))
	}
	return globs, nil
}

func walkDir(ignores map[string]struct{}) func(string, fs.DirEntry, error) error {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walk [%v]: %w", path, err)
		}

		if !d.IsDir() && strings.HasSuffix(path, ".gitignore") {
			file, err := os.OpenFile(path, os.O_RDONLY, 0)
			if err != nil {
				return fmt.Errorf("open [%v]: %w", path, err)
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				t := strings.TrimSpace(scanner.Text())
				if len(t) > 0 && t[0] != '#' {
					ignores[t] = struct{}{}
				}
			}
		}
		return nil
	}
}
