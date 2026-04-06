package dotfile

import (
	"encoding/json"
	"fmt"
)

func Show(targetDir string) error {
	mapping, err := readMapping(targetDir)
	if err != nil {
		return fmt.Errorf("show: %w", err)
	}

	hierarchy := buildHierarchy(targetDir, mapping.DotFiles)

	marshal, err := json.MarshalIndent(hierarchy, " ", "  ")
	if err != nil {
		return fmt.Errorf("show: writing hierarchy: %w", err)
	}

	fmt.Println(string(marshal))
	return nil
}
