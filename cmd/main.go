package main

import (
	"flag"
	"fmt"
	"github.com/tanelmae/nats-tls/internal/config"
	"github.com/tanelmae/nats-tls/internal/pemgen"
	"log"
	"os"
)

// Version of this tool
var Version string = "dev"

func main() {
	confPath := flag.String("config", "nats-tls.yaml", "Path to config file")
	debug := flag.Bool("debug", false, "Run in debug mode")
	v := flag.Bool("v", false, "Version")

	flag.Parse()

	if *v {
		fmt.Printf("Version: %s\n", Version)
		os.Exit(0)
	}

	if _, err := os.Stat(*confPath); os.IsNotExist(err) {
		log.Printf("Config file %s doesn't exist", *confPath)
		os.Exit(1)
	}

	conf, err := config.ParseConfig(*confPath, *debug)
	if err != nil {
		log.Fatal(err)
	}

	ca, err := pemgen.GenCA(conf.CA)
	if err != nil {
		log.Fatal(err)
	}

	for _, signable := range conf.Signables {
		pemgen.GenSignedCerts(signable, *ca)
	}
}
