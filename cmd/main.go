package main

import (
	"flag"
	"github.com/tanelmae/nats-tls/internal/config"
	"github.com/tanelmae/nats-tls/internal/pemgen"
	"log"
	"os"
)

func main() {
	confPath := flag.String("config", "nats-tls.yaml", "Path to config file")
	debug := flag.Bool("debug", false, "Run in debug mode")
	flag.Parse()

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

	pemgen.GenSignedCerts(conf.Route, *ca)
	pemgen.GenSignedCerts(conf.Server, *ca)
	pemgen.GenSignedCerts(conf.Client, *ca)
}
