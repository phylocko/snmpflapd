package retriever

import (
	"errors"
	"fmt"
	"snmpflapd/internal/config"
	"sync"

	snmp "github.com/gosnmp/gosnmp"
)

// Mutex lock to limit network connections
var mx sync.Mutex

func DoRequest(oid string, ip string) (*snmp.SnmpPacket, error) {

	c := snmp.Default
	c.Community = config.Config.Community
	c.Target = ip

	if err := c.Connect(); err != nil {
		return nil, err
	}
	defer c.Conn.Close()

	return snmp.Default.Get([]string{oid})
}

func GetString(oid string, ip string) (*string, error) {

	mx.Lock()
	defer mx.Unlock()

	packet, err := DoRequest(oid, ip)
	if err != nil {
		return nil, err
	}

	value := packet.Variables[0].Value
	valueBytes, ok := value.([]byte)
	if ok {
		valueString := string(valueBytes)
		return &valueString, nil
	}

	return nil, errors.New(fmt.Sprintf("device %s returned a wrong value", ip))
}
