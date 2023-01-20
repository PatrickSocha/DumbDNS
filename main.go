package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// Debugging good to find a blocked domains on mobile.
// Default is false to preserve privacy.
var debugging = false
var TTL = 15 * time.Minute
var refreshFreq = 6 * time.Hour

func main() {
	dns.HandleFunc(".", handleDnsRequest)

	server := &dns.Server{Addr: ":53", Net: "udp"}
	log.Printf("Starting DumbDNS (with AdBlock) at %s\n", server.Addr)
	go updateBlockList()

	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to start server: %s\n ", err.Error())
	}

	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered. Error:\n", r)
		}
	}()

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
	cleanIP := strings.Split(w.RemoteAddr().String(), ":")
	ip := net.ParseIP(cleanIP[0])
	if !(ip.IsPrivate() || ip.IsLoopback()) {
		w.Close()
		return
	}

	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false
	switch r.Opcode {
	case dns.OpcodeQuery:
		parseQuery(m)
	}

	err := w.WriteMsg(m)
	if err != nil {
		log.Println(err)
	}
}

// getIPs checks if the domain is blocked
// then checks if we have a local record of it
// then checks if it's still within our TTL (fetches otherwise)
// then returns
func getIPs(address string) []string {
	blockMux.RLock()
	if _, blocked := blockListDatabase[address]; blocked {
		if debugging {
			log.Println("blocked", address)
		}
		blockMux.RUnlock()
		return []string{"127.0.0.1"}
	}
	blockMux.RUnlock()

	dbMux.RLock()
	record, ok := database[address]
	dbMux.RUnlock()
	if !ok {
		r := addToDatabase(address)
		return r.ips
	}

	if time.Now().After(record.expiresAt) {
		record = addToDatabase(address)
	}

	return record.ips
}

// QueryAuthority uses 1.1.1.1 to fetch actual DNS data
func QueryAuthority(address string) []string {
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(5000),
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
	dbMux.Lock()
	database[address] = r
	dbMux.Unlock()

	return r
}
