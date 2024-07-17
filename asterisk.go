package main

import (
	"fmt"

	"github.com/CyCoreSystems/ari/v5"
	"github.com/CyCoreSystems/ari/v5/client/native"
	"github.com/inconshreveable/log15"
)

// Call is a currently active call
type Call struct {
	Caller string
	To     string
}

var (
	logger log15.Logger

	// Currently active channels
	channels = make(map[string]*Call)
	client   *native.Client
)

func startAsterisk() {
	logger = native.Logger
	logger.SetHandler(log15.StderrHandler)

	client = native.New(&native.Options{
		Application:  "zammad",
		Username:     config.Asterisk.Username,
		Password:     config.Asterisk.Password,
		URL:          fmt.Sprintf("http://%s:%d/ari", config.Asterisk.Host, config.Asterisk.Port),
		WebsocketURL: fmt.Sprintf("ws://%s:%d/ari/events", config.Asterisk.Host, config.Asterisk.Port),
	})

	logger.Info("connecting")
	err := client.Connect()

	if err != nil {
		panic(err)
	}

	logger.Info("subscribing")
	sub := client.Bus().Subscribe(nil, "StasisStart")

	logger.Info("waiting for events")
	go func() {
		for msg := range sub.Events() {
			handleEvent(msg)
		}
	}()
}

func handleEvent(msg ari.Event) {
	switch event := msg.(type) {
	case *ari.StasisStart:
		// New incoming call that should be notified
		channel := client.Channel().Get(event.Channel.Key)
		var to string

		if channel == nil {
			logger.Error("unable to get channel")
			return
		}

		if len(event.Args) > 0 {
			to = event.Args[0]
		}

		logger.Info("stasis start", "event", fmt.Sprintf("%+v", event))

		callerNumber := event.Channel.Caller.Number

		channels[event.Channel.ID] = &Call{
			Caller: callerNumber,
			To:     to,
		}

		deliverNotification(&Notification{
			CallID:    event.Channel.ID,
			Event:     "newCall",
			Direction: "in",
			From:      callerNumber,
			To:        to,
		})

		channel.Continue("", "", 0)

	case *ari.ChannelVarset:
		channelID := event.Channel.ID
		call := channels[channelID]
		if call == nil || event.Variable != "DIALSTATUS" || event.Value == "" {
			return
		}

		if event.Channel.State == "Ring" && event.Value == "ANSWER" {
			// Answered, now active
			logger.Info("Call answered")
			deliverNotification(&Notification{
				CallID: channelID,
				Event:  "answer",
				To:     call.To,
			})
		} else {
			// Finished
			logger.Info("Call finished", "reason", event.Value)
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
