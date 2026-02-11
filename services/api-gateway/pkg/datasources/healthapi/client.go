package healthapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Send a GET request using QueryRequest params
func (c *Client) GetConditions(req *QueryRequest) (interface{}, error) {
	params := url.Values{}
	if len(req.Terms) > 0 {
		params.Set("terms", strings.Join(req.Terms, " "))
	}
	if req.MaxList > 0 {
		params.Set("maxList", strconv.Itoa(req.MaxList))
	}
	if req.Count > 0 {
		params.Set("count", strconv.Itoa(req.Count))
	}
	if req.Offset > 0 {
		params.Set("offset", strconv.Itoa(req.Offset))
	}
	if req.Q != "" {
		params.Set("q", req.Q)
	}
	if req.Df != "" {
		params.Set("df", req.Df)
	}
	if len(req.Sf) > 0 {
		params.Set("sf", strings.Join(req.Sf, ","))
	}
	if req.Cf != "" {
		params.Set("cui", req.Cf)
	}
	if len(req.Ef) > 0 {
		params.Set("ef", strings.Join(req.Ef, ","))
	}

	requestURL := c.baseURL
	if qs := params.Encode(); qs != "" {
		requestURL += "?" + qs
	}

	httpReq, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, err
	}
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}
