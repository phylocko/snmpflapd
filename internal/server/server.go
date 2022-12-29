package server

import (
	"net"

	"snmpflapd/internal/flap"
	"snmpflapd/internal/logger"

	snmp "github.com/gosnmp/gosnmp"
)

func New() (listener *snmp.TrapListener) {
	listener = snmp.NewTrapListener()
	listener.Params = snmp.Default
	listener.OnNewTrap = HandleFlap
	return listener

}

func HandleFlap(p *snmp.SnmpPacket, addr *net.UDPAddr) {

	f := flap.New(p, addr.IP)

	if err := f.Save(); err != nil {
		logger.L.Printf("unable to save flap: %s", err)
		return
	}

	f.FetchMissingData()
	if err := f.Update(); err != nil {
		logger.L.Printf("unable to update flap: %s", err)
		return
	}
}
