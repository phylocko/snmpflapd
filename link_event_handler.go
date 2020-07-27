// This file is responsible for handling linkUP/linkDown snmp traps.
// It performed the following actions:
// - creates a linkEvent from an snmpPacket
// - fills it's fields with additional info, missing in the snmpPacket
// - reads/writes to cache DB tables using the *Connector
// - finally, stores linkEvents to a database using the *Connector

package main

import (
	"fmt"
	"github.com/chilts/sid"
	g "github.com/soniah/gosnmp"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	oidReference             = ".1.3.6.1.6.3.1.1.4.1.0"
	timeTicksReference       = ".1.3.6.1.2.1.1.3.0"
	linkUP                   = ".1.3.6.1.6.3.1.1.5.4"
	linkDOWN                 = ".1.3.6.1.6.3.1.1.5.3"
	ifIndexOIDPrefix         = ".1.3.6.1.2.1.2.2.1.1"
	ifNameOIDPrefix          = ".1.3.6.1.2.1.31.1.1.1.1."
	ifAliasOIDPrefix         = ".1.3.6.1.2.1.31.1.1.1.18."
	ifAdminStatusOIDPrefix   = ".1.3.6.1.2.1.2.2.1.7"
	ifOperStatusOIDPrefix    = ".1.3.6.1.2.1.2.2.1.8"
	ifNameVarBindPrefixJunOS = ".1.3.6.1.2.1.31.1.1.1.1"
	sysNameOID               = ".1.3.6.1.2.1.1.5.0"
	ifAdminStatusUP          = 1
	ifAdminStatusDOWN        = 2
	ifOperStatusUP           = 1
	ifOperStatusDOWN         = 2
	cacheHostnameMinutes     = 60
	cacheIfNameMinutes       = 30
	cacheIfAliasMinutes      = 5
	dateLayout               = "2006-01-02 15:04:05"
)

var (
	snmpSema RequestSemaphore
)

type linkEvent struct {
	sid           string
	ifIndex       int
	ifAdminStatus int
	ifOperStatus  int
	ifName        *string
	ifAlias       *string
	hostName      *string
	ipAddress     net.IP
	time          time.Time
	timeTicks     uint
}

func (le *linkEvent) String() string {
	eventTime := le.time.Format(dateLayout)

	hostName := le.ipAddress.String()
	if le.hostName != nil {
		hostName = *le.hostName
	}

	ifName := "NULL"
	if le.ifName != nil {
		ifName = *le.ifName
	}
	ifAlias := "NULL"
	if le.ifAlias != nil {
		ifAlias = *le.ifAlias
	}

	return fmt.Sprintf("eventTime=%s host=%s ifName=%s ifIndex=%d ifAlias=%s status=%s",
		eventTime, hostName, ifName, le.ifIndex, ifAlias, le.ifStateText())
}

func (le *linkEvent) ifStateText() string {
	var ifState string

	switch le.ifAdminStatus {
	case ifAdminStatusDOWN:
		ifState = "admin down"

	case ifAdminStatusUP:
		switch le.ifOperStatus {
		case ifOperStatusUP:
			ifState = "up"
		default:
			ifState = "down"
		}
	}
	return ifState
}

// FromSnmpPacket returns linkEvent from SnmpPacket and net.UDPAddr
func (le *linkEvent) FromSnmpPacket(p *g.SnmpPacket, addr net.IP) {
	if !isLinkEvent(p) {
		// Don't waste my CPU time!
		return
	}

	le.ipAddress = addr

	// Fill the linkEvent with variables from a packet
	for _, variable := range p.Variables {

		if strings.Contains(variable.Name, ifIndexOIDPrefix) {
			le.ifIndex = variable.Value.(int)
			continue
		}

		if strings.Contains(variable.Name, ifAdminStatusOIDPrefix) {
			le.ifAdminStatus = variable.Value.(int)
			continue
		}

		if strings.Contains(variable.Name, ifOperStatusOIDPrefix) {
			le.ifOperStatus = variable.Value.(int)
			continue
		}

		if strings.Contains(variable.Name, ifNameVarBindPrefixJunOS) {
			ifNameBytes, ok := variable.Value.([]uint8)
			ifName := string(ifNameBytes)
			if ok {
				le.ifName = &ifName
			} else {
				log.Println(le, "empty ifNameVarBindPrefixJunOS")
			}
			continue
		}

		if strings.Contains(variable.Name, timeTicksReference) {

			timeTicks, ok := variable.Value.(uint)

			if ok {
				le.timeTicks = timeTicks
			} else {
				log.Println(le, "missing timeTicks in the SNMP trap")
			}
			continue
		}

	}
}

