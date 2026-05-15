package dotfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindSysDir(t *testing.T) {
	tests := []struct {
		name    string
		symlink string
		want    string
	}{
		{
			name:    "config_with_subpath",
			symlink: "/home/u/.config/nvim/init.lua",
			want:    "/home/u/.config/nvim",
		},
		{
			name:    "local_bin_with_subpath",
			symlink: "/home/u/.local/bin/script/sub",
			want:    "/home/u/.local/bin/script",
		},
		{
			name:    "no_match_returns_empty",
			symlink: "/home/u/.bashrc",
			want:    "",
		},
		{
			name:    "config_direct_child_returns_empty",
			symlink: "/home/u/.config/file",
			want:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findSysDir(tt.symlink)
			if got != tt.want {
				t.Errorf("findSysDir(%q) = %q, want %q", tt.symlink, got, tt.want)
			}
		})
	}
}

func TestPruneEmptyDirs(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) (root string, expectExists []string, expectGone []string)
	}{
		{
			name: "removes_empty_tree",
			setup: func(t *testing.T) (string, []string, []string) {
				root := filepath.Join(t.TempDir(), "a")
				if err := os.MkdirAll(filepath.Join(root, "b", "c"), 0o755); err != nil {
					t.Fatalf("mkdir: %v", err)
				}
				return root, nil, []string{root, filepath.Join(root, "b"), filepath.Join(root, "b", "c")}
			},
		},
		{
			name: "keeps_dir_with_untracked_file",
			setup: func(t *testing.T) (string, []string, []string) {
				root := filepath.Join(t.TempDir(), "a")
				if err := os.MkdirAll(filepath.Join(root, "b"), 0o755); err != nil {
					t.Fatalf("mkdir: %v", err)
				}
				keep := filepath.Join(root, "b", "user_file")
				writeFile(t, keep, "x")
				return root, []string{root, filepath.Join(root, "b"), keep}, nil
			},
		},
		{
			name: "removes_empty_subdirs_keeps_parent_with_files",
			setup: func(t *testing.T) (string, []string, []string) {
				root := filepath.Join(t.TempDir(), "a")
				if err := os.MkdirAll(filepath.Join(root, "empty"), 0o755); err != nil {
					t.Fatalf("mkdir: %v", err)
				}
				keep := filepath.Join(root, "f")
				writeFile(t, keep, "x")
				return root, []string{root, keep}, []string{filepath.Join(root, "empty")}
			},
		},
		{
			name: "missing_root_no_error",
			setup: func(t *testing.T) (string, []string, []string) {
				return filepath.Join(t.TempDir(), "ghost"), nil, nil
			},
		},
		{
			name: "non_dir_no_op",
			setup: func(t *testing.T) (string, []string, []string) {
				p := filepath.Join(t.TempDir(), "f")
				writeFile(t, p, "x")
				return p, []string{p}, nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, expectExists, expectGone := tt.setup(t)
			if err := pruneEmptyDirs(root); err != nil {
				t.Fatalf("pruneEmptyDirs err = %v, want nil", err)
			}
			for _, p := range expectExists {
				if _, err := os.Stat(p); err != nil {
					t.Errorf("expected %q to exist: stat err = %v", p, err)
				}
			}
			for _, p := range expectGone {
				if _, err := os.Stat(p); !os.IsNotExist(err) {
					t.Errorf("expected %q to be removed: stat err = %v", p, err)
				}
			}
		})
	}
}

func TestUninstall_Roundtrip(t *testing.T) {
	source := setupSourceWithProfile(t, "work", map[string]string{
		"f1":    "x",
		"d1/aa": "y",
	})
	target := t.TempDir()

	statuses, err := Install(source, target, "work")
	if err != nil {
		t.Fatalf("Install err = %v, want nil", err)
	}

	uStatuses, err := Uninstall(target)
	if err != nil {
		t.Fatalf("Uninstall err = %v, want nil", err)
	}
	if got, want := len(uStatuses), len(statuses); got != want {
		t.Errorf("uninstall statuses len = %d, want %d", got, want)
	}
	for _, s := range uStatuses {
		if s.Status != Ok {
			t.Errorf("uninstall status for %q = %v (err=%v), want Ok", s.Dotfile.Key, s.Status, s.Error)
		}
		if _, statErr := os.Lstat(s.Dotfile.SymLink); !os.IsNotExist(statErr) {
			t.Errorf("symlink %q still exists after uninstall: %v", s.Dotfile.SymLink, statErr)
		}
	}

	if _, err := os.Stat(filepath.Join(target, mappingsFileName)); !os.IsNotExist(err) {
		t.Errorf("mapping file still exists after Uninstall: stat err = %v", err)
	}
}

func TestUninstall_MissingMapping(t *testing.T) {
	if _, err := Uninstall(t.TempDir()); err == nil {
		t.Errorf("Uninstall without mapping err = nil, want error")
	}
}
