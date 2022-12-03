package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/miekg/dns"
)

var TTL time.Duration = 5

func main() {
	// attach request handler func
	dns.HandleFunc(".", handleDnsRequest)

	// start server
	server := &dns.Server{Addr: ":53", Net: "udp"}
	log.Printf("Starting DumbDNS (with AdBlock) at %s\n", server.Addr)
	updateBlockList()

	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to start server: %s\n ", err.Error())
	}

	defer server.Shutdown()
}

func parseQuery(m *dns.Msg) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			ips := getIPs(q.Name)
			if len(ips) > 0 {
				for _, ip := range ips {
					rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
					if err == nil {
						m.Answer = append(m.Answer, rr)
					}
				}
			}
		}
	}
}

func handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		parseQuery(m)
	}

	w.WriteMsg(m)
}

// getIPs checks if the domain is blocked
// then checks if we have a local record of it
// then checks if it's still within our TTL (fetches otherwise)
// then returns
func getIPs(address string) []string {
	if _, ok := blockListDatabase[address]; ok {
		log.Println("Blocked", address)
		return []string{"127.0.0.1"}
	}

	record, ok := database[address]
	if !ok {
		r := addToDatabase(address)
		return r.ips
	}

	if time.Now().After(record.expiresAt) {
		record = addToDatabase(address)
	}

	return record.ips
}

// QueryAuthority uses 8.8.8.8 to fetch actual DNS data
func QueryAuthority(address string) []string {
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			return d.DialContext(ctx, network, "8.8.8.8:53")
		},
	}
	ip, _ := r.LookupHost(context.Background(), address)

	return ip
}

// addToDatabase adds domains to the local cache
func addToDatabase(address string) record {
	authorityResponse := QueryAuthority(address)
	r := record{
		expiresAt: time.Now().Add(TTL * time.Minute),
		ips:       authorityResponse,
	}
	database[address] = r

	return r
}