// LinkEventHandler handles linkUP/linkDOWN snmp traps
func LinkEventHandler(p *g.SnmpPacket, addr *net.UDPAddr) {
	event := linkEvent{time: time.Now().Local()}
	event.sid = sid.Id() // This is for unique trap identification
	event.FromSnmpPacket(p, addr.IP)

	logVerbose(fmt.Sprintln(event.sid, "trap received:", event.String()))

	if err := connector.SaveLinkEvent(&event); err != nil {
		log.Println(event.sid, "unable to save link event", err)
		return
	}

	// Fetch missing data and update the linkEvent
	event.FetchMissingData()
	if err := connector.UpdateLinkEvent(&event); err != nil {
		log.Println(event.sid, "unable to update link event:", err)
	}

}

// getEventOID returns oid from OID Reference that is in an SnmpPacket
func getEventOID(p *g.SnmpPacket) string {
	for _, variable := range p.Variables {
		if variable.Name == oidReference {
			return variable.Value.(string)
		}
	}
	return ""
}

// isLinkEvent returns true if an SNMP trap is about Link UP/DOWN event
func isLinkEvent(p *g.SnmpPacket) bool {
	eventOid := getEventOID(p)
	if eventOid == linkUP || eventOid == linkDOWN {
		return true
	}
	return false
}

func (le *linkEvent) FetchMissingData() {

	logVerbose(fmt.Sprintln(le.sid, "fetching missing data"))

	if le.hostName == nil {
		le.FillHostName()
	}

	if le.ifName == nil {
		le.FillIfName()
	}

	if le.ifAlias == nil {
		le.FillIfAlias()
	}
}

// FillHostName tries to get a hostname from cache, then from the device via SNMP request
func (le *linkEvent) FillHostName() {

	logVerbose(fmt.Sprintln(le.sid, "filling hostname"))

	// 1. Try to get the value from cache
	if connector.getCachedHostname(le) {
		logVerbose(fmt.Sprintln(le.sid, "used cached hostName", *le.hostName))
		return
	}

	// 2. Get value from SNMP and put it to the cache
	if hostName, err := getSNMPString(sysNameOID, le.ipAddress); err != nil {
		log.Println(le.sid, "unable to get hostname via SNMP:", err)
		return

	} else {
		le.hostName = hostName
		logVerbose(fmt.Sprintf("%s received hostname '%s' from %s via SNMP", le.sid, *le.hostName, le.ipAddress))
	}

	connector.putCachedHostname(le)

}

// FillHostName tries to get a ifName from cache, then from the device via SNMP request
func (le *linkEvent) FillIfName() {

	logVerbose(fmt.Sprintln(le.sid, "filling ifName"))

	// 1. Try to get the value from cache
	if connector.getCachedIfName(le) {
		logVerbose(fmt.Sprintf("%s used cached ifName %s", le.sid, *le.ifName))
		return
	}

	// 2. Get value from SNMP and put it to the cache
	if ifName, err := getSNMPString(ifNameOIDPrefix+strconv.Itoa(le.ifIndex), le.ipAddress); err != nil {
		log.Println(le.sid, "unable to get ifName vie SNMP:", err)
		return

	} else {
		le.ifName = ifName
		logVerbose(fmt.Sprintf("%s received ifName '%s' from %s via SNMP", le.sid, *le.ifName, le.ipAddress))
	}

	connector.putCachedIfName(le)

}

