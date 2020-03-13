package main

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	g "github.com/soniah/gosnmp"
	"log"
	"net"
	"os"
	"time"
)

const (
	defaultConfigFilename  = "settings.conf"
	defaultLogFilename     = "snmpflapd.log"
	defaultListenAddress   = "0.0.0.0"
	defaultListenPort      = 162
	defaultDBUser          = "root"
	defaultDBName          = "snmpflapd"
	defaultDBPassword      = ""
	defaultCommunity       = ""
	defaultSendMail        = false
	queueInterval          = 30
	defaultCleanUpInterval = 60
)

type Config struct {
	LogFilename     string
	ListenAddress   string
	ListenPort      int
	DBHost          string
	DBName          string
	DBUser          string
	DBPassword      string
	Community       string
	SendMail        bool
	MailList        []string
	CleanUpInterval int
}

// flags
var (
	flagVerbose        bool
	flagConfigFilename string
)

var config = Config{
	LogFilename:     defaultLogFilename,
	ListenAddress:   defaultListenAddress,
	ListenPort:      defaultListenPort,
	DBName:          defaultDBName,
	DBUser:          defaultDBUser,
	DBPassword:      defaultDBPassword,
	Community:       defaultCommunity,
	SendMail:        defaultSendMail,
	CleanUpInterval: defaultCleanUpInterval}

func init() {

	// Reading flags
	flag.StringVar(&flagConfigFilename, "f", defaultConfigFilename, "config file")
	flag.BoolVar(&flagVerbose, "v", false, "verbose logging")
	flag.Parse()

	// Reading config
	readConfig(&flagConfigFilename)

}

func main() {
	var err error

	// Logging setup
	f, err := os.OpenFile(config.LogFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	log.SetOutput(f)
	log.Println("snmpflapd started")

	connector, err = MakeDB(config.DBName, config.DBUser, config.DBPassword)
	if err != nil {
		log.Fatalln(err)
	}
	defer connector.db.Close()

	snmpSema = RequestSemaphore{}

	// Notify queue
	mailQueue = Queue{}
	go RunQueue()

	// Periodic DB clean up
	go RunDBCleanUp()

	tl := g.NewTrapListener()
	tl.OnNewTrap = handleTrap
	tl.Params = g.Default

	listenSocket := fmt.Sprintf("%v:%v", config.ListenAddress, config.ListenPort)
	tlErr := tl.Listen(listenSocket)
	if tlErr != nil {
		log.Fatalln(tlErr)
	}

}

func readConfig(file *string) {
	if _, err := toml.DecodeFile(*file, &config); err != nil {
		log.Fatalln(err)
	}
}

func handleTrap(packet *g.SnmpPacket, addr *net.UDPAddr) {
	go asyncHandleTrap(packet, addr)
}

func asyncHandleTrap(packet *g.SnmpPacket, addr *net.UDPAddr) {

	if isLinkEvent(packet) {
		LinkEventHandler(packet, addr)
	}
}

func RunDBCleanUp() {
	for {
		time.Sleep(time.Hour * 6)
		connector.CleanUp()
	}
}

func logVerbose(s string) {
	if flagVerbose {
		log.Print(s)
	}
}
