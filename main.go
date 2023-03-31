package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"
)

// Config holds the runtime configuration for this program. It is parsed
// from a TOML file on startup.
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

var config = Config{}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage:", os.Args[0], "path/to/config.cfg")
		os.Exit(1)
	}

	// Configure log format
	flags := log.Lshortfile
	if os.Getenv("JOURNAL_STREAM") == "" {
		// not running as systemd service, add timestamps
		flags |= log.LstdFlags
	}
	log.SetFlags(flags)

	// Parse config file
	if _, err := toml.DecodeFile(os.Args[1], &config); err != nil {
		log.Fatal(err)
	}

	startAsterisk()

	// Wait for SIGINT or SIGTERM
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
