package dotfile

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/solerf/dtm/profile"
)

func TestMapping_WriteAndRead(t *testing.T) {
	target := t.TempDir()
	prof := &profile.Info{Name: "p", Path: "/src/p"}
	items := []Item{
		{Profile: prof, Key: ".rc", SourcePath: "/src/p/rc", SymLink: filepath.Join(target, ".rc")},
		{Profile: prof, Key: ".vimrc", SourcePath: "/src/p/vimrc", SymLink: filepath.Join(target, ".vimrc")},
	}

	m := newMapping(target, items)
	if err := m.write(); err != nil {
		t.Fatalf("write err = %v, want nil", err)
	}

	if _, err := os.Stat(filepath.Join(target, mappingsFileName)); err != nil {
		t.Fatalf("mappings file missing: %v", err)
	}

	got, err := readMapping(target)
	if err != nil {
		t.Fatalf("readMapping err = %v, want nil", err)
	}
	if got.InstallDir != target {
		t.Errorf("InstallDir = %q, want %q", got.InstallDir, target)
	}
	if len(got.DotFiles) != len(items) {
		t.Fatalf("DotFiles len = %d, want %d", len(got.DotFiles), len(items))
	}
	for i, it := range items {
		if got.DotFiles[i].Key != it.Key {
			t.Errorf("DotFiles[%d].Key = %q, want %q", i, got.DotFiles[i].Key, it.Key)
		}
		if got.DotFiles[i].SymLink != it.SymLink {
			t.Errorf("DotFiles[%d].SymLink = %q, want %q", i, got.DotFiles[i].SymLink, it.SymLink)
		}
	}
}

func TestMapping_Delete(t *testing.T) {
	target := t.TempDir()
	m := newMapping(target, nil)
	if err := m.write(); err != nil {
		t.Fatalf("write: %v", err)
	}

	if err := m.Delete(); err != nil {
		t.Fatalf("Delete err = %v, want nil", err)
	}
	if _, err := os.Stat(filepath.Join(target, mappingsFileName)); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("mapping file still exists after Delete: stat err = %v", err)
	}
}

func TestReadMapping_Missing(t *testing.T) {
	if _, err := readMapping(t.TempDir()); err == nil {
		t.Errorf("readMapping err = nil, want error for missing file")
	}
}

func TestReadMapping_InvalidJSON(t *testing.T) {
	target := t.TempDir()
	if err := os.WriteFile(filepath.Join(target, mappingsFileName), []byte("not json"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := readMapping(target); err == nil {
		t.Errorf("readMapping err = nil, want error for invalid JSON")
	}
}
