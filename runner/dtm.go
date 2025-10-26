package runner

import (
	"errors"
	"fmt"
	"log"
	"os"
)

var (
	sysDirs = []string{".config", ".local/bin"}
)

func Install(profileNames []string, sourceDir string, targetDir string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("installer: %w", err)
	}

	if err = tryStat(sourceDir); err != nil {
		return fmt.Errorf("installer: %w", err)
	}

	if err = tryStat(targetDir); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("installer: %w", err)
		}

		if errMkdir := os.MkdirAll(targetDir, fileMode); errMkdir != nil {
			return fmt.Errorf("installer: %w", errMkdir)
		}
	}

	profiles, err := toProfiles(sourceDir, profileNames)
	if err != nil {
		return fmt.Errorf("installer: %w", err)
	}

	notIgnoreEntryF := notIgnoreFilter(sourceDir)

	dotfiles, err := collectDotFiles(notIgnoreEntryF, profiles...)
	if err != nil {
		return err
	}

	msgTempl := "mapping [%v] to [%v]: %v"
	runAction(dotfiles, func(dt dotFile) string {
		if e := dt.install(targetDir); e != nil {
			return fmt.Sprintf(msgTempl, dt.Key, targetDir, e)
		}
		return fmt.Sprintf(msgTempl, dt.Key, targetDir, "OK")
	})

	if err = writeMappings(homeDir, targetDir, profiles, dotfiles); err != nil {
		return err
	}
	return nil
}

func Uninstall() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("uninstaller: %w", err)
	}

	mappings, err := readMappings(homeDir)
	if err != nil {
		return fmt.Errorf("uninstaller: %w", err)
	}

	msgTempl := "removed [%v]: %v"
	runAction(mappings.DotFiles, func(d dotFile) string {
		if e := d.uninstall(mappings.InstallDir); e != nil {
			return fmt.Sprintf(msgTempl, d.Key, e)
		}
		return fmt.Sprintf(msgTempl, d.Key, "OK")
	})

	if err = deleteMappings(homeDir); err != nil {
		return fmt.Errorf("uninstaller: %w", err)
	}
	return nil
}

func runAction(dotfiles []dotFile, run func(dotFile) string) {
	for i := 0; i < len(dotfiles); i++ {
		log.Println(run(dotfiles[i]))
	}
}