// FillIfAlias tries to get an ifAlias from cache, then from the device via SNMP request
func (le *linkEvent) FillIfAlias() {

	logVerbose(fmt.Sprintln(le.sid, "filling ifAlias"))

	// 1. Try to get the value from cache
	if connector.getCachedIfAlias(le) {
		logVerbose(fmt.Sprintf("%s used cached ifAlias %s", le.sid, *le.ifAlias))
		return
	}

	// 2. Get value from SNMP and put it to the cache
	ifAlias, err := getSNMPString(ifAliasOIDPrefix+strconv.Itoa(le.ifIndex), le.ipAddress)
	if err != nil {
		log.Println(le.sid, "unable to get ifAlias via SNMP:", err)
		return

	} else {
		le.ifAlias = ifAlias
		logVerbose(fmt.Sprintf("%s received ifAlias '%s' from %s via SNMP", le.sid, *ifAlias, &le.ipAddress))
	}

	connector.putCachedIfAlias(le)
}

func (c *Connector) SaveLinkEvent(le *linkEvent) error {

	if le.timeTicks == 0 {
		log.Println("SNMP Trap has no timeTicks", le)
	}

	ifAdminStatus, ifOperStatus := "down", "down"
	if le.ifAdminStatus == ifAdminStatusUP {
		ifAdminStatus = "up"
	}

	if le.ifOperStatus == ifOperStatusUP {
		ifOperStatus = "up"
	}

	sql := `INSERT INTO ports 
			(ipaddress, hostname, ifIndex, ifName, ifAlias, ifAdminStatus, ifOperStatus, time, sid, timeTicks)
			VALUES 
			(:ipaddress, :hostname, :ifIndex, :ifName, :ifAlias, :ifAdminStatus, :ifOperStatus, :time, :sid, :timeTicks)`

	args := map[string]interface{}{
		"ipaddress":     le.ipAddress.String(),
		"hostname":      le.hostName,
		"ifIndex":       le.ifIndex,
		"ifName":        le.ifName,
		"ifAlias":       le.ifAlias,
		"ifAdminStatus": ifAdminStatus,
		"ifOperStatus":  ifOperStatus,
		"time":          le.time.Format(dateLayout),
		"sid":           le.sid,
		"timeTicks":     le.timeTicks}

	c.mx.Lock()
	defer c.mx.Unlock()

	if _, err := c.db.NamedExec(sql, args); err != nil {
		log.Println(le.sid, "unable to exec SQL query", err)
		return err
	}

	logVerbose(fmt.Sprintln(le.sid, "link event saved", le.String()))
	return nil
}

func (c *Connector) UpdateLinkEvent(le *linkEvent) error {

	sql := `UPDATE ports SET  hostname = :hostname, ifName = :ifName, ifAlias = :ifAlias WHERE sid = :sid;`

	args := map[string]interface{}{
		"hostname": le.hostName,
		"ifAlias":  le.ifAlias,
		"ifName":   le.ifName,
		"sid":      le.sid}

	c.mx.Lock()
	defer c.mx.Unlock()

	if _, err := c.db.NamedExec(sql, args); err != nil {
		log.Println(le.sid, "unable to exec SQL query", err)
		return err
	}

	logVerbose(fmt.Sprintln(le.sid, "link event updated", le.String()))
	return nil
}

func (c *Connector) getCachedIfName(le *linkEvent) bool {

	c.mx.Lock()
	defer c.mx.Unlock()

	sql := "SELECT ifName FROM cache_ifname	WHERE time > now() - INTERVAL ? MINUTE AND ipaddress = ? AND ifIndex = ?;"

	cachedIfName := ""
	if err := c.db.Get(&cachedIfName, sql, cacheIfNameMinutes, le.ipAddress.String(), le.ifIndex); err != nil {
		logVerbose(fmt.Sprintln(le.sid, "no cached ifName"))
		return false
	}

	le.ifName = &cachedIfName
	return true
}

