package dnsClient

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"dumbdns/database"
	"dumbdns/dohClient"
	"github.com/miekg/dns"
)

type DnsServer struct {
	DnsServer *dns.Server
	dohClient *dohClient.DohClient
	db        *database.Database

	refreshFreq time.Duration
}

func Start(dohClient *dohClient.DohClient, db *database.Database) *DnsServer {
	d := &DnsServer{
		dohClient: dohClient,
		db:        db,
	}

	dns.HandleFunc(".", d.handleDnsRequest)
	d.DnsServer = &dns.Server{Addr: ":53", Net: "udp"}

	log.Printf("Starting DumbDNS (with AdBlock) at %s\n", d.DnsServer.Addr)
	err := d.DnsServer.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to start server: %s\n ", err.Error())
	}

	return d
}

func (d *DnsServer) handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	cleanIP := strings.Split(w.RemoteAddr().String(), ":")
	ip := net.ParseIP(cleanIP[0])
	if !(ip.IsPrivate() || ip.IsLoopback()) {
		w.Close()
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false
	switch r.Opcode {
	case dns.OpcodeQuery:
		d.ParseQuery(ctx, m)
	}

	err := w.WriteMsg(m)
	if err != nil {
		log.Println(err)
	}
}

func (d *DnsServer) ParseQuery(ctx context.Context, m *dns.Msg) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			ips := d.getIPs(ctx, q.Name)
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

func (d *DnsServer) getIPs(ctx context.Context, address string) []string {
	// remove the "." from the passed in address
	address = address[:len(address)-1]

	record, err := d.db.GetRecord(address)
	if errors.Is(err, database.ErrNotFound) {
		ips := d.dohClient.QueryAuthority(ctx, address)
		record := d.db.AddRecord(address, ips)

		return record.IPs
	}

	return record.IPs
}
