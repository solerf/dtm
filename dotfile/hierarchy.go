package dotfile

import (
	"encoding/json"
	"path/filepath"
	"strings"
)

type node struct {
	Profile       string `json:"profile"`
	Source        string `json:"source_path"`
	Target        string `json:"target_path"`
	SourceMissing bool   `json:"source_missing"`
}

type Hierarchy struct {
	Key       string       `json:"key"`
	Structure []*Hierarchy `json:"structure,omitempty"`

	// should only exist in the end node
	DotFile *node `json:"dotfile,omitempty"`

	children map[string]*Hierarchy
}

func (h *Hierarchy) toJson() (string, error) {
	marshal, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return "", err
	}
	return string(marshal), nil
}

func buildHierarchy(targetDir string, dotfiles []Item) *Hierarchy {
	root := &Hierarchy{Key: targetDir}
	for i := range dotfiles {
		insert(root, dotfiles[i].Key, &dotfiles[i])
	}
	return root
}

func insert(parent *Hierarchy, key string, df *Item) {
	sep := byte(filepath.Separator)
	for {
		end := strings.IndexByte(key, sep)
		if end == -1 {
			exists, _ := df.sourceExists()
			parent.Structure = append(parent.Structure, &Hierarchy{
				Key: key,
				DotFile: &node{
					Profile:       df.Profile.Name,
					Source:        df.SourcePath,
					Target:        df.SymLink,
					SourceMissing: !exists,
				},
			})
			return
		}

		base := key[:end]
		child, ok := parent.children[base]
		if !ok {
			child = &Hierarchy{Key: base}
			if parent.children == nil {
				parent.children = make(map[string]*Hierarchy)
			}
			parent.children[base] = child
			parent.Structure = append(parent.Structure, child)
		}
		parent = child
		key = key[end+1:]
	}
}
