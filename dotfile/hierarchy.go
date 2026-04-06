package dotfile

import (
	"path"
	"slices"
	"strings"

	"github.com/solerf/dtm/common"
)

type node struct {
	Source        string `json:"source_path"`
	Target        string `json:"target_path"`
	SourceMissing bool   `json:"source_missing"`
}

type Hierarchy struct {
	Key       string       `json:"key"`
	Structure []*Hierarchy `json:"structure,omitempty"`
	// should only exist in the end node
	DotFile *node `json:"dotfile,omitempty"`
}

func buildHierarchy(targetDir string, dotfiles []entry) *Hierarchy {
	var build func(parent *Hierarchy, parentPath string, p string, dotfile *entry)
	build = func(parent *Hierarchy, parentPath string, p string, dotfile *entry) {
		paths := strings.Split(p, common.PathSeparator)

		var current *Hierarchy

		if len(paths) == 1 {
			// when last part of path
			exists := common.MustPathExists(dotfile.SourcePath)
			current = &Hierarchy{
				Key: path.Base(dotfile.SourcePath),
				DotFile: &node{
					Source:        dotfile.SourcePath,
					Target:        dotfile.TargetPath,
					SourceMissing: !exists,
				}}
			parent.Structure = append(parent.Structure, current)
			return
		}

		base := paths[0]
		rest := path.Join(paths[1:]...)

		foundBaseIdx := slices.IndexFunc(parent.Structure, func(i *Hierarchy) bool {
			return i.Key == base
		})

		if foundBaseIdx != -1 {
			current = parent.Structure[foundBaseIdx]
			build(current, base, rest, dotfile)
			return
		}

		current = &Hierarchy{Key: base, Structure: make([]*Hierarchy, 0, len(rest))}
		build(current, base, rest, dotfile)
		parent.Structure = append(parent.Structure, current)
	}

	root := &Hierarchy{Key: targetDir}
	for _, df := range dotfiles {
		build(root, targetDir, df.Key, &df)
	}
	return root
}
