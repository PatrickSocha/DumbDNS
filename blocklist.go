package main

import (
	"bufio"
	"log"
	"net/http"
	"regexp"
)

func updateBlockList() {
	log.Println("Getting block list")

	var compRegEx = regexp.MustCompile(`0.0.0.0\s+(?P<url>\S+)`)
	resp, err := http.Get("https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts")
	if err != nil {
		log.Fatalln(err)
	}
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		v := getParams(compRegEx, scanner.Text())
		if v != nil {

			// domains come in as `domain.com.` so we add a `.` to the end so it can be found in the map
			structuredDomain := *v + "."
			blockListDatabase[structuredDomain] = struct{}{}
		}
	}

	log.Println("Block list updated")
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
