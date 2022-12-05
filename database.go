package main

import "time"

type record struct {
	expiresAt time.Time
	ips       []string
}

var database = map[string]record{}
var blockListDatabase = map[string]interface{}{}
var whitelistDatabase = map[string]interface{}{
	"spclient.wg.spotify.com.": struct{}{},
}
