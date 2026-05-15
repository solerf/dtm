package dotfile

import (
	"fmt"
)

func Show(targetDir string) error {
	mapping, err := readMapping(targetDir)
	if err != nil {
		return fmt.Errorf("show: %w", err)
	}

	hierarchy := buildHierarchy(targetDir, mapping.DotFiles)

	h, err := hierarchy.toJson()
	if err != nil {
		return fmt.Errorf("show: writing hierarchy: %w", err)
	}

	fmt.Println(h)
	return nil
}
