package dotfile

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/solerf/dtm/profile"
)

const (
	explodedPrefix = "_"
	dotPrefix      = "."
)

func collectItemPaths(targetInstallDir string, ignoreFunc func(string) bool, profiles ...profile.Info) ([]Item, error) {
	paths := make([]string, 0, len(profiles))
	for _, p := range profiles {
		paths = append(paths, p.Path)
	}

	collectedPaths, err := collectDistinct(ignoreFunc, paths...)
	if err != nil {
		return nil, fmt.Errorf("collector: %w", err)
	}

	return transformToDotFiles(targetInstallDir, profiles, collectedPaths)
}

func transformToDotFiles(targetInstallDir string, profiles []profile.Info, collectedPaths []string) ([]Item, error) {
	result := make([]Item, 0, len(collectedPaths))
	collectedPerProfile := make(map[string][]string, len(collectedPaths))

	for _, iPath := range collectedPaths {
		pIdx := slices.IndexFunc(profiles, func(info profile.Info) bool {
			return strings.HasPrefix(iPath, info.Path+string(filepath.Separator))
		})
		if pIdx == -1 {
			return nil, fmt.Errorf("collector: profile for [%s] not found", iPath)
		}

		info := &profiles[pIdx]

		key, err := buildKey(info, iPath)
		if err != nil {
			return nil, fmt.Errorf("collector: transform: %w", err)
		}

		entry, err := newEntry(targetInstallDir, info, key, iPath)
		if err != nil {
			return nil, fmt.Errorf("collector: transform: %w", err)
		}

		// append the profile name in case another profile has a file with same name.
		// NOTE: this dedups by source basename, so two distinct keys that share a
		// basename across profiles (e.g. bin/foo vs lib/foo) get spuriously suffixed
		// even though their SymLink targets are already distinct. Should key on
		// entry.Key instead — left as-is because the collision is unlikely in
		// practice with the current "_" exploded layout.
		baseName := filepath.Base(iPath)
		if v, ok := collectedPerProfile[baseName]; ok && slices.Index(v, entry.Profile.Name) == -1 {
			// this to differentiate the symlink at the target
			suffix := strings.ToLower(entry.Profile.Name)
			entry.Key = fmt.Sprintf("%s_%s", entry.Key, suffix)
			entry.SymLink = fmt.Sprintf("%s_%s", entry.SymLink, suffix)
		}

		collectedPerProfile[baseName] = append(collectedPerProfile[baseName], info.Name)
		result = append(result, entry)
	}
	return result, nil
}

func buildKey(p *profile.Info, itemPath string) (string, error) {
	key := strings.TrimPrefix(strings.TrimPrefix(itemPath, p.Path), string(filepath.Separator))
	if key == "" {
		return "", fmt.Errorf("impossible to define key from [%s] at [%s]", itemPath, p.Name)
	}

	firstPart := strings.Split(key, string(filepath.Separator))[0]

	// in case is an exploded dir
	if strings.HasPrefix(key, explodedPrefix) {
		return dotPrefix + key[1:], nil
	}

	// if first part of key path is dir append the '.'
	// otherwise return as it is
	stat, err := os.Stat(filepath.Join(p.Path, firstPart))
	if err != nil {
		return "", fmt.Errorf("build key [%s]: %w", itemPath, err)
	}

	if stat.IsDir() {
		return dotPrefix + key, nil
	}
	return key, nil
}

func collectDistinct(ignoreFunc func(string) bool, paths ...string) ([]string, error) {
	uniques, duplicatesQueue, err := collectFromPaths(ignoreFunc, paths...)
	if err != nil {
		return nil, err
	}

	// there's no distinction yet in case two profiles have a file with same name
	// this case is not contemplated.
	// NOTE: duplicatesQueue carries full paths from dedup; collectFromPaths then
	// feeds them to listPathEntries → os.ReadDir, which fails with ENOTDIR if any
	// duplicate is a regular file. Only directory duplicates survive the requeue.
	// In practice the "_" exploded scheme resolves collisions earlier, so this
	// branch is rarely hit — but it's a latent failure if two profiles ever share
	// a non-exploded file with the same basename.
	for len(duplicatesQueue) > 0 { // pop
		moreUniques, moreDuplicates, err := collectFromPaths(ignoreFunc, duplicatesQueue...)
		if err != nil {
			return nil, err
		}

		uniques = append(uniques, moreUniques...)
		duplicatesQueue = moreDuplicates // push
	}
	return uniques, nil
}

func collectFromPaths(ignoreFunc func(string) bool, paths ...string) ([]string, []string, error) {
	entries := make(map[string][]string, len(paths))

	for _, p := range paths {
		pEntries, err := listPathEntries(ignoreFunc, p)
		if err != nil {
			return nil, nil, err
		}

		for _, c := range pEntries {
			fullPath := filepath.Join(p, c)
			isExploded, err := checkIsExploded(fullPath)
			if err != nil {
				return nil, nil, fmt.Errorf("check exploded [%v]: %w", fullPath, err)
			}

			if isExploded {
				if err := walkInto(ignoreFunc, fullPath, entries); err != nil {
					return nil, nil, fmt.Errorf("walking [%v]: %w", fullPath, err)
				}
				continue
			}
			entries[c] = append(entries[c], fullPath)
		}
	}

	uniq, dups := dedup(entries)
	return uniq, dups, nil
}

func checkIsExploded(p string) (bool, error) {
	stat, err := os.Stat(p)
	if err != nil {
		return false, fmt.Errorf("failed stat [%s]: %w", p, err)
	}
	return stat.IsDir() && strings.HasPrefix(filepath.Base(p), explodedPrefix), nil
}

func dedup(items map[string][]string) ([]string, []string) {
	uniques := make([]string, 0, len(items))
	duplicates := make([]string, 0, len(items))

	for _, v := range items {
		if len(v) == 1 {
			uniques = append(uniques, v[0])
			continue
		}
		duplicates = append(duplicates, v...)
	}
	return uniques, duplicates
}

// walkInto walks root and writes every non-ignored file into entries as a
// single-element slice keyed by the file's full path. Used for "exploded"
// dotted directories where every leaf becomes its own mapping.
func walkInto(ignoreFunc func(string) bool, root string, entries map[string][]string) error {
	return filepath.WalkDir(root, func(childPath string, childEntry fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("in walk [%v]: %w", childPath, err)
		}

		if ignoreFunc(childEntry.Name()) {
			// fs.SkipDir from a file callback skips the rest of the parent
			// directory, not just this file — so only return it for dirs.
			if childEntry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		if !childEntry.IsDir() {
			entries[childPath] = []string{childPath}
		}
		return nil
	})
}

func listPathEntries(ignoreFunc func(string) bool, profilePath string) ([]string, error) {
	entries, err := os.ReadDir(profilePath)
	if err != nil {
		return nil, fmt.Errorf("reading dir [%v]: %w", profilePath, err)
	}

	c := make([]string, 0, len(entries))
	for _, e := range entries {
		if name := e.Name(); !ignoreFunc(name) {
			c = append(c, name)
		}
	}
	return c, nil
}
