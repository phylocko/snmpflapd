package flap

import (
	"fmt"
	"net"
	"snmpflapd/internal/cache"
	"snmpflapd/internal/db"
	"snmpflapd/internal/logger"
	"snmpflapd/internal/retriever"
	"strings"
	"time"

	"github.com/chilts/sid"
	snmp "github.com/gosnmp/gosnmp"
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
	dateLayout               = "2006-01-02 15:04:05"
)

type Flap struct {
	sid           string
	ifIndex       int
	ifAdminStatus int
	ifOperStatus  int
	ifName        *string
	ifAlias       *string
	hostName      *string
	ipAddress     string
	time          time.Time
	timeTicks     uint
}

// New creates a new Flap from an SNMP packet
// You have to provide an IP address
// as SNMP packet doesn't contain it
func New(p *snmp.SnmpPacket, addr net.IP) (flap *Flap) {
	flap = &Flap{
		sid:       sid.Id(),
		ipAddress: addr.String(),
	}

	for _, variable := range p.Variables {

		if strings.Contains(variable.Name, ifIndexOIDPrefix) {
			flap.ifIndex = variable.Value.(int)
			continue
		}

		if strings.Contains(variable.Name, ifAdminStatusOIDPrefix) {
			flap.ifAdminStatus = variable.Value.(int)
			continue
		}

		if strings.Contains(variable.Name, ifOperStatusOIDPrefix) {
			flap.ifOperStatus = variable.Value.(int)
			continue
		}

		if strings.Contains(variable.Name, ifNameVarBindPrefixJunOS) {
			ifNameBytes, ok := variable.Value.([]uint8)
			if ok {
				ifName := string(ifNameBytes)
				flap.ifName = &ifName
			}
			continue
		}

		if strings.Contains(variable.Name, timeTicksReference) {
			timeTicks, ok := variable.Value.(uint)
			if ok {
				flap.timeTicks = timeTicks
			}
			continue
		}

	}

	return
}

// GetIfStatus returns string representation of an integer SNMP status code
func GetIfStatus(status int) string {
	if status == ifAdminStatusUP {
		return "up"
	}
	return "down"
}

// GetIfStatus returns string representation of a time for SQL insertion
func GetDate(time time.Time) string {
	return time.Format(dateLayout)
}

func (f *Flap) Save() error {

	q := `INSERT INTO ports 
	(ipaddress, hostname, ifIndex, ifName, ifAlias, ifAdminStatus, ifOperStatus, time, sid, timeTicks)
	VALUES 
	(:ipaddress, :hostname, :ifIndex, :ifName, :ifAlias, :ifAdminStatus, :ifOperStatus, :time, :sid, :timeTicks)`

	args := map[string]interface{}{
		"ipaddress":     f.ipAddress,
		"hostname":      f.hostName,
		"ifIndex":       f.ifIndex,
		"ifName":        f.ifName,
		"ifAlias":       f.ifAlias,
		"ifAdminStatus": GetIfStatus(f.ifAdminStatus),
		"ifOperStatus":  GetIfStatus(f.ifOperStatus),
		"time":          GetDate(f.time),
		"sid":           f.sid,
		"timeTicks":     f.timeTicks}

	_, err := db.NamedExec(q, args)
	return err
}

func (f *Flap) Update() error {

	q := `UPDATE ports 
	SET hostname = :hostname, ifName = :ifName, ifAlias = :ifAlias 
	WHERE sid = :sid;`

	args := map[string]interface{}{
		"hostname": f.hostName,
		"ifAlias":  f.ifAlias,
		"ifName":   f.ifName,
		"sid":      f.sid}

	_, err := db.NamedExec(q, args)
	return err
}

func (f *Flap) FetchMissingData() {
	if f.hostName == nil {
		f.FillHostname()
	}

	if f.ifName == nil {
		f.FillIfName()
	}

	if f.ifAlias == nil {
		f.FillIfAlias()
	}
}

// FillHostName tries to get a hostname from cache, then from the device via SNMP request
func (f *Flap) FillHostname() {

	f.hostName = cache.GetHostname(f.ipAddress)

	if f.hostName != nil {
		return
	}

	// 2. Get value from SNMP and put it to the cache
	if value, err := retriever.GetString(sysNameOID, f.ipAddress); err != nil {
		logger.L.Printf("unable to get hostname via snmp from %s. %s", f.ipAddress, err)
		return

	} else {
		f.hostName = value
	}

	cache.PutHostname(f.ipAddress, *f.hostName)
}

// FillIfName tries to get a interface name from cache, then from the device via SNMP request
func (f *Flap) FillIfName() {

	f.ifName = cache.GetIfName(f.ipAddress, f.ifIndex)

	if f.ifName != nil {
		return
	}

	// 2. Get value from SNMP and put it to the cache
	oid := fmt.Sprintf("%s%d", ifNameOIDPrefix, f.ifIndex)
	if value, err := retriever.GetString(oid, f.ipAddress); err != nil {
		logger.L.Printf("unable to get ifalias via snmp from %s for %d. %s", f.ipAddress, f.ifIndex, err)
		return

	} else {
		f.ifName = value
	}

	cache.PutIfName(f.ipAddress, f.ifIndex, *f.ifName)
}

// FillIfAlias tries to get a interface name from cache, then from the device via SNMP request
func (f *Flap) FillIfAlias() {

	f.ifAlias = cache.GetIfAlias(f.ipAddress, f.ifIndex)

	if f.ifAlias != nil {
		return
	}

	// 2. Get value from SNMP and put it to the cache
	oid := fmt.Sprintf("%s%d", ifAliasOIDPrefix, f.ifIndex)
	if value, err := retriever.GetString(oid, f.ipAddress); err != nil {
		logger.L.Printf("unable to get ifalias via snmp from %s for %d. %s", f.ipAddress, f.ifIndex, err)
		return

	} else {
		f.ifAlias = value
	}

	cache.PutIfAlias(f.ipAddress, f.ifIndex, *f.ifName)
}
