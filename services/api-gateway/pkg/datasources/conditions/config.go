package conditions

import (
	"os"
	"time"
)

type Config struct {
	APIBaseURL   string
	APIKey       string
	DatabaseName string
	SyncInterval time.Duration
}

func LoadConfig() *Config {
	return &Config{
		APIBaseURL:   getEnv("HEALTH_API_BASE_URL", "https://clinicaltables.nlm.nih.gov/api/conditions/v3/search"),
		APIKey:       getEnv("HEALTH_API_KEY", ""),
		DatabaseName: getEnv("HEALTH_API_DATABASE", "health_api_data"),
		SyncInterval: 1 * time.Hour,
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Field	Field Description
// primary_name	The name of the medical condition.
// consumer_name	A more consumer-friendly version of the disease name in the primary_name field.
// key_id	A unique idenitifier (a code) for the medical condition
// icd10cm_codes	A comma-separated list of suggested ICD-10-CM codes for the condition. Note that some of the ICD10CM codes may have a trailing question mark. In the SNOMED to ICD-10-CM mapping, question marks are used as placeholders for episode in case of injury codes like T code or S code, since there is no episode information in SNOMED concepts. When coding the patients data, the user will need to replace the question marks with episode codes like A,....S if the information is available in the patients' records
// icd10cm	An list of code and display name pairs for the suggested ICD-10-CM codes for the condition. When requested with the "ef" parameter, this will be an array of objects, where each object will have "code" and "name" properties. Note that some of the ICD10CM codes may have a trailing question mark, see "icd10cm_codes" above for more details.
// term_icd9_code	The ICD-9-CM code for the term.
// term_icd9_text	The ICD-9-CM display text for the term.
// word_synonyms	Synonyms for each of the words in the term.
// synonyms	Synonyms for the whole term.
// info_link_data	Links to information about the condition. The returned data is an array of arrays. There is one inner array for each available link about the condition, and each inner array contains the link and the linked page's title (which is useful if you are displaying the list of links for the user to pick one).
