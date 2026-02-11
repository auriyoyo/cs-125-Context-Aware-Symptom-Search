package healthapi

// QueryRequest models the JSON body sent to the Health API.
// All fields except Terms are optional.
type QueryRequest struct {

	// REQUIRED:
	Terms []string `json:"terms"` // search query
	// OPTIONAL:
	MaxList int `json:"maxList,omitempty"` // max number of results, not compatible with pagination

	// Pagination specs
	Count  int `json:"count,omitempty"`  // number of results to fetch
	Offset int `json:"offset,omitempty"` // offset (page number)

	// Additional query specs
	Q  string   `json:"q,omitempty"`  // additional wildcard queries
	Sf []string `json:"sf,omitempty"` // comma-separated list of field names to search through

	// Result manipulation
	Cf string   `json:"cui,omitempty"` // field name for key used in return key list
	Df string   `json:"df,omitempty"`  // field names for values in return values list
	Ef []string `json:"ef,omitempty"`  // additional info (not bound to string type) to be attached to results
}
