package antigravity

import "time"

// Internal projection types shared by the API parser and quota metrics.
type quotaPayload struct {
	Model      quotaModel                  `json:"model"`
	Quota      map[string]quotaBucketState `json:"quota"`
	Product    string                      `json:"product"`
	PlanTier   string                      `json:"plan_tier"`
	Email      string                      `json:"email"`
	ReceivedAt time.Time                   `json:"received_at"`
}

type quotaModel struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

type quotaBucketState struct {
	RemainingFraction *float64 `json:"remaining_fraction"`
	ResetTime         string   `json:"reset_time"`
	ResetInSeconds    *int64   `json:"reset_in_seconds"`
	Disabled          bool     `json:"disabled"`
}
