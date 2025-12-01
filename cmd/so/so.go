package main

import (
	"context"
	"encoding/json"
	"flag"
	"log/slog"
	"net/http"
	"time"

	"spikeotter"
)

var (
	fLoadFactor = flag.Int("loadfactor", 10, "loadFactor")
	fUniques    = flag.Int("unique", 3000000, "unique values")
	fMaxsize    = flag.Int("maxsize", 1000000, "max cache size")
	fExpiry     = flag.Duration("expiry", time.Minute, "expiry")
	fRefresh    = flag.Duration("refresh", time.Minute, "refresh")
)

func main() {
	flag.Parse()
	ctx := context.TODO()
	slog.LogAttrs(ctx, slog.LevelInfo, "hello")

	cache := spikeotter.NewCache(*fUniques, *fLoadFactor, *fMaxsize, *fExpiry, *fRefresh)
	bcache := spikeotter.NewBCache(*fUniques, *fLoadFactor, *fMaxsize, *fExpiry, *fRefresh)

	http.HandleFunc("/cache", func(w http.ResponseWriter, r *http.Request) {
		ids := cache.GenIDs()
		models, err := cache.BulkGet(r.Context(), ids)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		encoder := json.NewEncoder(w)
		err = encoder.Encode(models)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/bcache", func(w http.ResponseWriter, r *http.Request) {
		ids := bcache.GenIDs()
		models, err := bcache.BulkGet(r.Context(), ids)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		encoder := json.NewEncoder(w)
		err = encoder.Encode(models)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	source := cache.Source()

	http.HandleFunc("/source", func(w http.ResponseWriter, r *http.Request) {
		ids := source.GenIDs()
		models, err := source.BulkGet(r.Context(), ids)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		encoder := json.NewEncoder(w)
		err = encoder.Encode(models)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	go func() {
		cache.StatsLoop(ctx)
	}()

	go func() {
		bcache.StatsLoop(ctx)
	}()
	err := http.ListenAndServeTLS(":8443", "crt.pem", "key.pem", http.DefaultServeMux)
	slog.LogAttrs(ctx, slog.LevelError, "http.ListenAndServeTLS", slog.Any("error", err))
}
