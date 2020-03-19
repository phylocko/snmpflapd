package main

import (
	"fmt"
	"log"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// Connector is an object to connect the database
type Connector struct {
	db *sqlx.DB
	mx sync.Mutex
}

var connector *Connector

// MakeDB returns an SQL Connector object to make queries
func MakeDB(dbName, dbUser, dbPassword string) (*Connector, error) {

	dataSourceName := fmt.Sprintf("%s:%s@/%s", dbUser, dbPassword, dbName)
	db, err := sqlx.Open("mysql", dataSourceName)
	if err != nil {
		return nil, err
	}

	c := &Connector{db: db}
	if err := c.db.Ping(); err != nil {
		return nil, err
	}

	logVerbose("DB connection OK")
	return c, nil
}

// CleanUp deletes old cached values from DB
func (c *Connector) CleanUp() {

	c.mx.Lock()
	defer c.mx.Unlock()

	log.Printf("Cleanup DB started")
	cleanUpHostnameSQL := `DELETE FROM cache_hostname WHERE time < now() - INTERVAL ? MINUTE;`
	cleanUpIfNameSQL := `DELETE FROM cache_ifname WHERE time < now() - INTERVAL ? MINUTE;`
	cleanUpIfAliasSQL := `DELETE FROM cache_ifalias WHERE time < now() - INTERVAL ? MINUTE;`

	if _, err := c.db.Exec(cleanUpHostnameSQL, cacheHostnameMinutes); err != nil {
		log.Println(err)
	}

	if _, err := c.db.Exec(cleanUpIfNameSQL, cacheIfNameMinutes); err != nil {
		log.Println(err)
	}

	if _, err := c.db.Exec(cleanUpIfAliasSQL, cacheIfAliasMinutes); err != nil {
		log.Println(err)
	}
}
