package localjson

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

type DiseaseDoc struct {
	ID       string   `json:"_id"`
	Name     string   `json:"name"`
	Symptoms []string `json:"symptoms"`
}

type Result struct {
	Disease string   `json:"disease"`
	Score   int      `json:"score"`
	Matched []string `json:"matched"`
}

type Store struct {
	set map[string]map[string]bool
}

func Load(path string) (*Store, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read diseases.json: %w", err)
	}
	var docs []DiseaseDoc
	if err := json.Unmarshal(b, &docs); err != nil {
		return nil, fmt.Errorf("parse diseases.json: %w", err)
	}

	set := make(map[string]map[string]bool, len(docs))
	for _, d := range docs {
		s := make(map[string]bool, len(d.Symptoms))
		for _, sym := range d.Symptoms {
			s[strings.ToLower(strings.TrimSpace(sym))] = true
		}
		set[d.Name] = s
	}
	return &Store{set: set}, nil
}

func (st *Store) Search(symptoms []string, topN int) []Result {
	q := normalize(symptoms)
	results := make([]Result, 0, 64)

	for disease, s := range st.set {
		matched := make([]string, 0, len(q))
		for _, sym := range q {
			if s[sym] {
				matched = append(matched, sym)
			}
		}
		if len(matched) > 0 {
			results = append(results, Result{Disease: disease, Score: len(matched), Matched: matched})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		return results[i].Disease < results[j].Disease
	})

	if topN <= 0 {
		topN = 10
	}
	if len(results) > topN {
		results = results[:topN]
	}
	return results
}

func normalize(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, x := range in {
		s := strings.ToLower(strings.TrimSpace(x))
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}
