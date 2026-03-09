package liveevents

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type EventType string

const (
	EventTypeLocationUpdate EventType = "location_update"
	EventTypePublicAlert    EventType = "public_alert"
)

type AreaInfo struct {
	Country string `bson:"country" json:"country"`
	State   string `bson:"state" json:"state"`
	City    string `bson:"city" json:"city"`
}

type LiveEvent struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	EventType EventType     `bson:"event_type,omitempty" json:"event_type,omitempty"`
	UserID    string        `bson:"user_id,omitempty" json:"user_id,omitempty"`
	ZipCode   string        `bson:"zip_code" json:"zip_code"`
	CreatedAt time.Time     `bson:"created_at" json:"created_at"`
	ExpiresAt time.Time     `bson:"expires_at" json:"expires_at"`
	AreaInfo  AreaInfo      `bson:"area_info" json:"area_info"`
	Data      bson.M        `bson:"data" json:"data"`
}

type LocationPayload struct {
	Country string `json:"country"`
	State   string `json:"state"`
	City    string `json:"city"`
	ZipCode string `json:"zipCode"`
}

type ActiveWindow struct {
	Start time.Time `bson:"start" json:"start"`
	End   time.Time `bson:"end" json:"end"`
}

type AlertPayload struct {
	Provider     string       `json:"provider"`
	Hazard       string       `json:"hazard"`
	Severity     string       `json:"severity"`
	ActiveWindow ActiveWindow `json:"activeWindow"`
	SymptomTags  []string     `json:"symptomTags"`
	GuidanceURL  string       `json:"guidanceUrl"`
}

type UserContextProjection struct {
	UserID        string    `bson:"user_id" json:"userId"`
	LastKnownZip  string    `bson:"last_known_zip" json:"lastKnownZip"`
	LastUpdatedAt time.Time `bson:"last_updated_at" json:"lastUpdatedAt"`
	ExpiresAt     time.Time `bson:"expires_at" json:"expiresAt"`
}

type AreaRiskProjection struct {
	ZipCode     string       `bson:"zip_code" json:"zip_code"`
	ActiveRisks []ActiveRisk `bson:"active_risks" json:"active_risks"`
	UpdatedAt   time.Time    `bson:"updated_at" json:"updated_at"`
	ExpiresAt   time.Time    `bson:"expires_at" json:"expires_at"`
}

type ActiveRisk struct {
	Hazard      string    `bson:"hazard" json:"hazard"`
	Severity    string    `bson:"severity" json:"severity"`
	SymptomTags []string  `bson:"symptom_tags" json:"symptomTags"`
	GuidanceURL string    `bson:"guidance_url" json:"guidanceUrl"`
	ExpiresAt   time.Time `bson:"expires_at" json:"expiresAt"`
}

type RiskWarning struct {
	Hazard          string   `json:"hazard"`
	Severity        string   `json:"severity"`
	Message         string   `json:"message"`
	MatchedSymptoms []string `json:"matchedSymptoms,omitempty"`
	GuidanceURL     string   `json:"guidanceUrl,omitempty"`
}

type SearchEnrichment struct {
	RiskWarnings []RiskWarning `json:"riskWarnings,omitempty"`
	UserArea     string        `json:"userArea,omitempty"`
}

