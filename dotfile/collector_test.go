package dotfile

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/solerf/dtm/profile"
)

var noIgnore = func(s string) bool { return false }

func TestBuildKey(t *testing.T) {
	root := "testdata/source"
	tests := []struct {
		name      string
		profile   profile.Info
		key       string
		want      string
		wantError bool
	}{
		{
			name:    "directory_gets_dot_prefix",
			profile: profile.Info{Name: "a", Path: filepath.Join(root, "a")},
			key:     "d1/aa",
			want:    ".d1/aa",
		},
		{
			name:    "top_level_file_gets_dot_prefix_only_if_dir",
			profile: profile.Info{Name: "a", Path: filepath.Join(root, "a")},
			key:     "f1",
			want:    "f1",
		},
		{
			name:    "exploded_prefix_replaced_with_dot",
			profile: profile.Info{Name: "_d", Path: filepath.Join(root, "_d")},
			key:     "_xx/11",
			want:    ".xx/11",
		},
		{
			name:      "empty_key_errors",
			profile:   profile.Info{Name: "a", Path: filepath.Join(root, "a")},
			key:       "",
			wantError: true,
		},
		{
			name:      "missing_first_part_errors",
			profile:   profile.Info{Name: "a", Path: filepath.Join(root, "a")},
			key:       "ghost/x",
			wantError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildKey(&tt.profile, tt.key)
			if (err != nil) != tt.wantError {
				t.Fatalf("buildKey err = %v, wantError = %v", err, tt.wantError)
			}
			if got != tt.want {
				t.Errorf("buildKey = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCheckIsExploded(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"exploded_dir", "testdata/source/_d", true},
		{"normal_dir", "testdata/source/a", false},
		{"normal_file", "testdata/source/a/f1", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checkIsExploded(tt.path)
			if err != nil {
				t.Fatalf("checkIsExploded err = %v, want nil", err)
			}
			if got != tt.want {
				t.Errorf("checkIsExploded(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestCheckIsExploded_Missing(t *testing.T) {
	if _, err := checkIsExploded("testdata/source/ghost"); err == nil {
		t.Errorf("checkIsExploded missing path err = nil, want error")
	}
}

func TestDedup(t *testing.T) {
	tests := []struct {
		name        string
		in          map[string][]string
		wantUniques []string
		wantDups    []string
	}{
		{
			name:        "all_unique",
			in:          map[string][]string{"a": {"/p1/a"}, "b": {"/p1/b"}},
			wantUniques: []string{"/p1/a", "/p1/b"},
			wantDups:    []string{},
		},
		{
			name:        "duplicate_basename",
			in:          map[string][]string{"x": {"/p1/x", "/p2/x"}},
			wantUniques: []string{},
			wantDups:    []string{"/p1/x", "/p2/x"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotU, gotD := dedup(tt.in)
			slices.Sort(gotU)
			slices.Sort(gotD)
			slices.Sort(tt.wantUniques)
			slices.Sort(tt.wantDups)
			if !slices.Equal(gotU, tt.wantUniques) {
				t.Errorf("uniques = %v, want %v", gotU, tt.wantUniques)
			}
			if !slices.Equal(gotD, tt.wantDups) {
				t.Errorf("duplicates = %v, want %v", gotD, tt.wantDups)
			}
		})
	}
}

func TestListPathEntries(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "keep"), "x")
	writeFile(t, filepath.Join(dir, "ignore_me"), "x")

	ignoreFn := func(s string) bool { return s == "ignore_me" }
	got, err := listPathEntries(ignoreFn, dir)
	if err != nil {
		t.Fatalf("listPathEntries err = %v, want nil", err)
	}
	want := []string{"keep"}
	if !slices.Equal(got, want) {
		t.Errorf("listPathEntries = %v, want %v", got, want)
	}
}

func TestListPathEntries_Missing(t *testing.T) {
	if _, err := listPathEntries(noIgnore, filepath.Join(t.TempDir(), "ghost")); err == nil {
		t.Errorf("listPathEntries err = nil, want error")
	}
}

func TestCollectFromPaths(t *testing.T) {
	root := "testdata/source"
	gotU, gotD, err := collectFromPaths(noIgnore, filepath.Join(root, "a"), filepath.Join(root, "c"))
	if err != nil {
		t.Fatalf("collectFromPaths err = %v, want nil", err)
	}

	// a has d1, f1, just at top-level; c has d1 -- d1 collides as duplicate
	wantU := []string{
		filepath.Join(root, "a", "f1"),
		filepath.Join(root, "a", "just"),
	}
	wantD := []string{
		filepath.Join(root, "a", "d1"),
		filepath.Join(root, "c", "d1"),
	}
	slices.Sort(gotU)
	slices.Sort(gotD)
	slices.Sort(wantU)
	slices.Sort(wantD)
	if !slices.Equal(gotU, wantU) {
		t.Errorf("uniques = %v, want %v", gotU, wantU)
	}
	if !slices.Equal(gotD, wantD) {
		t.Errorf("duplicates = %v, want %v", gotD, wantD)
	}
}

func TestCollectFromPaths_DuplicateBasename(t *testing.T) {
	root := "testdata/source"
	_, gotD, err := collectFromPaths(noIgnore, filepath.Join(root, "a"), filepath.Join(root, "b"))
	if err != nil {
		t.Fatalf("collectFromPaths err = %v, want nil", err)
	}
	// no duplicates: a has d1/just/f1; b has d2/d3/f2 — distinct basenames
	if len(gotD) != 0 {
		t.Errorf("duplicates = %v, want empty", gotD)
	}
}

func TestCollectFromPaths_Exploded(t *testing.T) {
	root := "testdata/source"
	gotU, gotD, err := collectFromPaths(noIgnore, root)
	if err != nil {
		t.Fatalf("collectFromPaths err = %v, want nil", err)
	}
	// _d is exploded; its files should appear individually as full paths
	wantContains := []string{
		filepath.Join(root, "_d", "xx", "11"),
		filepath.Join(root, "_d", "xx", "12"),
	}
	for _, w := range wantContains {
		if !slices.Contains(gotU, w) {
			t.Errorf("uniques = %v, want to contain %q", gotU, w)
		}
	}
	if len(gotD) != 0 {
		t.Errorf("duplicates = %v, want empty", gotD)
	}
}

func TestCollectItemPaths(t *testing.T) {
	target := t.TempDir()
	root, err := filepath.Abs("testdata/source")
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	profA := profile.Info{Name: "a", Path: filepath.Join(root, "a")}

	ignoreFn := func(s string) bool { return s == "f_IGNORE" || s == "just" }
	items, err := collectItemPaths(target, ignoreFn, profA)
	if err != nil {
		t.Fatalf("collectItemPaths err = %v, want nil", err)
	}

	gotKeys := make([]string, 0, len(items))
	for _, it := range items {
		gotKeys = append(gotKeys, it.Key)
	}
	slices.Sort(gotKeys)
	// non-exploded dirs are linked as a whole (key=".d1"), not file-by-file
	want := []string{".d1", "f1"}
	if !slices.Equal(gotKeys, want) {
		t.Errorf("keys = %v, want %v", gotKeys, want)
	}
}
