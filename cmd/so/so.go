package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"spikeotter"
)

func main() {
	ctx := context.TODO()
	slog.LogAttrs(ctx, slog.LevelInfo, "hello")

	cache := spikeotter.NewCache()

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

	err := http.ListenAndServeTLS(":8443", "crt.pem", "key.pem", http.DefaultServeMux)
	slog.LogAttrs(ctx, slog.LevelError, "http.ListenAndServeTLS", slog.Any("error", err))
}
