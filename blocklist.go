package main

import (
	"bufio"
	"log"
	"net/http"
	"regexp"
	"time"
)

func updateBlockList() {
	for {
		log.Println("Getting block list")
		blockListDatabase = make(map[string]interface{})

		for _, s := range blockListSources {
			var compRegEx = regexp.MustCompile(s.regex)
			resp, err := http.Get(s.url)
			if err != nil {
				log.Fatalln(err)
			}
			scanner := bufio.NewScanner(resp.Body)

			// populate the list
			for scanner.Scan() {
				v := getParams(compRegEx, scanner.Text())
				if v != nil {

					// domains come in as `domain.com.` so we add a `.` to the end so it can be found in the map
					structuredDomain := *v + "."
					blockListDatabase[structuredDomain] = struct{}{}
				}
			}
		}

		log.Printf("Block list updated with %d records\r\n", len(blockListDatabase))
		time.Sleep(24 * time.Hour)
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
