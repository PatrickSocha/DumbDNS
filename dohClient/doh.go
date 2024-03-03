package dohClient

import (
	"context"
	"log"

	"github.com/likexian/doh-go"
	dohDns "github.com/likexian/doh-go/dns"
)

type DohClient struct {
	Doh *doh.DoH
}

func Start(provider ...int) *DohClient {
	return &DohClient{
		Doh: doh.Use(provider...),
	}
}

// QueryAuthority makes DNS over HTTPS request
func (d *DohClient) QueryAuthority(ctx context.Context, address string) []string {
	rsp, err := d.Doh.Query(ctx, dohDns.Domain(address), dohDns.TypeA)
	if err != nil {
		log.Printf("%s : doh query failed, retrying: %s", address, err.Error())
		rsp, err = d.Doh.Query(ctx, dohDns.Domain(address), dohDns.TypeA)
		if err != nil {
			log.Printf("%s : doh query failed, giving up: %s", address, err.Error())
		}
	}

	resp := []string{}
	if rsp == nil {
		return []string{}
	}

	for _, answer := range rsp.Answer {
		resp = append(resp, answer.Data)
	}

	return resp
}
