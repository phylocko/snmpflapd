package main

import (
	"errors"
	g "github.com/soniah/gosnmp"
	"log"
	"net"
	"sync"
)

type RequestSemaphore struct {
	requestQueue []linkEvent
	mx           sync.Mutex
}

func doSNMPRequest(oid string, ip net.IP) (pdu *g.SnmpPacket, err error) {

	c := g.Default
	c.Community = config.Community
	c.Target = ip.String()


	if err = c.Connect(); err != nil {
		log.Println(err)
		return nil, err
	}
	defer c.Conn.Close()

	return g.Default.Get([]string{oid})
}

func getSNMPString(oid string, ip net.IP) (val *string, err error) {


	snmpSema.mx.Lock()
	defer snmpSema.mx.Unlock()

	pdu, err := doSNMPRequest(oid, ip)
	if err != nil {
		return nil, err
	}
	value := pdu.Variables[0].Value
	fromByte, ok := value.([]byte)
	if ok {
		s := string(fromByte)
		return &s, nil
	}
	return nil, errors.New("received nil from the device")
}
