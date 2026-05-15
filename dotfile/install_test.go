package dotfile

import (
	"os"
	"path/filepath"
	"testing"
)

func setupSourceWithProfile(t *testing.T, profileName string, files map[string]string) string {
	t.Helper()
	source := t.TempDir()
	for rel, content := range files {
		writeFile(t, filepath.Join(source, profileName, rel), content)
	}
	// Transform expects a "shared" profile too
	if err := os.MkdirAll(filepath.Join(source, "shared"), 0o755); err != nil {
		t.Fatalf("shared dir: %v", err)
	}
	return source
}

func TestInstall_SourceEqualsTarget(t *testing.T) {
	dir := t.TempDir()
	if _, err := Install(dir, dir, "work"); err == nil {
		t.Errorf("Install(same,same) err = nil, want error")
	}
}

func TestInstall_MissingSource(t *testing.T) {
	target := t.TempDir()
	src := filepath.Join(t.TempDir(), "ghost")
	if _, err := Install(src, target, "work"); err == nil {
		t.Errorf("Install missing source err = nil, want error")
	}
}

func TestInstall_HappyPath(t *testing.T) {
	source := setupSourceWithProfile(t, "work", map[string]string{
		"f1":    "x",
		"d1/aa": "y",
	})
	target := t.TempDir()

	statuses, err := Install(source, target, "work")
	if err != nil {
		t.Fatalf("Install err = %v, want nil", err)
	}
	if len(statuses) == 0 {
		t.Fatalf("Install statuses empty, want >0")
	}

	for _, s := range statuses {
		if s.Status != Ok {
			t.Errorf("status for %q = %v, want Ok (err=%v)", s.Dotfile.Key, s.Status, s.Error)
		}
		if _, err := os.Lstat(s.Dotfile.SymLink); err != nil {
			t.Errorf("expected symlink %q to exist: %v", s.Dotfile.SymLink, err)
		}
	}

	// mapping persisted
	if _, err := os.Stat(filepath.Join(target, mappingsFileName)); err != nil {
		t.Errorf("mapping file missing after install: %v", err)
	}
	mapping, err := readMapping(target)
	if err != nil {
		t.Fatalf("readMapping err = %v, want nil", err)
	}
	if got, want := len(mapping.DotFiles), len(statuses); got != want {
		t.Errorf("mapping DotFiles len = %d, want %d (only successful installs)", got, want)
	}
}

func TestInstall_IdempotentSkipsExisting(t *testing.T) {
	source := setupSourceWithProfile(t, "work", map[string]string{"f1": "x"})
	target := t.TempDir()

	if _, err := Install(source, target, "work"); err != nil {
		t.Fatalf("first Install err = %v, want nil", err)
	}
	statuses, err := Install(source, target, "work")
	if err != nil {
		t.Fatalf("second Install err = %v, want nil", err)
	}

	for _, s := range statuses {
		if s.Status != Skipped {
			t.Errorf("re-install status = %v, want Skipped", s.Status)
		}
	}
}

func TestRemoveStaleEntries_RemovesDanglingLink(t *testing.T) {
	target := t.TempDir()
	gone := filepath.Join(t.TempDir(), "gone")
	link := filepath.Join(target, ".rc")
	if err := os.Symlink(gone, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	m := newMapping(target, []Item{
		{Key: ".rc", SourcePath: gone, SymLink: link},
	})
	if err := m.write(); err != nil {
		t.Fatalf("write mapping: %v", err)
	}

	removeStaleEntries(target)

	if _, err := os.Lstat(link); !os.IsNotExist(err) {
		t.Errorf("stale link still exists: stat err = %v", err)
	}
}

func TestRemoveStaleEntries_KeepsLiveLink(t *testing.T) {
	target := t.TempDir()
	srcDir := t.TempDir()
	src := filepath.Join(srcDir, "f")
	writeFile(t, src, "x")

	link := filepath.Join(target, ".rc")
	if err := os.Symlink(src, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	m := newMapping(target, []Item{{Key: ".rc", SourcePath: src, SymLink: link}})
	if err := m.write(); err != nil {
		t.Fatalf("write mapping: %v", err)
	}

	removeStaleEntries(target)

	if _, err := os.Lstat(link); err != nil {
		t.Errorf("live link removed: stat err = %v, want exists", err)
	}
}

func TestRemoveStaleEntries_NoMappingNoOp(t *testing.T) {
	// just ensure it doesn't panic / error out loud
	removeStaleEntries(t.TempDir())
}
