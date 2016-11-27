package main

import (
	"log"
	"strings"
	"time"

	"github.com/party79/gami"
	"github.com/party79/gami/event"
)

var (
	// Currently active calls
	calls = make(map[string]bool)
)

func startAsterisk() {
	if len(config.Asterisk.Incoming) == 0 {
		log.Fatal("No incoming channels configured")
	}

	// Connect to asterisk
	log.Println("Connecting to", config.Asterisk.Endpoint)
	ami, err := gami.Dial(config.Asterisk.Endpoint)
	if err != nil {
		log.Fatal(err)
	}

	ami.Run()

	// Sorgt dafür, dass nur bestimmte Events empfangen werden
	setEventMask := func() {
		if _, err := ami.Action("Events", gami.Params{"EventMask": "call"}); err != nil {
			log.Println("failed to set event mask:", err)
		}
	}

	// install manager
	go func() {
		defer ami.Close()
		for {
			select {
			//handle network errors
			case err := <-ami.NetError:
				if shutdown {
					return
				}
				log.Println("Network Error:", err)

				// try new connection every second
				<-time.After(time.Second)
				if err := ami.Reconnect(); err == nil {
					setEventMask()
				}

			case err := <-ami.Error:
				log.Println("error:", err)

			case ev := <-ami.Events:
				handleEvent(ev)
			}
		}
	}()

	log.Println("Logging in as", config.Asterisk.Username)
	if err := ami.Login(config.Asterisk.Username, config.Asterisk.Password); err != nil {
		log.Fatal(err)
	}
	log.Println("Login successful")

	setEventMask()
}

func handleEvent(ev *gami.AMIEvent) {
	// log.Printf("EVENT: %+v", ev)
	switch e := event.New(ev).(type) {
	case *event.Dial:
		if (e.DialStatus == "ANSWER" || e.DialStatus == "CANCEL") && isCall(e.UniqueID) {
			removeCall(e.UniqueID)
			deliverEvent(&Event{
				Type:   e.DialStatus,
				CallId: e.UniqueID,
			})
		}
	case *event.Newstate:
		if e.ChannelStateDesc == "Ring" {
			channel := e.Channel
			// strip channel extension
			if i := strings.IndexRune(channel, '-'); i != -1 {
				channel = channel[:i]
			}

			if isIncomingChannel(channel) {
				addCall(e.UniqueID)
				deliverEvent(&Event{
					Type:      "newCall",
					Direction: "in",
					CallId:    e.UniqueID,
					From:      e.CallerIDNum,
					To:        channel,
				})
			}
		}
	default:
		//log.Printf("Other Event: %+v", e)
	}
}

// Adds the call ID to the active calls
func addCall(id string) {
	calls[id] = true
}

// Checks whether the call ID belongs to an active call
func isCall(id string) bool {
	_, found := calls[id]
	return found
}

// Removes the call ID from the active calls
func removeCall(id string) {
	delete(calls, id)
}

// Checks whether the given channel is incoming
func isIncomingChannel(channel string) bool {
	for _, c := range config.Asterisk.Incoming {
		if c == channel {
			return true
		}
	}

	return false
}
