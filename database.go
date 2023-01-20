package main

import (
	"sync"
	"time"
)

type record struct {
	expiresAt time.Time
	ips       []string
}

type sources struct {
	regex string
	url   string
}

var dbMux = &sync.RWMutex{}
var database = map[string]record{}
var blockMux = &sync.RWMutex{}
var blockListDatabase = map[string]interface{}{}
var blockListSources = []sources{
	{
		regex: `0.0.0.0\s+(?P<url>\S+)`,
		url:   "https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts",
	},
	{
		regex: `127.0.0.1\s+(?P<url>\S+)`,
		url:   "https://adaway.org/hosts.txt",
	},
	{
		regex: `(?P<url>\S+)`,
		url:   "https://v.firebog.net/hosts/Easyprivacy.txt",
	},
	{
		regex: `0.0.0.0\s+(?P<url>\S+)`,
		url:   "https://raw.githubusercontent.com/d3ward/toolz/master/src/d3host.txt",
	},
	{
		regex: `(?P<url>\S+)`,
		url:   "https://blokada.org/blocklists/ddgtrackerradar/standard/hosts.txt",
	},
	{
		regex: `0.0.0.0\s+(?P<url>\S+)`,
		url:   "https://raw.githubusercontent.com/lightswitch05/hosts/master/docs/lists/ads-and-tracking-extended.txt",
	},
	{
		regex: `0.0.0.0\s+(?P<url>\S+)`,
		url:   "https://raw.githubusercontent.com/lightswitch05/hosts/master/docs/lists/tracking-aggressive-extended.txt",
	},
}

var whitelistDatabase = map[string]interface{}{
	"spclient.wg.spotify.com": struct{}{}, // spotify
}
