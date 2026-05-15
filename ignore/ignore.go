package ignore

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
)

func FilterFunc(source string) (func(string) bool, error) {
	globs, err := ignoreList(source)
	if err != nil {
		return nil, err
	}

	return func(s string) bool {
		for _, g := range globs {
			if g.Match(s) {
				return true
			}
		}
		return false
	}, nil
}

// patterns holds the ordered ignore-pattern list plus a dedup set populated
// as nested .gitignore files are merged in.
type patterns struct {
	list    []string
	control map[string]struct{}
}

func (p *patterns) add(s string) {
	if _, ok := p.control[s]; ok {
		return
	}
	p.control[s] = struct{}{}
	p.list = append(p.list, s)
}

func ignoreList(source string) ([]glob.Glob, error) {
	ps := &patterns{
		list:    make([]string, 0, 8),
		control: make(map[string]struct{}, 8),
	}
	for _, p := range []string{".gitignore", "-*", "*_IGNORE"} {
		ps.add(p)
	}

	if strings.TrimSpace(source) != "" {
		if err := filepath.WalkDir(source, walkDir(ps)); err != nil {
			return nil, fmt.Errorf("ignore: %w", err)
		}
	}

	globs := make([]glob.Glob, 0, len(ps.list))
	for _, p := range ps.list {
		compile, err := glob.Compile(p)
		if err != nil {
			log.Printf("[WARN] failed glob [%s]: %v\n", p, err)
			continue
		}
		globs = append(globs, compile)
	}
	return globs, nil
}

func walkDir(ps *patterns) func(string, fs.DirEntry, error) error {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walk [%v]: %w", path, err)
		}

		if !d.IsDir() && strings.HasSuffix(path, ".gitignore") {
			gIgnore, err := readGitIgnore(path)
			if err != nil {
				return fmt.Errorf("open [%v]: %w", path, err)
			}

			for _, i := range gIgnore {
				ps.add(i)
			}
		}
		return nil
	}
}

func readGitIgnore(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("reading [%v]: %w", path, err)
	}
	defer file.Close()

	ignores := make([]string, 0, 3)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if t := strings.TrimSpace(scanner.Text()); strings.TrimSpace(t) != "" && t[0] != '#' {
			ignores = append(ignores, t)
		}
	}

	if scanner.Err() != nil {
		return nil, fmt.Errorf("scanning [%v]: %w", path, scanner.Err())
	}
	return ignores, nil
}
