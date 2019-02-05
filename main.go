package main

import (
	"fmt"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"

	rconfig "github.com/Luzifer/rconfig/v2"
)

type deleteRequest struct{ Repo, Digest string }

var (
	cfg = struct {
		GoogleApplicationCredentials string `flag:"account" default:"" description:"Path to account.json file with GCR access"`
		Listen                       string `flag:"listen" default:":3000" description:"Port/IP to listen on"`
		LogLevel                     string `flag:"log-level" default:"info" description:"Log level (debug, info, warn, error, fatal)"`
		NoOp                         bool   `flag:"noop,n" default:"true" description:"Do not execute destructive DELETE operation"`
		Parallel                     int    `flag:"parallel,p" default:"10" description:"How many deletions to execute in parallel"`
		Registry                     string `flag:"registry" default:"gcr.io" description:"The registry used (gcr.io, eu.gcr.io, us.gcr.io, ...)"`
		VersionAndExit               bool   `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

	auth string

	version = "dev"
)

func init() {
	rconfig.AutoEnv(true)
	if err := rconfig.ParseAndValidate(&cfg); err != nil {
		log.Fatalf("Unable to parse commandline options: %s", err)
	}

	if cfg.VersionAndExit {
		fmt.Printf("gcr-clean %s\n", version)
		os.Exit(0)
	}

	if l, err := log.ParseLevel(cfg.LogLevel); err != nil {
		log.WithError(err).Fatal("Unable to parse log level")
	} else {
		log.SetLevel(l)
	}
}

func main() {
	args := rconfig.Args()
	if len(args) < 2 {
		log.Fatal("Expecting one or more positional arguments: gcr-clean <project id> [project id...]")
	}

	var (
		wg      = new(sync.WaitGroup)
		delChan = make(chan deleteRequest, 1000)
	)

	go deleteTags(delChan, wg)

	wg.Add(1)
	if err := fetchRepositories(args[1:], delChan, wg); err != nil {
		log.WithError(err).Error("An error occurred while fetching repos")
	}

	wg.Wait()
}
