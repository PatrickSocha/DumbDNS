package dnsClient

import (
	"context"
	"dumbdns/models"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"dumbdns/database"
	"dumbdns/dohClient"

	dohDns "github.com/likexian/doh-go/dns"
	"github.com/miekg/dns"
)

type DnsServer struct {
	DnsServer *dns.Server
	dohClient *dohClient.DohClient
	db        *database.Database

	refreshFreq time.Duration
}

func Start(port string, dohClient *dohClient.DohClient, db *database.Database) (*DnsServer, error) {
	d := &DnsServer{
		dohClient: dohClient,
		db:        db,
	}

	dns.HandleFunc(".", d.handleDnsRequest)
	d.DnsServer = &dns.Server{Addr: port, Net: "udp"}

	log.Printf("Starting DumbDNS (with AdBlock) at %s\n", d.DnsServer.Addr)
	err := d.DnsServer.ListenAndServe()
	if err != nil {
		return nil, fmt.Errorf("error starting service: %w", err)
	}

	return d, nil
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
		fmt.Errorf("error writing response message: %w", err)
	}
}

func (d *DnsServer) ParseQuery(ctx context.Context, m *dns.Msg) {
	for _, q := range m.Question {
		queryType, err := models.QueryToDoHType(q.Qtype)
		if err != nil {
			fmt.Errorf("error getting query type %w", err)
			return
		}

		records, err := d.getRecords(ctx, q.Name, queryType)
		if err != nil {
			fmt.Errorf("error fetching records: %w", err)
			return
		}

		switch q.Qtype {
		case dns.TypeA:
			for _, v := range records.A {
				rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, v))
				if err != nil {
					fmt.Errorf("error generating query response: %w ", err)
					return
				}
				m.Answer = append(m.Answer, rr)
			}
		case dns.TypeAAAA:
			for _, v := range records.AAAA {
				rr, err := dns.NewRR(fmt.Sprintf("%s AAAA %s", q.Name, v))
				if err != nil {
					fmt.Errorf("error generating query response: %w ", err)
					return
				}
				m.Answer = append(m.Answer, rr)
			}
		case dns.TypeMX:
			for _, v := range records.MX {
				rr, err := dns.NewRR(fmt.Sprintf("%s MX %s", q.Name, v))
				if err != nil {
					fmt.Errorf("error generating query response: %w ", err)
					return
				}
				m.Answer = append(m.Answer, rr)
			}
		case dns.TypeCNAME:
			if records.CNAME == "" {
				return
			}
			rr, err := dns.NewRR(fmt.Sprintf("%s IN CNAME %s", q.Name, records.CNAME))
			if err != nil {
				fmt.Errorf("error generating query response: %w ", err)
				return
			}
			m.Answer = append(m.Answer, rr)
		}
	}
}

func (d *DnsServer) getRecords(ctx context.Context, address string, queryType dohDns.Type) (*models.Record, error) {
	// remove the "." from the end of the passed in address (google.com.)
	address = address[:len(address)-1]

	record, err := d.db.GetRecord(address, queryType)
	if errors.Is(err, database.ErrNotFound) {
		resp := d.dohClient.QueryAuthority(ctx, address, queryType)
		if len(resp) == 0 {
			return record, fmt.Errorf("no response found")
		}

		now := time.Now().UTC()
		record, err := d.db.AddRecord(now, address, queryType, resp)
		if err != nil {
			return record, fmt.Errorf("error adding record: %w", err)
		}

		return record, nil
	}

	return record, nil
}
