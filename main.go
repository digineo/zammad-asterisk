package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Asterisk struct {
		Host     string
		Port     int
		Username string
		Password string
	}
	Zammad struct {
		Endpoint string
		Token    string
	}
}

var (
	config   = Config{}
	shutdown = false
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage:", os.Args[0], "path/to/config.cfg")
		os.Exit(1)
	}

	// Parse config file
	if _, err := toml.DecodeFile(os.Args[1], &config); err != nil {
		log.Fatal(err)
	}

	startAsterisk()
	startZammad()

	// Wait for SIGINT or SIGTERM
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	log.Println("Shutting down")
	shutdown = true
}
