package main

import (
	"flag"
	"fmt"
	"os"
)

type ClientConfig struct {
	server		string
	timeout		uint
}

func initializeConfig() ClientConfig {
	var config = ClientConfig{}

	// Describe all flags
	flag.StringVar(&config.server, "server", "1.1.1.1", "Upstream DNS server")
	flag.UintVar(&config.timeout, "timeout", 3, "Request timeout, in seconds")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			"Usage: %s [options] hostname1 hostname2 ...\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Parse any command-line arguments
	flag.Parse()
	if (flag.NArg() == 0) {
		flag.Usage()
	}

	return(config)
}


func main() {
	config := initializeConfig()
	for _, host := range flag.Args() {
		resolve(config, host)
	}

	return
}
