package database

import (
	"bufio"
	"log"
	"net/http"
	"regexp"
	"time"
)

func (db *Database) UpdateBlockList(refreshRate time.Duration) {
	config, err := readConfigFromDisk()
	if err != nil {
		log.Fatalf("error reading config: %v", err)
		return
	}
	db.config = config

	for {
		log.Println("Getting block list")
		db.blockMux.Lock()
		db.blockListDatabase = make(map[string]interface{})

		for _, s := range config.blocklists {
			var compRegEx = regexp.MustCompile(s.Regex)
			resp, err := http.Get(s.Url)
			if err != nil {
				log.Println("Error:", err)
			}
			scanner := bufio.NewScanner(resp.Body)

			// populate the list
			for scanner.Scan() {
				v := getParams(compRegEx, scanner.Text())
				if v != nil {
					db.blockListDatabase[*v] = struct{}{}
				}
			}
		}
		for domain, _ := range config.whitelistDomains {
			delete(db.blockListDatabase, domain)
		}
		db.blockMux.Unlock()
		log.Printf("Block list updated with %d records\r\n", len(db.blockListDatabase))

		log.Println("Purging old database records")
		db.dbMux.Lock()
		for domain, v := range db.database {
			if time.Now().After(v.ExpiresAt) {
				delete(db.database, domain)
			}
		}
		db.dbMux.Unlock()

		log.Println("Refresh Go routine sleeping")
		time.Sleep(refreshRate)
	}
}

func getParams(compRegEx *regexp.Regexp, url string) *string {
	match := compRegEx.FindStringSubmatch(url)

	paramsMap := make(map[string]string)
	for i, name := range compRegEx.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}
	var domain *string
	if v, ok := paramsMap["url"]; ok {
		domain = &v
	}

	return domain
}
