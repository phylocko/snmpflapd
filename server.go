package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"
	g "github.com/gosnmp/gosnmp"
)

const (
	defaultConfigFilename  = "settings.conf"
	defaultLogFilename     = "snmpflapd.log"
	defaultListenAddress   = "0.0.0.0"
	defaultListenPort      = 162
	defaultDBHost          = "127.0.0.1"
	defaultDBUser          = "root"
	defaultDBName          = "snmpflapd"
	defaultDBPassword      = ""
	defaultCommunity       = ""
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
	CleanUpInterval int
}

// flags
var (
	version            string
	build              string
	flagVerbose        bool
	flagConfigFilename string
	flagVersion        bool
)

var config = Config{
	LogFilename:     defaultLogFilename,
	ListenAddress:   defaultListenAddress,
	ListenPort:      defaultListenPort,
	DBHost:          defaultDBHost,
	DBName:          defaultDBName,
	DBUser:          defaultDBUser,
	DBPassword:      defaultDBPassword,
	Community:       defaultCommunity,
	CleanUpInterval: defaultCleanUpInterval,
}

func init() {

	// Reading flags
	flag.StringVar(&flagConfigFilename, "f", defaultConfigFilename, "Location of config file")
	flag.BoolVar(&flagVerbose, "v", false, "Enable verbose logging")
	flag.BoolVar(&flagVersion, "V", false, "Print version information and quit")
	flag.Parse()

	// Reading config
	readConfigFile(&flagConfigFilename)
	readConfigEnv()

}

func main() {

	if flagVersion {
		build := fmt.Sprintf("FlapMyPort snmpflapd version %s, build %s", version, build)
		fmt.Println(build)
		os.Exit(0)
	}

	var err error

	// Logging setup
	f, err := os.OpenFile(config.LogFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
		log.Fatalln(err)
	}
	defer f.Close()
	log.SetOutput(f)
	log.Println("snmpflapd started")

	connector, err = MakeDB(config.DBHost, config.DBName, config.DBUser, config.DBPassword)
	if err != nil {
		fmt.Println(err)
		log.Fatalln(err)
	}
	defer connector.db.Close()

	snmpSema = RequestSemaphore{}

	// Periodic DB clean up
	go RunDBCleanUp()

	tl := g.NewTrapListener()
	tl.OnNewTrap = handleTrap
	tl.Params = g.Default

	listenSocket := fmt.Sprintf("%v:%v", config.ListenAddress, config.ListenPort)
	tlErr := tl.Listen(listenSocket)
	if tlErr != nil {
		fmt.Println(tlErr)
		log.Fatalln(tlErr)
	}

}

func readConfigFile(file *string) {
	if _, err := toml.DecodeFile(*file, &config); err != nil {
		msg := fmt.Sprintf("%s not found. Suppose we're using environment variables", *file)
		fmt.Println(msg)
		log.Println(msg)
	}
}

func readConfigEnv() {

	if logFilename, exists := os.LookupEnv("LOGFILE"); exists {
		config.LogFilename = logFilename
	}

	if listenAddress, exists := os.LookupEnv("LISTEN_ADDRESS"); exists {
		config.ListenAddress = listenAddress
	}

	if listenPort, exists := os.LookupEnv("LISTEN_PORT"); exists {
		if intPort, error := strconv.Atoi(listenPort); error != nil {
			msg := "Wrong environment variable LISTEN_PORT"
			fmt.Println(msg)
			log.Fatalln(msg)

		} else {
			config.ListenPort = intPort
		}

	}

	if dbHost, exists := os.LookupEnv("DBHOST"); exists {
		config.DBHost = dbHost
	}

	if dbName, exists := os.LookupEnv("DBNAME"); exists {
		config.DBName = dbName
	}

	if dbUser, exists := os.LookupEnv("DBUSER"); exists {
		config.DBUser = dbUser
	}

	if dbPassword, exists := os.LookupEnv("DBPASSWORD"); exists {
		config.DBPassword = dbPassword
	}

	if community, exists := os.LookupEnv("COMMUNITY"); exists {
		config.Community = community
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
