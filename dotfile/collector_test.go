package dotfile

import (
	"strings"
	"testing"

	"github.com/solerf/dtm/profile"
)

var ignore = func(_ string) bool { return false }

func Test_Collect_Dir_Entries(t *testing.T) {
	entries, _ := collectDirEntries(ignore, "testdata/source")
	if actual := entries; len(actual) == 0 {
		t.Errorf("unexpected empty dirs")
	}
}

func Test_Collect_From_Paths(t *testing.T) {
	uniq, dup, _ := collectPaths(ignore, "testdata/source/_a", "testdata/source/_c")

	if len(uniq) != 2 {
		t.Fatalf("unexpected size uniques, expected 2, got %d", len(uniq))
	}

	if len(dup) != 2 {
		t.Fatalf("unexpected size duplicates, expected 2, got %d", len(dup))
	}
}

func Test_Collect_DotFiles(t *testing.T) {
	dotfiles, _ := Collect(
		ignore,
		profile.New("testdata/source", "a"),
		profile.New("testdata/source", "b"),
		profile.New("testdata/source", "c"),
	)

	if len(dotfiles) != 7 {
		t.Errorf("unexpected size duplicates, expected 7, got %d", len(dotfiles))
	}
}

func Test_Collect_Ignore_Entry(t *testing.T) {
	dotfiles, _ := Collect(
		func(s string) bool { return !strings.HasSuffix(s, "_IGNORE") },
		profile.New("testdata/source", "a"),
	)

	for _, d := range dotfiles {
		if strings.Contains(d, "IGNORE") {
			t.Fatalf("unexpected ignored entry, %v", d)
		}
	}
}
