package logger

import (
	"fmt"
	"os"

	"github.com/apex/log"
	"github.com/apex/log/handlers/multi"
	"github.com/apex/log/handlers/text"
)

func SetUpLogger(filename string, level string) {

	multiHandler := multi.New()

	if filename == "" {
		multiHandler.Handlers = append(multiHandler.Handlers, text.New(os.Stdout))
	} else {

		file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Println("Unable to open log file.")
			os.Exit(1)
		}
		multiHandler.Handlers = append(multiHandler.Handlers, text.New(file))

	}

	logLevel, err := log.ParseLevel(level)
	if err != nil {
		fmt.Println("Wrong log level configured.")
		os.Exit(1)
	}

	log.SetHandler(multiHandler)
	log.SetLevel(logLevel)

}
