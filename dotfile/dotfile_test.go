package dotfile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/solerf/dtm/profile"
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

func TestItem_Install(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T) (sourcePath, symLink string)
		wantStatus Status
		wantErr    bool
	}{
		{
			name: "fresh_install",
			setup: func(t *testing.T) (string, string) {
				dir := t.TempDir()
				src := filepath.Join(dir, "src")
				writeFile(t, src, "x")
				return src, filepath.Join(dir, "target", ".rc")
			},
			wantStatus: Ok,
		},
		{
			name: "already_linked_to_same_source",
			setup: func(t *testing.T) (string, string) {
				dir := t.TempDir()
				src := filepath.Join(dir, "src")
				writeFile(t, src, "x")
				link := filepath.Join(dir, ".rc")
				if err := os.Symlink(src, link); err != nil {
					t.Fatalf("setup symlink: %v", err)
				}
				return src, link
			},
			wantStatus: Skipped,
		},
		{
			name: "link_points_to_different_source",
			setup: func(t *testing.T) (string, string) {
				dir := t.TempDir()
				src := filepath.Join(dir, "src")
				other := filepath.Join(dir, "other")
				writeFile(t, src, "x")
				writeFile(t, other, "y")
				link := filepath.Join(dir, ".rc")
				if err := os.Symlink(other, link); err != nil {
					t.Fatalf("setup symlink: %v", err)
				}
				return src, link
			},
			wantStatus: Failed,
			wantErr:    true,
		},
		{
			name: "regular_file_in_the_way",
			setup: func(t *testing.T) (string, string) {
				dir := t.TempDir()
				src := filepath.Join(dir, "src")
				writeFile(t, src, "x")
				link := filepath.Join(dir, ".rc")
				writeFile(t, link, "blocking")
				return src, link
			},
			wantStatus: Failed,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, link := tt.setup(t)
			item := &Item{SourcePath: src, SymLink: link}
			st, err := item.install()
			if st != tt.wantStatus {
				t.Errorf("install() status = %v, want %v", st, tt.wantStatus)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("install() err = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestItem_Uninstall(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T) string
		wantStatus Status
		wantErr    bool
		wantExists bool
	}{
		{
			name: "removes_existing_symlink",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				src := filepath.Join(dir, "src")
				writeFile(t, src, "x")
				link := filepath.Join(dir, ".rc")
				if err := os.Symlink(src, link); err != nil {
					t.Fatalf("setup: %v", err)
				}
				return link
			},
			wantStatus: Ok,
		},
		{
			name: "missing_symlink_skipped",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "missing")
			},
			wantStatus: Skipped,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			link := tt.setup(t)
			item := &Item{SymLink: link}
			st, err := item.uninstall()
			if st != tt.wantStatus {
				t.Errorf("uninstall() status = %v, want %v", st, tt.wantStatus)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("uninstall() err = %v, wantErr = %v", err, tt.wantErr)
			}
			if _, statErr := os.Lstat(link); !os.IsNotExist(statErr) && !tt.wantExists {
				t.Errorf("symlink %q still exists after uninstall", link)
			}
		})
	}
}

func TestItem_LinkExists(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	writeFile(t, src, "x")
	link := filepath.Join(dir, "link")
	if err := os.Symlink(src, link); err != nil {
		t.Fatalf("setup: %v", err)
	}
	regular := filepath.Join(dir, "reg")
	writeFile(t, regular, "x")

	tests := []struct {
		name string
		path string
		want bool
	}{
		{"existing_symlink", link, true},
		{"regular_file_not_symlink", regular, false},
		{"missing", filepath.Join(dir, "ghost"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &Item{SymLink: tt.path}
			got, err := item.linkExists()
			if err != nil {
				t.Fatalf("linkExists() err = %v, want nil", err)
			}
			if got != tt.want {
				t.Errorf("linkExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestItem_SourceExists(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "f")
	writeFile(t, existing, "x")

	tests := []struct {
		name string
		path string
		want bool
	}{
		{"present", existing, true},
		{"missing", filepath.Join(dir, "ghost"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &Item{SourcePath: tt.path}
			got, err := item.sourceExists()
			if err != nil {
				t.Fatalf("sourceExists() err = %v, want nil", err)
			}
			if got != tt.want {
				t.Errorf("sourceExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewEntry(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	src := filepath.Join(dir, "src", "f")
	writeFile(t, src, "x")
	prof := &profile.Info{Name: "p", Path: filepath.Join(dir, "src")}

	t.Run("existing_source", func(t *testing.T) {
		got, err := newEntry(target, prof, ".f", src)
		if err != nil {
			t.Fatalf("newEntry err = %v, want nil", err)
		}
		want := Item{
			Profile:    prof,
			Key:        ".f",
			SourcePath: src,
			SymLink:    filepath.Join(target, ".f"),
		}
		if got != want {
			t.Errorf("newEntry = %+v, want %+v", got, want)
		}
	})

	t.Run("missing_source_errors", func(t *testing.T) {
		if _, err := newEntry(target, prof, ".g", filepath.Join(dir, "src", "ghost")); err == nil {
			t.Errorf("newEntry err = nil, want error for missing source")
		}
	})
}

func TestCreateIfMissing(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "a", "b", "c")
	if err := createIfMissing(target); err != nil {
		t.Fatalf("createIfMissing err = %v, want nil", err)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("stat after createIfMissing: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("createIfMissing path is not a dir: mode = %v", info.Mode())
	}

	if err := createIfMissing(target); err != nil {
		t.Errorf("createIfMissing on existing dir err = %v, want nil", err)
	}
}

func TestIsSymlinkMode(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	writeFile(t, src, "x")
	link := filepath.Join(dir, "l")
	if err := os.Symlink(src, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{"regular_file", src, false},
		{"symlink", link, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := os.Lstat(tt.path)
			if err != nil {
				t.Fatalf("lstat: %v", err)
			}
			if got := isSymlinkMode(info.Mode()); got != tt.want {
				t.Errorf("isSymlinkMode = %v, want %v", got, tt.want)
			}
		})
	}
}
