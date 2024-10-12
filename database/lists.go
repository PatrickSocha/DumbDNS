package database

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type Sources struct {
	Regex string `json:"regex"`
	Url   string `json:"url"`
}

type ConfigJson struct {
	BlockLists       []Sources         `json:"blockLists"`
	WhitelistDomains []string          `json:"whiteList"`
	Hosts            map[string]string `json:"hostsFile"`
}

type Config struct {
	blocklists       []Sources
	whitelistDomains map[string]interface{}
	hosts            map[string]string
}

func readConfigFromDisk() (*Config, error) {
	configFile := "dumbdns.json"
	file, err := os.Open("./" + configFile)
	if err != nil {
		exePath, err := os.Executable()
		if err != nil {
			log.Fatalf("error getting executable path: %v", err)
		}
		wd := filepath.Dir(exePath)

		fullPath := filepath.Join(wd, configFile)
		file, err = os.Open(fullPath)
		if err != nil {
			log.Fatalf("could not open file with full path: %v", err)
		}
	}
	defer file.Close()

	var cj ConfigJson
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cj)
	if err != nil {
		return nil, fmt.Errorf("error decoding JSON: %w", err)
	}

	domainMap := make(map[string]interface{})
	for _, domain := range cj.WhitelistDomains {
		domainMap[domain] = struct{}{}
	}

	return &Config{
		blocklists:       cj.BlockLists,
		whitelistDomains: domainMap,
		hosts:            cj.Hosts,
	}, nil
}
