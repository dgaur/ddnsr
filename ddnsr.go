package main

import (
	"flag"
	"fmt"
	"net"
	"os"
)

type ClientConfig struct {
	raw			bool
	recursive	bool
	rtype		string
	server		string
	timeout		uint
}


func initializeConfig() ClientConfig {
	var config = ClientConfig{}

	// Describe all flags
	flag.BoolVar(&config.raw, "raw", false, "Show the raw packet bytes?")
	flag.BoolVar(&config.recursive, "recursive", true,
		"Send a recursive DNS query?")
	flag.StringVar(&config.rtype, "rtype", "A",
		"DNS record type (A, ALL, CNAME, MX, PTR, SOA, TXT, etc)")
	flag.StringVar(&config.server, "server", "1.1.1.1",
		"IP address of upstream DNS server")
	flag.UintVar(&config.timeout, "timeout", 3, "Request timeout, in seconds")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			"Usage: %s [options] hostname1 hostname2 ...\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Parse + validate any command-line arguments
	flag.Parse()
	if (flag.NArg() == 0) {
		flag.Usage()
	}
	if (net.ParseIP(config.server) == nil) {
		fmt.Fprintf(flag.CommandLine.Output(),
			"Invalid DNS server: %s\n", config.server)
		flag.Usage()
	}
	if (RecordTypeMapToType[config.rtype] == 0) {
		fmt.Fprintf(flag.CommandLine.Output(),
			"Invalid record type: %s", config.rtype)
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
