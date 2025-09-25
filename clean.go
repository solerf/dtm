package main

import (
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"os"
	"path/filepath"
	"strings"
)

var (
	sysDirs = []string{".config", ".local/bin"}
)

func Clean() error {
	mappings, err := getMappings()
	if err != nil {
		return err
	}

	for l := range maps.Keys(collectLinksFromMappings(mappings)) {
		if e := os.RemoveAll(l); e != nil {
			return e
		}
		log.Printf("removed [%v]", l)
	}
	return nil
}

func collectLinksFromMappings(mappings Mappings) map[string]struct{} {
	linksToRemove := make(map[string]struct{})

	for link := range maps.Keys(mappings.Entries) {
		targetToRemove := link

		var targetNoSysDir string
		var systemDir string
		for _, sd := range sysDirs {
			if strings.HasPrefix(link, sd) {
				targetNoSysDir = strings.ReplaceAll(link, sd+"/", "")
				systemDir = sd
				break
			}
		}

		// as it merges, check if it needs to delete the whole new config folder
		// or delete just the symlink
		if len(targetNoSysDir) != 0 {
			linkFirstPart := strings.Split(targetNoSysDir, string(filepath.Separator))[0]
			linkFirstDir := filepath.Join(mappings.InstallDir, systemDir, linkFirstPart)

			if stat, errStat := os.Stat(linkFirstDir); errStat == nil {
				if stat.IsDir() {
					targetToRemove = linkFirstDir
				} else {
					targetToRemove = filepath.Join(mappings.InstallDir, link)
				}
			}
		}
		linksToRemove[targetToRemove] = struct{}{}
	}
	return linksToRemove
}

func getMappings() (Mappings, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Mappings{}, err
	}

	mappingsPath := filepath.Join(homeDir, dtmMappingsFileName)
	file, err := os.ReadFile(mappingsPath)
	if err != nil {
		return Mappings{}, fmt.Errorf("failed reading mappings in [%s]: %w", homeDir, err)
	}

	var mappings Mappings
	err = json.Unmarshal(file, &mappings)
	if err != nil {
		return Mappings{}, fmt.Errorf("failed deserialize mappings in [%s]: %w", homeDir, err)
	}
	return mappings, nil
}
