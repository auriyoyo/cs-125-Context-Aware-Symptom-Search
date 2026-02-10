package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/auriyoyo/cs-125-Context-Aware-Symptom-Search/services/api-gateway/pkg/datasources/localjson"
)

type SearchResponse struct {
	Query   []string           `json:"query"`
	Results []localjson.Result `json:"results"`
}

func startHTTP(store *localjson.Store) {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		raw := r.URL.Query().Get("symptoms")
		if raw == "" {
			http.Error(w, "missing ?symptoms=fever,cough", http.StatusBadRequest)
			return
		}
		symptoms := strings.Split(raw, ",")
		results := store.Search(symptoms, 10)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(SearchResponse{Query: symptoms, Results: results})
	})

	log.Println("HTTP server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
