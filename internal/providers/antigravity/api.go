package antigravity

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	defaultQuotaEndpoint = "https://daily-cloudcode-pa.googleapis.com/v1internal"
	quotaUserAgent       = "antigravity"
)

// APIError represents a typed HTTP error returned by the Antigravity quota API.
type APIError struct {
	StatusCode int
	Message    string
	RetryAfter time.Duration
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("retrieveUserQuotaSummary HTTP %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("retrieveUserQuotaSummary HTTP %d", e.StatusCode)
}

// IsUnauthorized reports whether the error is HTTP 401 Unauthorized.
func (e *APIError) IsUnauthorized() bool {
	return e != nil && e.StatusCode == http.StatusUnauthorized
}

// IsForbidden reports whether the error is HTTP 403 Forbidden.
func (e *APIError) IsForbidden() bool {
	return e != nil && e.StatusCode == http.StatusForbidden
}

// IsRateLimited reports whether the error is HTTP 429 Too Many Requests.
func (e *APIError) IsRateLimited() bool {
	return e != nil && e.StatusCode == http.StatusTooManyRequests
}

type quotaSummaryResponse struct {
	Groups      []quotaGroup `json:"groups"`
	Description string       `json:"description,omitempty"`
}

type quotaGroup struct {
	Buckets     []quotaBucket `json:"buckets"`
	DisplayName string        `json:"displayName,omitempty"`
	Description string        `json:"description,omitempty"`
}

type quotaBucket struct {
	BucketID          string   `json:"bucketId"`
	DisplayName       string   `json:"displayName,omitempty"`
	Window            string   `json:"window,omitempty"`
	ResetTime         string   `json:"resetTime,omitempty"`
	Description       string   `json:"description,omitempty"`
	RemainingFraction *float64 `json:"remainingFraction"`
}

func retrieveUserQuotaSummary(ctx context.Context, accessToken, baseURL string, client *http.Client) (quotaSummaryResponse, error) {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultQuotaEndpoint
	}
	apiURL := strings.TrimRight(baseURL, "/") + ":retrieveUserQuotaSummary"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader([]byte("{}")))
	if err != nil {
		return quotaSummaryResponse{}, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", quotaUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return quotaSummaryResponse{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var retryAfter time.Duration
		if raStr := resp.Header.Get("Retry-After"); raStr != "" {
			if secs, err := strconv.Atoi(raStr); err == nil && secs > 0 {
				retryAfter = time.Duration(secs) * time.Second
			}
		}
		return quotaSummaryResponse{}, &APIError{
			StatusCode: resp.StatusCode,
			Message:    http.StatusText(resp.StatusCode),
			RetryAfter: retryAfter,
		}
	}

	var decoded quotaSummaryResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return quotaSummaryResponse{}, fmt.Errorf("parse retrieveUserQuotaSummary: %w", err)
	}
	return decoded, nil
}

func quotaMapFromSummary(summary quotaSummaryResponse) map[string]quotaBucketState {
	out := make(map[string]quotaBucketState)
	for _, group := range summary.Groups {
		for _, bucket := range group.Buckets {
			id := strings.TrimSpace(bucket.BucketID)
			if id == "" {
				continue
			}
			out[id] = quotaBucketState{
				RemainingFraction: bucket.RemainingFraction,
				ResetTime:         strings.TrimSpace(bucket.ResetTime),
			}
		}
	}
	return out
}
