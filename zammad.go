package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/fatih/color"
)

// Notification models the payload sent to the Zammad API endpoint.
type Notification struct {
	CallID    string `json:"callId"`
	Event     string `json:"event"`
	Direction string `json:"direction"`
	From      string `json:"from"`
	To        string `json:"to"`
	Cause     string `json:"cause"`
}

// Enqueues an notification for delivery
func deliverNotification(n *Notification) {
	go func() {
		n.deliver()
	}()
}

// Delivers an notification
func (n *Notification) deliver() {
	color.New(color.FgCyan).Printf("Delivering Notification: %+v\n", n)

	jsonValue, _ := json.Marshal(n)
	req, err := http.NewRequest("POST", config.Zammad.Endpoint, bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Println("failed to create POST request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token token="+config.Zammad.Token)

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Println("failed to deliver notification:", err)
		return
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		log.Println("unexpected status code for notification:", res.StatusCode)
		return
	}
}
