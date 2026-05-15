package profile

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func mkProfileTree(t *testing.T, names ...string) string {
	t.Helper()
	root := t.TempDir()
	for _, n := range names {
		if err := os.MkdirAll(filepath.Join(root, n), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", n, err)
		}
	}
	return root
}

func TestNew(t *testing.T) {
	root := mkProfileTree(t, "exists")

	tests := []struct {
		name        string
		source      string
		profileName string
		wantErr     bool
		wantInfo    Info
	}{
		{
			name:        "existing_profile",
			source:      root,
			profileName: "exists",
			wantErr:     false,
			wantInfo:    Info{Name: "exists", Path: filepath.Join(root, "exists")},
		},
		{
			name:        "missing_profile",
			source:      root,
			profileName: "missing",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.source, tt.profileName)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("New(%q,%q) error = nil, want error", tt.source, tt.profileName)
				}
				return
			}
			if err != nil {
				t.Fatalf("New(%q,%q) error = %v, want nil", tt.source, tt.profileName, err)
			}
			if got != tt.wantInfo {
				t.Errorf("New(%q,%q) = %+v, want %+v", tt.source, tt.profileName, got, tt.wantInfo)
			}
		})
	}
}

func TestTransform(t *testing.T) {
	root := mkProfileTree(t, "shared", "work", "home")

	tests := []struct {
		name      string
		names     []string
		wantNames []string
	}{
		{
			name:      "auto_prepend_shared",
			names:     []string{"work"},
			wantNames: []string{"shared", "work"},
		},
		{
			name:      "shared_already_present_no_duplicate",
			names:     []string{"shared", "work"},
			wantNames: []string{"shared", "work"},
		},
		{
			name:      "missing_profile_skipped",
			names:     []string{"work", "ghost", "home"},
			wantNames: []string{"shared", "work", "home"},
		},
		{
			name:      "all_missing_yields_only_shared",
			names:     []string{"ghost1", "ghost2"},
			wantNames: []string{"shared"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Transform(root, tt.names)
			gotNames := make([]string, 0, len(got))
			for _, p := range got {
				gotNames = append(gotNames, p.Name)
			}
			if !slices.Equal(gotNames, tt.wantNames) {
				t.Errorf("Transform names = %v, want %v", gotNames, tt.wantNames)
			}
			for _, p := range got {
				wantPath := filepath.Join(root, p.Name)
				if p.Path != wantPath {
					t.Errorf("Transform path for %q = %q, want %q", p.Name, p.Path, wantPath)
				}
			}
		})
	}
}

func TestTransform_MissingShared(t *testing.T) {
	root := mkProfileTree(t, "work")
	got := Transform(root, []string{"work"})

	gotNames := make([]string, 0, len(got))
	for _, p := range got {
		gotNames = append(gotNames, p.Name)
	}
	want := []string{"work"}
	if !slices.Equal(gotNames, want) {
		t.Errorf("Transform without shared dir = %v, want %v", gotNames, want)
	}
}
