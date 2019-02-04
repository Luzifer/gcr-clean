package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func deleteTags(delChan <-chan deleteRequest, wg *sync.WaitGroup) {
	limiter := make(chan struct{}, cfg.Parallel)

	for req := range delChan {
		limiter <- struct{}{}
		wg.Add(1)

		go func(req deleteRequest, limiter <-chan struct{}, wg *sync.WaitGroup) {
			defer func() { <-limiter }()
			defer wg.Done()

			logger := log.WithFields(log.Fields{
				"manifest": req.Digest,
				"repo":     req.Repo,
			})

			if !cfg.NoOp {
				if _, err := request(http.MethodDelete, fmt.Sprintf("%s/manifests/%s", req.Repo, req.Digest)); err != nil {
					logger.WithError(err).Error("Failed to delete manifest")
					return
				}
			}
			logger.WithField("noop", cfg.NoOp).Info("Manifest deleted")
		}(req, limiter, wg)
	}
}

func fetchRepositories(projectIDs []string, delChan chan deleteRequest, wg *sync.WaitGroup) error {
	defer wg.Done()
	log.Info("Fetching repositories...")

	response := struct {
		Repositories []string `json:"repositories"`
	}{}

	body, err := request(http.MethodGet, "_catalog")
	if err != nil {
		return errors.Wrap(err, "Could not fetch catalog")
	}

	if err := json.NewDecoder(body).Decode(&response); err != nil {
		return errors.Wrap(err, "Unable to unmarshal JSON response")
	}

	for _, repo := range response.Repositories {
		process := false
		for _, projectID := range projectIDs {
			if strings.HasPrefix(repo, projectID) {
				process = true
				continue
			}
		}

		if !process {
			log.WithField("repo", repo).Debug("Not in project scope, ignoring")
			continue
		}

		wg.Add(1)
		go func(repo string) {
			defer wg.Done()
			if err := fetchUntaggedManifests(repo, delChan, wg); err != nil {
				log.WithField("repo", repo).WithError(err).Error("Unable to fetch manifests")
			}
		}(repo)
	}

	return nil
}

func fetchUntaggedManifests(repo string, delChan chan deleteRequest, wg *sync.WaitGroup) error {
	body, err := request(http.MethodGet, fmt.Sprintf("%s/tags/list", repo))
	if err != nil {
		return errors.Wrap(err, "Unable to list tags")
	}

	response := struct {
		Manifests map[string]struct {
			Tags []string `json:"tag"`
		} `json:"manifest"`
	}{}

	if err := json.NewDecoder(body).Decode(&response); err != nil {
		return errors.Wrap(err, "Unable to unmarshal JSON response")
	}

	for digest, info := range response.Manifests {
		if len(info.Tags) == 0 {
			delChan <- deleteRequest{repo, digest}
			continue
		}

		log.WithFields(log.Fields{
			"repo":   repo,
			"digest": digest,
			"tags":   len(info.Tags),
		}).Debug("Manifest has tags, ignoring")
	}

	return nil
}

func request(method string, path string) (io.Reader, error) {
	logger := log.WithFields(log.Fields{
		"method": method,
		"path":   path,
	})

	uri := fmt.Sprintf("https://%s/v2/%s", cfg.Registry, path)

	req, err := http.NewRequest(method, uri, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create HTTP request")
	}

	req.SetBasicAuth("_json_key", getAuth())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.WithError(err).Debug("HTTP request failed")
		return nil, errors.Wrap(err, "HTTP request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		logger.WithField("status", resp.StatusCode).Debug("Status code indicated error")
		return nil, errors.Errorf("HTTP request failed with status HTTP %d", resp.StatusCode)
	}

	logger.Debug("Request success")

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, resp.Body)

	return buf, errors.Wrap(err, "Unable to read response body")
}
