package dotfile

import (
	"fmt"
	"log"
	"maps"
	"os"
	"path"
	"strings"

	"github.com/solerf/dtm/common"
)

func Uninstall(targetDir string) error {
	mapping, err := readMapping(targetDir)
	if err != nil {
		return fmt.Errorf("uninstall: %w", err)
	}

	remainingDirs := make(map[string]struct{}, len(mapping.DotFiles)/2)

	msgTempl := "removed [%v]: %v"
	for i := 0; i < len(mapping.DotFiles); i++ {
		e := mapping.DotFiles[i]
		if err = e.Uninstall(); err != nil {
			log.Printf("[ERROR] %v", fmt.Sprintf(msgTempl, e.TargetPath, err))
			continue
		}
		log.Println(fmt.Sprintf(msgTempl, e.TargetPath, "OK"))

		// collect any remaining dir previously installed
		if isSysDir, sysDir := common.HasAny(strings.Contains, e.TargetPath, sysDirs); isSysDir {
			base := path.Dir(e.TargetPath)
			if split := strings.Split(base, sysDir); len(split) > 1 {
				if len(split[1]) > 0 {
					base = path.Join(split[0], sysDir, split[1])
					remainingDirs[base] = struct{}{}
				}
			}
		}
	}

	// remove remaining dirs
	for d := range maps.Keys(remainingDirs) {
		if common.MustPathExists(d) {
			if err = os.RemoveAll(d); err != nil {
				log.Printf("[ERROR] %v", fmt.Sprintf(msgTempl, d, err))
				continue
			}
			log.Print(fmt.Sprintf(msgTempl, d, "OK"))
		}
	}

	if err = mapping.Delete(); err != nil {
		return fmt.Errorf("uninstall: %w", err)
	}
	return nil
}
