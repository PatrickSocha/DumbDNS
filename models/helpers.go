package models

import (
	"errors"

	dohDns "github.com/likexian/doh-go/dns"
	"github.com/miekg/dns"
)

func QueryToDoHType(t uint16) (dohDns.Type, error) {
	switch t {
	case dns.TypeA:
		return dohDns.TypeA, nil
	case dns.TypeAAAA:
		return dohDns.TypeAAAA, nil
	case dns.TypeMX:
		return dohDns.TypeMX, nil
	case dns.TypeCNAME:
		return dohDns.TypeCNAME, nil
	case dns.TypeNS:
		return dohDns.TypeNS, nil
	//case dns.TypeSRV:
	//	return dohDns.TypeSRV

	default:
		return "", errors.New("query type not supported")
	}
}
