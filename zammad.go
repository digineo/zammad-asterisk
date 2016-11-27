package main

import (
	"bytes"
	"encoding/json"
	"github.com/fatih/color"
	"log"
	"net/http"
)

type Event struct {
	Type      string
	CallId    string
	Direction string `json:"omitempty"`
	From      string `json:"omitempty"`
	To        string `json:"omitempty"`
}

var (
	// Channel for outgoing events
	eventChan = make(chan *Event, 10)
)

func startZammad() {
	go func() {
		for event := range eventChan {
			event.deliver()
		}
	}()
}

// Enqueues an event for delivery
func deliverEvent(event *Event) {
	eventChan <- event
}

// Delivers an event
func (event *Event) deliver() {

	c := color.New(color.FgCyan)
	c.Printf("Delivering Event: %+v\n", event)

	if config.Zammad.Endpoint != "" {
		jsonValue, _ := json.Marshal(event)

		_, err := http.Post(config.Zammad.Endpoint, "application/json", bytes.NewBuffer(jsonValue))
		if err != nil {
			log.Println(err)
		}
	}
}
