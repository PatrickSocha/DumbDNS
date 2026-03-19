package models

type Config struct {
	Blocklists       []Sources
	WhitelistDomains map[string]interface{}
	Hosts            map[string]string
}

type Sources struct {
	Regex string `json:"regex"`
	Url   string `json:"url"`
}
