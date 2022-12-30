package main

import (
	"fmt"
	"os"
	"time"

	"snmpflapd/internal/cache"
	"snmpflapd/internal/config"
	"snmpflapd/internal/db"
	"snmpflapd/internal/logger"
	"snmpflapd/internal/server"

	"github.com/apex/log"
)

var (
	version = "unknown"
	build   = "unknown"
)

func main() {

	config.ReadFlags()

	if config.FlagVersion {
		fmt.Printf("FlapMyPort snmpflapd version %s, build %s\n", version, build)
		os.Exit(0)
	}

	config.ReadConfig()

	logger.SetUpLogger(config.Config.LogFilename, config.Config.LogLevel)

	db.CreateDB(
		config.Config.DBHost,
		config.Config.DBName,
		config.Config.DBUser,
		config.Config.DBPassword,
	)
	defer db.DB.Close()

	// Cache cleanup
	go func() {
		for {
			time.Sleep(time.Hour * 6)
			cache.CleanUp()
		}
	}()

	fmt.Println("Snmpflapd started.")

	listener := server.New()

	socket := fmt.Sprintf("%v:%v", config.Config.ListenAddress, config.Config.ListenPort)
	err := listener.Listen(socket)
	if err != nil {
		log.WithError(err).Error("Unable to create server.")
		os.Exit(1)
	}

}
