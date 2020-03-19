package main

import (
	"log"
	"net"
)

type lastEvent struct {
	Ipaddress string `db:"ipaddress"`
	IfIndex   int    `db:"ifIndex"`
}

func (c *Connector) fillCache() {
	c.mx.Lock()

	filtered := []lastEvent{}
	sql := "SELECT DISTINCT a.ipaddress, a.ifIndex FROM (SELECT * FROM ports WHERE time > now() - INTERVAL 60 MINUTE AND (ifAlias='' OR hostname='')) AS a;"
	if err := c.db.Select(&filtered, sql); err != nil {
		log.Println(err)
	}
	c.mx.Unlock()

	for _, tmp := range filtered {
		event := &linkEvent{
			ifIndex:   tmp.IfIndex,
			ipAddress: net.ParseIP(tmp.Ipaddress).To4(),
		}
		event.FetchMissingData()

		sql = `UPDATE ports SET  hostname = :hostname, ifName = :ifName, ifAlias = :ifAlias WHERE ipaddress = :ipaddress AND ifIndex = :ifIndex AND time > now() - INTERVAL 60 MINUTE;`
		args := map[string]interface{}{
			"hostname":  event.hostName,
			"ifAlias":   event.ifAlias,
			"ifName":    event.FillIfName,
			"ipaddress": event.ipAddress,
			"ifIndex":   event.ifIndex,
		}

		c.mx.Lock()

		if _, err := c.db.NamedExec(sql, args); err != nil {
			log.Println(err)
		}

		c.mx.Unlock()
	}
}
