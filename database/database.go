package database

import (
	"dumbdns/models"
	"errors"
	"sync"
	"time"

	"github.com/likexian/doh-go/dns"
)

var (
	ErrNotFound = errors.New("not found")
)

type Database struct {
	TTL               time.Duration
	database          map[string]*models.Record
	config            *Config
	dbMux             *sync.RWMutex
	blockMux          *sync.RWMutex
	blockListDatabase map[string]interface{}
}

func Start(ttl time.Duration) *Database {
	db := &Database{
		TTL:               ttl,
		dbMux:             &sync.RWMutex{},
		blockMux:          &sync.RWMutex{},
		database:          map[string]*models.Record{},
		blockListDatabase: map[string]interface{}{},
	}

	return db
}

func (db *Database) GetRecord(address string, queryType dns.Type) (*models.Record, error) {
	// Check custom hosts file for host:ip mapping file
	// e.g: archive.is blocks CloudFlare DNS, so we add
	// a manual mapping to get around that.
	if ip, ok := db.config.hosts[address]; ok {
		return &models.Record{A: []string{ip}}, nil
	}

	db.blockMux.RLock()
	// Check if in block list
	if _, blocked := db.blockListDatabase[address]; blocked {
		db.blockMux.RUnlock()
		return &models.Record{
			A:     []string{"127.0.0.1"},
			AAAA:  []string{"::1"},
			NS:    []string{"localhost"},
			MX:    []string{"localhost"},
			SRV:   []string{"_http._tcp.local."},
			CNAME: "localhost",
		}, nil
	}
	db.blockMux.RUnlock()

	// Now we can safely lock the database for record checking
	db.dbMux.RLock()
	defer db.dbMux.RUnlock()
	if record, ok := db.database[address]; ok {
		if time.Now().After(record.ExpiresAt) {
			// Expired record, delete and return not found
			db.dbMux.RUnlock() // Unlock the read lock before locking for delete
			db.dbMux.Lock()    // Now acquire the write lock
			delete(db.database, address)
			db.dbMux.Unlock() // Unlock the write lock after deleting
			return nil, ErrNotFound
		}

		if hasQueryType(record, queryType) {
			return record, nil
		}
	}

	return nil, ErrNotFound
}

func hasQueryType(r *models.Record, queryType dns.Type) bool {
	if r == nil {
		return false
	}

	switch queryType {
	case dns.TypeA:
		return len(r.A) > 0
	case dns.TypeAAAA:
		return len(r.AAAA) > 0
	case dns.TypeNS:
		return len(r.NS) > 0
	case dns.TypeMX:
		return len(r.MX) > 0
	//case dns.TypeSRV:
	//	return len(r.SRV) > 0
	case dns.TypeCNAME:
		return r.CNAME != ""
	default:
		return false
	}
}

func (db *Database) AddRecord(now time.Time, address string, queryType dns.Type, recordValue []string) (*models.Record, error) {
	db.dbMux.RLock()
	defer db.dbMux.RUnlock()
	record, ok := db.database[address]
	if !ok {
		// We create a new record to be populated
		record = &models.Record{}
	}

	switch queryType {
	case dns.TypeA:
		record.A = recordValue
	case dns.TypeAAAA:
		record.AAAA = recordValue
	case dns.TypeNS:
		record.NS = recordValue
	case dns.TypeMX:
		record.MX = recordValue
	//case dns.TypeSRV:
	//	return len(r.SRV) > 0
	case dns.TypeCNAME:
		if len(recordValue) == 1 {
			record.CNAME = recordValue[0]
		}
	default:
		return nil, errors.New("could not update value for query type")
	}

	if record.ExpiresAt.IsZero() {
		record.ExpiresAt = now.Add(db.TTL)
	}

	db.database[address] = record

	return record, nil
}
