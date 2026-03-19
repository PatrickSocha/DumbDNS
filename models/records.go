package models

import "time"

type Record struct {
	ExpiresAt time.Time

	A     []string
	AAAA  []string
	NS    []string
	MX    []string
	SRV   []string
	TXT   []string
	CNAME string
	SOA   string
	PTR   []string
}
