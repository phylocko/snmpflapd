package main

import (
	"fmt"
	"net"
	"time"
)

func sendTraps() {
	time.Sleep(time.Second * 2)
	fmt.Println("Test started")

	for {
		for i := 0; i < 200; i++ {
			event := linkEvent{}
			event.ipAddress = net.IP{192, 168, 48, 43}
			event.ifIndex = 1010
			event.time = time.Now().Local()

			event.FillHostName()
			event.FillIfName()
			event.FillIfAlias()
			// err := connector.saveLinkEvent(&event)
			// if err != nil {
			// 	fmt.Printf("%v\n", err)
			// }
		}
		break
		//time.Sleep(time.Second * 4)
	}
}
