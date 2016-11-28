package main

import (
	"log"

	"github.com/abourget/ari"
	"github.com/fatih/color"
)

type Call struct {
	Caller string
}

var (
	// Currently active channels
	channels = make(map[string]*Call)
)

func startAsterisk() {
	client := ari.NewClient(
		config.Asterisk.Username,
		config.Asterisk.Password,
		config.Asterisk.Host,
		config.Asterisk.Port,
		"zammad")
	client.Debug = true
	client.SubscribeAll = true

	go func() {
		for msg := range client.LaunchListener() {
			log.Printf("%+v", msg)
			handleEvent(msg)
		}
	}()
}

func handleEvent(msg ari.Eventer) {
	switch event := msg.(type) {
	case *ari.ChannelCreated:
		channels[event.Channel.Id] = &Call{
			Caller: event.Channel.Caller.Number,
		}
	case *ari.ChannelDestroyed:
		delete(channels, event.Channel.Id)
	case *ari.ChannelVarset:
		if c, found := channels[event.Channel.Id]; found {
			if event.Variable == "zammad" {
				// New incoming call that should be notified
				deliverNotification(&Notification{
					CallID:    event.Channel.Id,
					Event:     "newCall",
					Direction: "in",
					From:      c.Caller,
					To:        event.Value,
				})
			} else if c.Caller != "" {
				if event.Variable == "DIALSTATUS" && event.Value != "" {
					if event.Channel.State == "Ring" && event.Value == "ANSWER" {
						// Answered, now active
						color.New(color.FgYellow).Printf("Call answered\n")
						deliverNotification(&Notification{
							CallID: event.Channel.Id,
							Event:  "answer",
						})
					} else {
						// Finished
						color.New(color.FgBlue).Printf("Call finished, reason=%s\n", event.Value)
						deliverNotification(&Notification{
							CallID: event.Channel.Id,
							Event:  "hangup",
							Cause:  event.Value,
						})
					}
				}
			}
		}
	}
}
