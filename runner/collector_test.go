package runner

import (
	"testing"
)

var notIgnore = func(_ string) bool { return true }

func Test_Collect_Dir_Entries(t *testing.T) {
	entries, _ := collectDirEntries(notIgnore, "testdata/source")
	if actual := entries; len(actual) == 0 {
		t.Errorf("unexpected empty dirs")
	}
}

func Test_Collect_From_Paths(t *testing.T) {
	uniq, dup, _ := collectFromPaths(notIgnore, "testdata/source/_a", "testdata/source/_c")

	if len(uniq) != 1 {
		t.Errorf("unexpected size uniques, expected 1, got %d", len(uniq))
	}

	if len(dup) != 2 {
		t.Errorf("unexpected size duplicates, expected 2, got %d", len(dup))
	}
}

func Test_Collect_DotFiles(t *testing.T) {
	dotfilesIter, _ := collectDotFiles(
		notIgnore,
		newProfile("testdata/source", "a"),
		newProfile("testdata/source", "b"),
		newProfile("testdata/source", "c"),
	)

	dotfiles := dotfilesIter
	if len(dotfiles) != 6 {
		t.Errorf("unexpected size duplicates, expected 6, got %d", len(dotfiles))
	}

}
