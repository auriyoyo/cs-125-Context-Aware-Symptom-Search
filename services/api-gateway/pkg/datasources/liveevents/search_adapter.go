package liveevents

import (
	"context"
	"strings"
)

func (s *Store) EnrichForUser(ctx context.Context, userID string, queriedSymptoms []string) (*SearchEnrichment, error) {
	if !s.cfg.EnableSearchEnrichment || userID == "" {
		return nil, nil
	}

	userCtx, err := s.GetUserContext(ctx, userID)
	if err != nil || userCtx == nil {
		return nil, err
	}

	risks, err := s.GetAreaRisks(ctx, userCtx.LastKnownZip)
	if err != nil || risks == nil || len(risks.ActiveRisks) == 0 {
		return nil, err
	}

	warnings := buildWarnings(risks.ActiveRisks, queriedSymptoms)
	if len(warnings) == 0 {
		return nil, nil
	}

	return &SearchEnrichment{
		RiskWarnings: warnings,
		UserArea:     userCtx.LastKnownZip,
	}, nil
}

func buildWarnings(risks []ActiveRisk, symptoms []string) []RiskWarning {
	symptomSet := make(map[string]bool, len(symptoms))
	for _, s := range symptoms {
		symptomSet[strings.ToLower(strings.TrimSpace(s))] = true
	}

	var warnings []RiskWarning
	for _, r := range risks {
		matched := findOverlap(r.SymptomTags, symptomSet)
		if len(matched) > 0 {
			warnings = append(warnings, RiskWarning{
				Hazard:          r.Hazard,
				Severity:        r.Severity,
				Message:         r.Hazard + " advisory active in your area (severity: " + r.Severity + ")",
				MatchedSymptoms: matched,
				GuidanceURL:     r.GuidanceURL,
			})
		}
	}
	return warnings
}

func findOverlap(tags []string, set map[string]bool) []string {
	var out []string
	for _, t := range tags {
		if set[strings.ToLower(t)] {
			out = append(out, t)
		}
	}
	return out
}
