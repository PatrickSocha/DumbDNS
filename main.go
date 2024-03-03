package main

import (
	"log"
	"time"

	"dumbdns/database"
	dnsClient "dumbdns/dns"
	"dumbdns/dohClient"

	"github.com/likexian/doh-go"
)

const (
	blockListRefreshRate = 24 * time.Hour
	cacheTTL             = 15 * time.Minute
)

func main() {
	dohClient := dohClient.Start(doh.CloudflareProvider, doh.GoogleProvider)
	defer dohClient.Doh.Close()

	db := database.Start(cacheTTL)
	go db.UpdateBlockList(blockListRefreshRate)

	server := dnsClient.Start(dohClient, db)

	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered. Error:\n", r)
		}
	}()

	defer server.DnsServer.Shutdown()
}
