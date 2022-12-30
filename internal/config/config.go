package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/BurntSushi/toml"
)

const (
	defaultConfigFilename  = "settings.conf"
	defaultLogFilename     = "snmpflapd.log"
	defaultLogLevel        = "warning"
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

var (
	FlagVerbose        bool
	FlagVersion        bool
	FlagConfigFilename string
)

type config struct {
	LogFilename     string
	LogLevel        string
	ListenAddress   string
	ListenPort      int
	DBHost          string
	DBName          string
	DBUser          string
	DBPassword      string
	Community       string
	CleanUpInterval int
}

var Config config

func ReadFlags() {
	flag.StringVar(&FlagConfigFilename, "f", defaultConfigFilename, "Location of config file")
	flag.BoolVar(&FlagVerbose, "v", false, "Enable verbose logging")
	flag.BoolVar(&FlagVersion, "V", false, "Print version information and quit")
	flag.Parse()
}

func ReadConfig() {
	cfg := config{
		LogFilename:     defaultLogFilename,
		LogLevel:        defaultLogLevel,
		ListenAddress:   defaultListenAddress,
		ListenPort:      defaultListenPort,
		DBHost:          defaultDBHost,
		DBName:          defaultDBName,
		DBUser:          defaultDBUser,
		DBPassword:      defaultDBPassword,
		Community:       defaultCommunity,
		CleanUpInterval: defaultCleanUpInterval,
	}

	ReadFile(FlagConfigFilename, &cfg)
	ReadEnv(&cfg)

	Config = cfg

}

func ReadFile(fileName string, cfg *config) {
	if _, err := toml.DecodeFile(fileName, &cfg); err != nil {
		msg := fmt.Sprintf("%s not found. Suppose we're using environment variables", fileName)
		fmt.Println(msg)
	}
}

func ReadEnv(cfg *config) {

	if logFilename, exists := os.LookupEnv("LOGFILE"); exists {
		cfg.LogFilename = logFilename
	}

	if logLevel, exists := os.LookupEnv("LOGLEVEL"); exists {
		cfg.LogLevel = logLevel
	}

	if listenAddress, exists := os.LookupEnv("LISTEN_ADDRESS"); exists {
		cfg.ListenAddress = listenAddress
	}

	if listenPort, exists := os.LookupEnv("LISTEN_PORT"); exists {
		if intPort, error := strconv.Atoi(listenPort); error != nil {
			msg := "Wrong environment variable LISTEN_PORT"
			fmt.Println(msg)

		} else {
			cfg.ListenPort = intPort
		}

	}

	if dbHost, exists := os.LookupEnv("DBHOST"); exists {
		cfg.DBHost = dbHost
	}

	if dbName, exists := os.LookupEnv("DBNAME"); exists {
		cfg.DBName = dbName
	}

	if dbUser, exists := os.LookupEnv("DBUSER"); exists {
		cfg.DBUser = dbUser
	}

	if dbPassword, exists := os.LookupEnv("DBPASSWORD"); exists {
		cfg.DBPassword = dbPassword
	}

	if community, exists := os.LookupEnv("COMMUNITY"); exists {
		cfg.Community = community
	}
}
