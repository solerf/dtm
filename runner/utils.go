package runner

import (
	"fmt"
	"os"
	"strings"
)

func tryStat(dir string) error {
	_, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("stat [%s]: %w", dir, err)
	}
	return nil
}

func hasPrefixAny(s string, prefixes []string) (bool, string) {
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return true, prefix
		}
	}
	return false, ""
}
