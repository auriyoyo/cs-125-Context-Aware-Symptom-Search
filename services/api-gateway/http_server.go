package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sort"

	"github.com/auriyoyo/cs-125-Context-Aware-Symptom-Search/services/api-gateway/pkg/datasources/localjson"
	"github.com/auriyoyo/cs-125-Context-Aware-Symptom-Search/services/api-gateway/pkg/personalmodel"
)

type SearchResponse struct {
	Query   []string           `json:"query"`
	Results []localjson.Result `json:"results"`
}

// *****
func startHTTP(store *localjson.Store, pm *personalmodel.Store) {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		raw := r.URL.Query().Get("symptoms")
		if raw == "" {
			http.Error(w, "missing ?symptoms=fever,cough", http.StatusBadRequest)
			return
		}
		symptomsRaw := strings.Split(raw, ",")
		symptoms := make([]string, 0, len(symptomsRaw))
		for _, s := range symptomsRaw {
			s = strings.TrimSpace(s)
			if s != "" {
				symptoms = append(symptoms, s)
			}
}

		userID := r.URL.Query().Get("user") 

		var history []string
		if pm != nil && userID != "" {
			_ = pm.AppendSymptoms(r.Context(), userID, symptoms)
			if m, err := pm.Get(r.Context(), userID); err == nil {
				history = m.RecentSymptoms
			}
		}

		results := store.Search(symptoms, 10)

		// personalization boost (requires MatchCount helper)
		if len(history) > 0 {
			for i := range results {
				boost := store.MatchCount(results[i].Disease, history)
				results[i].Score += boost

			}
			sort.Slice(results, func(i, j int) bool {
				if results[i].Score != results[j].Score {
					return results[i].Score > results[j].Score
				}
				return results[i].Disease < results[j].Disease
			})
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(SearchResponse{Query: symptoms, Results: results})
	})

	log.Println("HTTP server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

