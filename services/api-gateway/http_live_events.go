package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/auriyoyo/cs-125-Context-Aware-Symptom-Search/services/api-gateway/pkg/datasources/liveevents"
)

type locationRequest struct {
	UserID  string `json:"userId"`
	Country string `json:"country"`
	State   string `json:"state"`
	City    string `json:"city"`
	ZipCode string `json:"zipCode"`
}

func registerLiveEventHandlers(store *liveevents.Store) {
	if store == nil {
		return
	}
	http.HandleFunc("/events/location", handleLocationEvent(store))
	http.HandleFunc("/context/user/events", handleUserEvents(store))
	http.HandleFunc("/context/area/risks", handleAreaRisks(store))
}

func handleLocationEvent(store *liveevents.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		if r.Method == http.MethodOptions {
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req locationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		if req.UserID == "" || req.Country == "" || req.State == "" || req.City == "" || req.ZipCode == "" {
			http.Error(w, "userId, country, state, city, and zipCode are required", http.StatusBadRequest)
			return
		}

		loc := liveevents.LocationPayload{
			Country: req.Country,
			State:   req.State,
			City:    req.City,
			ZipCode: req.ZipCode,
		}
		if err := store.AppendLocationUpdate(r.Context(), req.UserID, loc, "mobile_app"); err != nil {
			http.Error(w, "failed to record location event", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

func handleUserEvents(store *liveevents.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		if r.Method == http.MethodOptions {
			return
		}

		userID := r.URL.Query().Get("user")
		if userID == "" {
			http.Error(w, "missing ?user=...", http.StatusBadRequest)
			return
		}

		events, err := store.QueryUserEvents(r.Context(), userID, 24*time.Hour)
		if err != nil {
			http.Error(w, "query failed", http.StatusInternalServerError)
			return
		}
		if events == nil {
			events = []liveevents.LiveEvent{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(events)
	}
}

func handleAreaRisks(store *liveevents.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		if r.Method == http.MethodOptions {
			return
		}

		zipCode := r.URL.Query().Get("zip")
		if zipCode == "" {
			http.Error(w, "missing ?zip=...", http.StatusBadRequest)
			return
		}

		risks, err := store.GetAreaRisks(r.Context(), zipCode)
		if err != nil {
			http.Error(w, "query failed", http.StatusInternalServerError)
			return
		}
		if risks == nil {
			risks = &liveevents.AreaRiskProjection{ZipCode: zipCode}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(risks)
	}
}
