package server

import (
	"encoding/json"
	"guilhermec-costa/go-web-crawler/crawler"
	"guilhermec-costa/go-web-crawler/crawler/cli"
	"log/slog"
	"net/http"
)

func setJSONContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

func health(w http.ResponseWriter, r *http.Request) {
	setJSONContentType(w)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func startCrawlerJob(w http.ResponseWriter, r *http.Request) {
	var payload cli.CrawlerFlagsJSON
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&payload); err != nil {
		slog.Info("Invalid json", "err", err)
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := payload.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	args := cli.DefaultArgs()
	args.RootUrl = payload.RootUrl
	args.Depth = payload.Depth
	args.Verbose = true
	go crawler.Bootstrap(args)

	w.WriteHeader(200)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "job started",
	})
}
