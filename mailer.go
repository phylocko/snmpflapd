package main

import (
	"sync"
	"time"
)

var mailQueue Queue

type Queue struct {
	events []linkEvent
	mx     sync.Mutex
}

// QueueLinkEvent stores events for further sending via email, etc...
func QueueLinkEvent(e linkEvent) {
	mailQueue.mx.Lock();
	defer mailQueue.mx.Unlock()
	mailQueue.events = append(mailQueue.events, e)
}

func grabQueue() []linkEvent {
	mailQueue.mx.Lock()
	defer mailQueue.mx.Unlock()

	events := make([]linkEvent, len(mailQueue.events))
	for i, event := range mailQueue.events {
		events[i] = event
	}
	events = make([]linkEvent, 0)
	return events
}

// RunQueue continuously checks and clears the notification queue
func RunQueue() {
	for {
		time.Sleep(queueInterval * time.Second)
		events := grabQueue()
		if config.SendMail {
			Notify(&events)
		}
	}
}

// Notify sends emails about linkEvents in current queue
func Notify(events *[]linkEvent) {
	// Will be implemented later

	//if len(*events) > 0 {
	//	for _, event := range *events {
	//		fmt.Println(event.String())
	//	}
	//}

}