func (c *Connector) putCachedIfName(le *linkEvent) {

	c.mx.Lock()
	defer c.mx.Unlock()

	var sql string

	sql = `DELETE FROM cache_ifname WHERE ipaddress = ? and ifIndex = ?;`
	if _, err := c.db.Exec(sql, le.ipAddress.String(), le.ifIndex); err != nil {
		log.Println(le.sid, err)
		return
	}

	sql = `INSERT INTO cache_ifname (ipaddress, ifIndex, ifName) VALUES (?, ?, ?);`
	if _, err := c.db.Exec(sql, le.ipAddress.String(), le.ifIndex, le.ifName); err != nil {
		log.Println(le.sid, err, le.String())
		return
	}
	logVerbose(fmt.Sprintf("%s put values ('%s', '%d', '%d') to cache_ifname", le.sid, *le.ifName, le.ifIndex, le.hostName))
}

func (c *Connector) getCachedIfAlias(le *linkEvent) bool {

	c.mx.Lock()
	defer c.mx.Unlock()

	sql := "SELECT ifAlias FROM cache_ifalias WHERE time > now() - INTERVAL ? MINUTE AND ipaddress = ? AND ifIndex = ?;"

	cachedIfAlias := ""
	if err := c.db.Get(&cachedIfAlias, sql, cacheIfAliasMinutes, le.ipAddress.String(), le.ifIndex); err != nil {
		logVerbose(fmt.Sprintln(le.sid, "no cached ifAlias"))
		return false
	}

	le.ifAlias = &cachedIfAlias
	return true
}

func (c *Connector) putCachedIfAlias(le *linkEvent) {

	c.mx.Lock()
	defer c.mx.Unlock()

	var sql string

	sql = `DELETE FROM cache_ifalias WHERE ipaddress = ? and ifindex = ?;`
	if _, err := c.db.Exec(sql, le.ipAddress.String(), le.ifIndex); err != nil {
		log.Println(le.sid, err)
		return
	}

	sql = `INSERT INTO cache_ifalias (ipaddress, ifIndex, ifAlias) VALUES (?, ?, ?);`
	if _, err := c.db.Exec(sql, le.ipAddress.String(), le.ifIndex, le.ifAlias); err != nil {
		log.Println(le.sid, err)
		return
	}
	logVerbose(fmt.Sprintf("%s put values ('%s', '%d', '%s') to cache_ifalias", le.sid, *le.ifAlias, le.ifIndex, le.ipAddress))

}

func (c *Connector) getCachedHostname(le *linkEvent) bool {

	c.mx.Lock()
	defer c.mx.Unlock()

	sql := "SELECT hostname FROM cache_hostname WHERE time > now() - INTERVAL ? MINUTE AND ipaddress = ?;"

	var cachedHostname string
	if err := c.db.Get(&cachedHostname, sql, cacheHostnameMinutes, le.ipAddress.String()); err != nil {
		logVerbose(fmt.Sprintln(le.sid, "no cached hostname"))
		return false
	}

	le.hostName = &cachedHostname
	return true
}

func (c *Connector) putCachedHostname(le *linkEvent) {

	c.mx.Lock()
	defer c.mx.Unlock()

	var sql string
	sql = `DELETE FROM cache_hostname WHERE ipaddress = ?;`
	if _, err := c.db.Exec(sql, le.ipAddress.String()); err != nil {
		log.Println(le.sid, err)
		return
	}

	sql = `INSERT INTO cache_hostname (ipaddress, hostname) VALUES (?, ?);`
	if _, err := c.db.Exec(sql, le.ipAddress.String(), le.hostName); err != nil {
		log.Println(err)
		return
	}
	logVerbose(fmt.Sprintf("%s put values ('%s', '%s') to cache_hostname", le.sid, *le.hostName, le.ipAddress))
}
