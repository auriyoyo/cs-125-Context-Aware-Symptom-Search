package healthapi

// QueryRequest models the JSON body sent to the Health API.
// All fields except Terms are optional.
type QueryRequest struct {
	Terms   []string `json:"terms"`
	MaxList int      `json:"maxList,omitempty"`
	Count   int      `json:"count,omitempty"`
	Offset  int      `json:"offset,omitempty"`
	Q       string   `json:"q,omitempty"`
	Df      string   `json:"df,omitempty"`
	Sf      []string `json:"sf,omitempty"`
	Cf      string   `json:"cui,omitempty"`
	Ef      []string `json:"ef,omitempty"`
}
