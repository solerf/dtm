package common

import (
	"errors"
	"os"
	"path/filepath"
)

const (
	FileMode      = os.FileMode(0766)
	PathSeparator = string(filepath.Separator)
)

func MustPathExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			panic(err)
		}
		return false
	}
	return true
}

func HasAny(f func(string, string) bool, s string, prefixes []string) (bool, string) {
	for _, prefix := range prefixes {
		if f(s, prefix) {
			return true, prefix
		}
	}
	return false, ""
}
