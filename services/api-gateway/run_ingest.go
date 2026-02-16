//go:build ingest

package main

import (
	"bufio"
	"context"
	"flag"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/auriyoyo/cs-125-Context-Aware-Symptom-Search/services/api-gateway/pkg/config"
	"github.com/auriyoyo/cs-125-Context-Aware-Symptom-Search/services/api-gateway/pkg/database"
	"github.com/auriyoyo/cs-125-Context-Aware-Symptom-Search/services/api-gateway/pkg/datasources/conditions"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Conditions API response: [ code_systems, display_arrays, ef hash, codes (cf), total ]
const (
	respIdxCodes = 3
	respIdxEF    = 2
)

func main() {
	keywordsPath := flag.String("keywords", "", "Path to file with one search keyword per line")
	flag.Parse()

	if *keywordsPath == "" {
		log.Println("No keywords file provided. Use -keywords <path> to specify a file (one keyword per line).")
		os.Exit(0)
	}

	keywords, err := readKeywords(*keywordsPath)
	if err != nil {
		log.Fatalf("Read keywords file: %v", err)
	}
	if len(keywords) == 0 {
		log.Println("No keywords in file.")
		os.Exit(0)
	}

	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	if config.MongoDBURI == "" {
		log.Fatal("MONGODB_URI is required for ingest")
	}
	if err := database.Connect(config.MongoDBURI); err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer database.Disconnect()

	cfg := conditions.LoadConfig()
	client := conditions.NewClient(cfg.APIBaseURL, cfg.APIKey)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	coll := database.GetDatabase("clinical_tables").Collection("conditions")
	opts := options.Replace().SetUpsert(true)
	ef := []string{
		"primary_name", "consumer_name", "key_id",
		"icd10cm_codes", "icd10cm",
		"term_icd9_code", "term_icd9_text",
		"word_synonyms", "synonyms", "info_link_data",
	}

	var total int
	for _, kw := range keywords {
		req := &conditions.QueryRequest{Terms: []string{kw}, MaxList: 500, Ef: ef}
		resp, err := client.GetConditions(req)
		if err != nil {
			log.Fatalf("Conditions API query for %q: %v", kw, err)
		}
		docs, err := conditionsResponseToDocs(resp)
		if err != nil {
			log.Fatalf("Parse response: %v", err)
		}
		for _, doc := range docs {
			id := doc["_id"]
			if _, err := coll.ReplaceOne(ctx, bson.M{"_id": id}, doc, opts); err != nil {
				log.Fatalf("ReplaceOne %v: %v", id, err)
			}
			total++
		}
	}
	log.Printf("Upserted %d documents to clinical_tables.conditions", total)
}

func readKeywords(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		s := strings.TrimSpace(sc.Text())
		if s != "" {
			out = append(out, s)
		}
	}
	return out, sc.Err()
}

func conditionsResponseToDocs(resp interface{}) ([]bson.M, error) {
	arr, ok := resp.([]interface{})
	if !ok || len(arr) < 4 {
		return nil, nil
	}
	codes, _ := arr[respIdxCodes].([]interface{})
	efRaw := arr[respIdxEF]
	n := len(codes)
	if n == 0 {
		return nil, nil
	}

	ef, _ := efRaw.(map[string]interface{})
	docs := make([]bson.M, 0, n)
	for i := 0; i < n; i++ {
		id := extractICD10(ef, i)
		if id == "" {
			id = toStr(codes[i])
		}
		doc := bson.M{"_id": id, "key_id": toBSONValue(codes[i])}
		if ef != nil {
			for field, values := range ef {
				slice, _ := values.([]interface{})
				if i < len(slice) {
					doc[field] = toBSONValue(slice[i])
				}
			}
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

func extractICD10(ef map[string]interface{}, i int) string {
	if ef == nil {
		return ""
	}
	if v, ok := ef["icd10cm_codes"]; ok {
		slice, _ := v.([]interface{})
		if i < len(slice) {
			s, _ := slice[i].(string)
			s = strings.TrimSpace(s)
			if idx := strings.Index(s, ","); idx > 0 {
				s = s[:idx]
			}
			if s != "" {
				return s
			}
		}
	}
	if v, ok := ef["icd10cm"]; ok {
		slice, _ := v.([]interface{})
		if i < len(slice) {
			items, _ := slice[i].([]interface{})
			if len(items) > 0 {
				first, _ := items[0].(map[string]interface{})
				if c, _ := first["code"].(string); c != "" {
					return c
				}
			}
		}
	}
	return ""
}

func toStr(v interface{}) string {
	switch x := v.(type) {
	case string:
		return x
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	default:
		return ""
	}
}

func toBSONValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	switch x := v.(type) {
	case map[string]interface{}:
		m := make(bson.M, len(x))
		for k, val := range x {
			m[k] = toBSONValue(val)
		}
		return m
	case []interface{}:
		a := make(bson.A, len(x))
		for i, val := range x {
			a[i] = toBSONValue(val)
		}
		return a
	default:
		return v
	}
}
