package dohClient

import (
	"context"
	"dumbdns/models"
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
func (d *DohClient) QueryAuthority(ctx context.Context, address string, questionQueryType dohDns.Type) []string {
	dohResp, err := d.Doh.Query(ctx, dohDns.Domain(address), questionQueryType)
	if err != nil {
		log.Printf("%s : doh query failed, retrying: %s", address, err.Error())
		dohResp, err = d.Doh.Query(ctx, dohDns.Domain(address), questionQueryType)
		if err != nil {
			log.Printf("%s : failed, giving up: %s", address, err.Error())
			return []string{}
		}
	}
	if dohResp == nil {
		return []string{}
	}

	queryResp := []string{}
	for _, answer := range dohResp.Answer {
		// DNS provider can sometimes return multiple types, only return the one we want
		// e.g: ipv6.googlg.com returns type 5 (CNAME) and 28 (AAAA) which would break AAAA response
		responseQueryType, err := models.QueryToDoHType(uint16(answer.Type))
		if err != nil {
			log.Printf("error finding doh query type response: %s", err)
			return []string{}
		}
		if responseQueryType != questionQueryType {
			continue
		}

		queryResp = append(queryResp, answer.Data)
	}

	return queryResp
}
