package ignore

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestFilterFunc_Defaults(t *testing.T) {
	root := t.TempDir()

	filter, err := FilterFunc(root)
	if err != nil {
		t.Fatalf("FilterFunc error = %v, want nil", err)
	}

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"gitignore_match", ".gitignore", true},
		{"underscore_prefix_match", "-foo", true},
		{"ignore_suffix_match", "thing_IGNORE", true},
		{"normal_file", "file.txt", false},
		{"dotfile_unrelated", ".bashrc", false},
		{"empty_string", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filter(tt.input)
			if got != tt.want {
				t.Errorf("filter(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestFilterFunc_LoadsGitIgnore(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".gitignore"), "*.log\n# a comment\n\nbuild\n")
	writeFile(t, filepath.Join(root, "nested", ".gitignore"), "*.tmp\n")

	filter, err := FilterFunc(root)
	if err != nil {
		t.Fatalf("FilterFunc error = %v, want nil", err)
	}

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"top_level_glob", "debug.log", true},
		{"top_level_literal", "build", true},
		{"nested_glob", "thing.tmp", true},
		{"comment_not_added", "# a comment", false},
		{"blank_line_not_added", "", false},
		{"unmatched", "main.go", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filter(tt.input)
			if got != tt.want {
				t.Errorf("filter(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestFilterFunc_EmptySource(t *testing.T) {
	filter, err := FilterFunc("")
	if err != nil {
		t.Fatalf("FilterFunc(\"\") error = %v, want nil", err)
	}

	if got := filter("file.go"); got != false {
		t.Errorf("filter(\"file.go\") = %v, want false", got)
	}
	if got := filter(".gitignore"); got != true {
		t.Errorf("filter(\".gitignore\") = %v, want true (default pattern)", got)
	}
}

func TestFilterFunc_MissingSource(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	if _, err := FilterFunc(missing); err == nil {
		t.Errorf("FilterFunc(%q) error = nil, want error", missing)
	}
}

func TestPatterns_AddDeduplicates(t *testing.T) {
	ps := &patterns{
		list:    make([]string, 0),
		control: make(map[string]struct{}),
	}
	ps.add("a")
	ps.add("a")
	ps.add("b")

	wantLen := 2
	if got := len(ps.list); got != wantLen {
		t.Errorf("len(list) = %d, want %d (list = %v)", got, wantLen, ps.list)
	}
}

func TestReadGitIgnore(t *testing.T) {
	root := t.TempDir()
	gi := filepath.Join(root, ".gitignore")
	writeFile(t, gi, "foo\n# comment\n\n  bar  \n")

	got, err := readGitIgnore(gi)
	if err != nil {
		t.Fatalf("readGitIgnore error = %v, want nil", err)
	}
	want := []string{"foo", "bar"}
	if len(got) != len(want) {
		t.Fatalf("readGitIgnore = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("readGitIgnore[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
