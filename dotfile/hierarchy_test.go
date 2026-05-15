package dotfile

import (
	"encoding/json"
	"path/filepath"
	"slices"
	"testing"

	"github.com/solerf/dtm/profile"
)

func TestBuildHierarchy(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	writeFile(t, filepath.Join(src, "a"), "x")
	prof := &profile.Info{Name: "p", Path: src}

	tests := []struct {
		name      string
		dotfiles  []Item
		wantCount int
		wantKeys  []string
	}{
		{
			name:      "empty",
			dotfiles:  nil,
			wantCount: 0,
			wantKeys:  nil,
		},
		{
			name: "flat_keys",
			dotfiles: []Item{
				{Profile: prof, Key: ".rc", SourcePath: filepath.Join(src, "a"), SymLink: filepath.Join(dir, ".rc")},
				{Profile: prof, Key: ".vimrc", SourcePath: filepath.Join(src, "a"), SymLink: filepath.Join(dir, ".vimrc")},
			},
			wantCount: 2,
			wantKeys:  []string{".rc", ".vimrc"},
		},
		{
			name: "nested_key_creates_intermediate",
			dotfiles: []Item{
				{Profile: prof, Key: ".config/nvim/init.lua", SourcePath: filepath.Join(src, "a"), SymLink: filepath.Join(dir, ".config/nvim/init.lua")},
				{Profile: prof, Key: ".config/nvim/lua.lua", SourcePath: filepath.Join(src, "a"), SymLink: filepath.Join(dir, ".config/nvim/lua.lua")},
			},
			wantCount: 1,
			wantKeys:  []string{".config"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := buildHierarchy(dir, tt.dotfiles)
			if h.Key != dir {
				t.Errorf("root.Key = %q, want %q", h.Key, dir)
			}
			if len(h.Structure) != tt.wantCount {
				t.Fatalf("len(Structure) = %d, want %d", len(h.Structure), tt.wantCount)
			}
			gotKeys := make([]string, 0, len(h.Structure))
			for _, c := range h.Structure {
				gotKeys = append(gotKeys, c.Key)
			}
			slices.Sort(gotKeys)
			slices.Sort(tt.wantKeys)
			if !slices.Equal(gotKeys, tt.wantKeys) {
				t.Errorf("top-level keys = %v, want %v", gotKeys, tt.wantKeys)
			}
		})
	}
}

func TestBuildHierarchy_NestedDepthAndLeaf(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src", "init.lua")
	writeFile(t, src, "x")
	prof := &profile.Info{Name: "p", Path: filepath.Join(dir, "src")}

	items := []Item{
		{Profile: prof, Key: ".config/nvim/init.lua", SourcePath: src, SymLink: filepath.Join(dir, ".config/nvim/init.lua")},
	}
	h := buildHierarchy(dir, items)

	// dir -> .config -> nvim -> init.lua (leaf with DotFile)
	if len(h.Structure) != 1 || h.Structure[0].Key != ".config" {
		t.Fatalf("expected .config child, got %+v", h.Structure)
	}
	cfg := h.Structure[0]
	if len(cfg.Structure) != 1 || cfg.Structure[0].Key != "nvim" {
		t.Fatalf("expected nvim child, got %+v", cfg.Structure)
	}
	nvim := cfg.Structure[0]
	if len(nvim.Structure) != 1 || nvim.Structure[0].Key != "init.lua" {
		t.Fatalf("expected init.lua leaf, got %+v", nvim.Structure)
	}
	leaf := nvim.Structure[0]
	if leaf.DotFile == nil {
		t.Fatalf("leaf DotFile = nil, want non-nil")
	}
	if leaf.DotFile.Profile != "p" {
		t.Errorf("leaf Profile = %q, want %q", leaf.DotFile.Profile, "p")
	}
	if leaf.DotFile.SourceMissing {
		t.Errorf("leaf SourceMissing = true, want false (source exists)")
	}
}

func TestHierarchy_ToJson(t *testing.T) {
	h := &Hierarchy{Key: "root"}
	got, err := h.toJson()
	if err != nil {
		t.Fatalf("toJson err = %v, want nil", err)
	}

	var round Hierarchy
	if err := json.Unmarshal([]byte(got), &round); err != nil {
		t.Fatalf("round-trip err = %v, want nil. payload = %q", err, got)
	}
	if round.Key != "root" {
		t.Errorf("round.Key = %q, want %q", round.Key, "root")
	}
}
