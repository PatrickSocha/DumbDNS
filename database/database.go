package database

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
)

type Record struct {
	ExpiresAt time.Time
	IPs       []string
}

type sources struct {
	regex string
	url   string
}

type Database struct {
	TTL time.Duration

	dbMux             *sync.RWMutex
	database          map[string]Record
	blockMux          *sync.RWMutex
	blockListDatabase map[string]interface{}
}

func Start(ttl time.Duration) *Database {
	db := &Database{
		TTL:               ttl,
		dbMux:             &sync.RWMutex{},
		blockMux:          &sync.RWMutex{},
		database:          map[string]Record{},
		blockListDatabase: map[string]interface{}{},
	}

	return db
}

// getIPs checks if the domain is blocked
// then checks if we have a local record of it
// then checks if it's still within our TTL (fetches otherwise)
// then returns, otherwise NotFound error
func (db *Database) GetRecord(address string) (*Record, error) {
	db.blockMux.RLock()

	// Check hosts file
	if ip, ok := hosts[address]; ok {
		db.blockMux.RUnlock()
		return &Record{IPs: []string{ip}}, nil
	}

	// Check if blocked
	if _, blocked := db.blockListDatabase[address]; blocked {
		db.blockMux.RUnlock()
		return &Record{
			IPs: []string{"127.0.0.1"},
		}, nil
	}

	// purge old records && return not found
	// Otherwise return record.
	if record, ok := db.database[address]; ok {
		if time.Now().After(record.ExpiresAt) {
			db.dbMux.Lock()
			delete(db.database, address)
			db.dbMux.Unlock()

			db.blockMux.RUnlock()
			return nil, ErrNotFound
		}

		db.blockMux.RUnlock()
		return &record, nil
	}
	db.blockMux.RUnlock()

	return nil, ErrNotFound
}

func (db *Database) AddRecord(address string, ips []string) Record {
	db.dbMux.RLock()
	record, ok := db.database[address]
	db.dbMux.RUnlock()
	if !ok {
		r := db.addToDatabase(address, ips)
		return r
	}

	if time.Now().After(record.ExpiresAt) {
		record = db.addToDatabase(address, ips)
	}

	return record
}

// addToDatabase adds domains to the local cache
func (db *Database) addToDatabase(address string, ips []string) Record {
	r := Record{
		ExpiresAt: time.Now().Add(db.TTL),
		IPs:       ips,
	}

	// if there's a DOH failure, don't cache. Return empty result.
	if len(ips) == 0 {
		return r
	}

	db.dbMux.Lock()
	db.database[address] = r
	db.dbMux.Unlock()

	return r
}
