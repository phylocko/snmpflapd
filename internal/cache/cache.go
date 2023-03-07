package cache

import (
	"snmpflapd/internal/db"

	"github.com/apex/log"
)

const (
	cacheHostnameMinutes = 360
	cacheIfNameMinutes   = 180
	cacheIfAliasMinutes  = 60
)

func CleanUp() {

	log.Info("DB Cleanup started")

	cleanUpHostnameSQL := `DELETE FROM cache_hostname WHERE time < now() - INTERVAL ? MINUTE;`
	if _, err := db.Exec(cleanUpHostnameSQL, cacheHostnameMinutes); err != nil {
		log.WithError(err).Error("Unable to execute cleanUpHostnameSQL.")
	}

	cleanUpIfNameSQL := `DELETE FROM cache_ifname WHERE time < now() - INTERVAL ? MINUTE;`
	if _, err := db.Exec(cleanUpIfNameSQL, cacheIfNameMinutes); err != nil {
		log.WithError(err).Error("Unable to execute cleanUpIfNameSQL.")
	}

	cleanUpIfAliasSQL := `DELETE FROM cache_ifalias WHERE time < now() - INTERVAL ? MINUTE;`
	if _, err := db.Exec(cleanUpIfAliasSQL, cacheIfAliasMinutes); err != nil {
		log.WithError(err).Error("Unable to execute cleanUpIfAliasSQL.")
	}
}

// Hostname

func GetHostname(ip string) *string {

	sql := `SELECT hostname 
	FROM cache_hostname 
	WHERE time > now() - INTERVAL ? MINUTE 
	AND ipaddress = ?;`

	var value string
	if err := db.Get(&value, sql, cacheHostnameMinutes, ip); err != nil {
		return nil
	}

	return &value
}

func PutHostname(ip, hostname string) error {

	q := `DELETE FROM cache_hostname WHERE ipaddress = ?;`
	if _, err := db.Exec(q, ip); err != nil {
		return err
	}

	q = `INSERT INTO cache_hostname (ipaddress, hostname) VALUES (?, ?);`
	if _, err := db.Exec(q, ip, hostname); err != nil {
		return err
	}

	return nil
}

// Ifname

func GetIfName(ip string, ifindex int) *string {

	sql := `SELECT ifName 
	FROM cache_ifname
	WHERE time > now() - INTERVAL ? MINUTE 
	AND ipaddress = ? 
	AND ifIndex = ?;`

	var value string
	if err := db.Get(&value, sql, cacheIfNameMinutes, ip); err != nil {
		return nil
	}

	return &value
}

func PutIfName(ip string, ifindex int, ifname string) error {

	q := `DELETE FROM cache_ifname WHERE ipaddress = ? and ifIndex = ?;`

	if _, err := db.Exec(q, ip); err != nil {
		return err
	}

	q = `INSERT INTO cache_ifname (ipaddress, ifIndex, ifName) VALUES (?, ?, ?);`

	if _, err := db.Exec(q, ip, ifindex, ifname); err != nil {
		return err
	}

	return nil
}

// Ifalias

func GetIfAlias(ip string, ifindex int) *string {

	sql := `SELECT ifAlias 
	FROM cache_ifalias 
	WHERE time > now() - INTERVAL ? MINUTE 
	AND ipaddress = ? 
	AND ifIndex = ?;`

	var value string
	if err := db.Get(&value, sql, cacheIfNameMinutes, ip); err != nil {
		return nil
	}

	return &value
}

func PutIfAlias(ip string, ifindex int, ifalias string) error {

	q := `DELETE FROM cache_ifalias WHERE ipaddress = ? and ifindex = ?;`

	if _, err := db.Exec(q, ip); err != nil {
		return err
	}

	q = `INSERT INTO cache_ifalias (ipaddress, ifIndex, ifAlias) VALUES (?, ?, ?);`

	if _, err := db.Exec(q, ip, ifindex, ifalias); err != nil {
		return err
	}

	return nil
}
