package dotfile

import (
	"fmt"
	"os"
	"path"

	"github.com/solerf/dtm/common"
	"github.com/solerf/dtm/profile"
)

var (
	sysDirs = []string{".config", ".local/bin"}
)

type entry struct {
	Key        string `json:"key"`
	SourcePath string `json:"source_path"`
	TargetPath string `json:"target_path"`
}

func (e *entry) Install() error {
	targetDir := path.Dir(e.TargetPath)
	if !common.MustPathExists(targetDir) {
		if err := os.MkdirAll(targetDir, common.FileMode); err != nil {
			return fmt.Errorf("mkdir all [%v]: %w", targetDir, err)
		}
	}

	if common.MustPathExists(e.TargetPath) {
		// check if points to same target that is trying to map
		foundLinkTarget, err := os.Readlink(e.TargetPath)
		if err != nil {
			return fmt.Errorf("reading link [%v]: %w", e.TargetPath, err)
		}

		if foundLinkTarget == e.SourcePath {
			return nil
		}

		// if different it is removed to then be recreated
		if err = os.Remove(e.TargetPath); err != nil {
			return fmt.Errorf("remove existent [%v]: %w", e.TargetPath, err)
		}
	}

	if err := os.Symlink(e.SourcePath, e.TargetPath); err != nil {
		return fmt.Errorf("linking [%v]: %w", e.TargetPath, err)
	}
	return nil
}

func (e *entry) Uninstall() error {
	// it will follow the link, if missing returns false
	if common.MustPathExists(e.TargetPath) {
		if err := os.RemoveAll(e.TargetPath); err != nil {
			return fmt.Errorf("removing [%v]: %w", e.TargetPath, err)
		}
	}
	return nil
}

func newEntry(targetDir string, sourceFullPath string) entry {
	key := profile.RemoveProfile(sourceFullPath)
	return entry{Key: key, SourcePath: sourceFullPath, TargetPath: fmt.Sprintf("%s/%s", targetDir, key)}
}

func batch(targetDir string, sourceFullPath ...string) []entry {
	es := make([]entry, 0, len(sourceFullPath))
	for i := 0; i < len(sourceFullPath); i++ {
		es = append(es, newEntry(targetDir, sourceFullPath[i]))
	}
	return es
}
