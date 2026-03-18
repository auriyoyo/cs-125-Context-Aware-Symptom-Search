package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/auriyoyo/cs-125-Context-Aware-Symptom-Search/services/api-gateway/pkg/datasources/liveevents"
	"github.com/auriyoyo/cs-125-Context-Aware-Symptom-Search/services/api-gateway/pkg/datasources/localjson"
	"github.com/auriyoyo/cs-125-Context-Aware-Symptom-Search/services/api-gateway/pkg/personalmodel"
)

type RankedResult struct {
	Result       localjson.Result `json:"result"`
	BaseScore    int              `json:"baseScore"`
	HistoryBoost int              `json:"historyBoost,omitempty"`
	EventBoost   int              `json:"eventBoost,omitempty"`
	FinalScore   int              `json:"finalScore"`
	Reasons      []string         `json:"reasons,omitempty"`
}

type SearchResponse struct {
	Query        []string                 `json:"query"`
	Results      []RankedResult       	  `json:"results"`
	RiskWarnings []liveevents.RiskWarning `json:"riskWarnings,omitempty"`
	UserArea     string                   `json:"userArea,omitempty"`
}

const (
	baseScoreWeight    = 2
	historyBoostWeight = 1
	eventBoostWeight   = 1
)

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func startHTTP(store *localjson.Store, pm *personalmodel.Store, le *liveevents.Store) {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)

		if r.Method == http.MethodOptions {
			return
		}
		w.Write([]byte("ok"))
	})

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)

		if r.Method == http.MethodOptions {
			return
		}

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

		if len(symptoms) == 0 {
			http.Error(w, "no valid symptoms provided", http.StatusBadRequest)
			return
		}

		userID := r.URL.Query().Get("user")

		var history []string
		if pm != nil && userID != "" {
			_ = pm.AppendSymptoms(r.Context(), userID, symptoms)
			if m, err := pm.Get(r.Context(), userID); err == nil {
				history = m.RecentSymptoms
			}
		}

		baseResults := store.Search(symptoms, 10)
		var warnings []liveevents.RiskWarning
		var userArea string

		if le != nil {
			if enrichment, err := le.EnrichForUser(r.Context(), userID, symptoms); err == nil && enrichment != nil {
				warnings = enrichment.RiskWarnings
				userArea = enrichment.UserArea
			}
		}

		ranked := make([]RankedResult, 0, len(baseResults))
		for _, res := range baseResults {
			baseScore := res.Score

			rr := RankedResult{
				Result:    res,
				BaseScore: baseScore,
				FinalScore: baseScore * baseScoreWeight,
				Reasons:   []string{},
			}

			if baseScore > 0 {
				rr.Reasons = append(rr.Reasons, "Matched the entered symptoms.")
			}		

			// personalization boost (requires MatchCount helper)
			if len(history) > 0 {
				rawHistoryBoost := store.MatchCount(rr.Result.Disease, history)
				historyBoost := rawHistoryBoost * historyBoostWeight
				rr.HistoryBoost = historyBoost
				rr.FinalScore += historyBoost

				if historyBoost > 0 {
					rr.Reasons = append(rr.Reasons, "Boosted by recent symptom history.")
				}
			}

			if len(warnings) > 0 {
				rawEventBoost, eventReasons := computeEventBoost(rr.Result, warnings, store)
				eventBoost := rawEventBoost * eventBoostWeight
				rr.EventBoost = eventBoost
				rr.FinalScore += eventBoost
				rr.Reasons = append(rr.Reasons, eventReasons...)
			}

			rr.Result.Score = rr.FinalScore
			ranked = append(ranked, rr)
		}

		sort.Slice(ranked, func(i, j int) bool {
			if ranked[i].Result.Score != ranked[j].Result.Score {
				return ranked[i].Result.Score > ranked[j].Result.Score
			}
			return ranked[i].Result.Disease < ranked[j].Result.Disease
		})

		resp := SearchResponse{
			Query:        symptoms,
			Results:      ranked,
			RiskWarnings: warnings,
			UserArea:     userArea,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
	log.Println("HTTP server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func computeEventBoost(result localjson.Result, warnings []liveevents.RiskWarning, store *localjson.Store) (int, []string) {
	boost := 0
	var reasons []string

	for _, w := range warnings {
		match := store.MatchCount(result.Disease, w.MatchedSymptoms)
		if match == 0 {
			continue
		}

		severityBonus := severityWeight(w.Severity)
		totalBoost := match + severityBonus
		boost += totalBoost

		reason := "Boosted by active local " + w.Hazard + " alert"
		if len(w.MatchedSymptoms) > 0 {
			reason += " matching symptoms: " + strings.Join(w.MatchedSymptoms, ", ")
		}
		if w.Severity != "" {
			reason += " (severity weight: " + w.Severity + ")"
		}
		reasons = append(reasons, reason+".")
	}

	return boost, reasons
}

func severityWeight(severity string) int {
	switch strings.ToLower(strings.TrimSpace(severity)) {
	case "high", "severe", "unhealthy":
		return 3
	case "moderate", "medium":
		return 2
	case "low", "minor":
		return 1
	default:
		return 0
	}
}