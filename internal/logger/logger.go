package logger

import (
	"io"
	"log"
	"os"
)

var L *log.Logger

func SetUpLogger(filename string) {

	var output io.Writer
	var err error

	if filename == "" {
		output = os.Stdout

	} else {
		output, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalln(err)
		}
	}

	L = log.New(output, "flapmyport", log.LstdFlags)

}
