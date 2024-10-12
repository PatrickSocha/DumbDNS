package models

import "time"

type Record struct {
	ExpiresAt time.Time

	A     []string
	AAAA  []string
	NS    []string
	MX    []string
	SRV   []string
	CNAME string
}
