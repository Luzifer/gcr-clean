package main

import (
	"io/ioutil"
	"os"

	"github.com/genuinetools/reg/repoutils"
	log "github.com/sirupsen/logrus"
)

func getAuth() string {
	if auth != "" {
		return auth
	}

	// If specified use Application Default Credentials
	if _, err := os.Stat(cfg.GoogleApplicationCredentials); err == nil {
		jsonData, err := ioutil.ReadFile(cfg.GoogleApplicationCredentials)
		if err != nil {
			log.WithError(err).Fatal("Unable to read GoogleApplicationCredentials file")
		}

		auth = string(jsonData)
	}

	// No luck yet? Try Docker auth
	if auth == "" {
		if ac, err := repoutils.GetAuthConfig("", "", cfg.Registry); err == nil && ac.Password != "" {
			auth = ac.Password
		}
	}

	if auth == "" {
		log.Fatal("No valid credentials found for registry")
	}

	return auth
}
