package main

import (
	"log"

	"github.com/abourget/ari"
	"github.com/fatih/color"
)

// Call is a currently active call
type Call struct {
	Caller string
	To     string
}

var (
	// Currently active channels
	channels = make(map[string]*Call)
	client   *ari.Client
)

func startAsterisk() {
	client = ari.NewClient(
		config.Asterisk.Username,
		config.Asterisk.Password,
		config.Asterisk.Host,
		config.Asterisk.Port,
		"zammad",
	)

	// get notified about answered and finished calls outside of the stasis
	client.SubscribeAll = true

	go func() {
		for msg := range client.LaunchListener() {
			// log.Printf("%+v", msg)
			handleEvent(msg)
		}
	}()
}

func handleEvent(msg ari.Eventer) {
	switch event := msg.(type) {
	case *ari.StasisStart:
		// New incoming call that should be notified
		channel, err := client.Channels.Get(event.Channel.ID)
		var to string

		if err != nil {
			log.Println("unable to get channel:", err)
			return
		}

		if len(event.Args) > 0 {
			to = event.Args[0]
		}

		channels[event.Channel.ID] = &Call{
			Caller: event.Channel.Caller.Number,
			To:     to,
		}

		deliverNotification(&Notification{
			CallID:    event.Channel.ID,
			Event:     "newCall",
			Direction: "in",
			From:      channel.Caller.Number,
			To:        to,
		})

		channel.ContinueInDialplan("", "", 0, "")

	case *ari.ChannelVarset:
		channelID := event.Channel.ID
		call := channels[channelID]
		if call == nil || event.Variable != "DIALSTATUS" || event.Value == "" {
			return
		}

		if event.Channel.State == "Ring" && event.Value == "ANSWER" {
			// Answered, now active
			color.New(color.FgYellow).Printf("Call answered\n")
			deliverNotification(&Notification{
				CallID: channelID,
				Event:  "answer",
				To:     call.To,
			})
		} else {
			// Finished
			color.New(color.FgBlue).Printf("Call finished, reason=%s\n", event.Value)
			deliverNotification(&Notification{
				CallID: channelID,
				Event:  "hangup",
				To:     call.To,
				Cause:  event.Value,
			})
		}

	case *ari.ChannelDestroyed:
		delete(channels, event.Channel.ID)
	}
}
