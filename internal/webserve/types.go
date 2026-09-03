package webserve

import (
	"time"

	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/core"
)

const schemaVersion = "1"

// CatalogEntry is a display-name lookup for a registered provider.
type CatalogEntry struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Envelope is the JSON payload served at GET /api/v1/snapshots.
type Envelope struct {
	SchemaVersion          string               `json:"schema_version"`
	GeneratedAt            time.Time            `json:"generated_at"`
	AgentUsageVersion      string               `json:"agentusage_version"`
	Source                 string               `json:"source"`
	TimeWindow             string               `json:"time_window"`
	Theme                  string               `json:"theme"`
	RefreshIntervalSeconds int                  `json:"refresh_interval_seconds"`
	UsageMode              string               `json:"usage_mode"`
	TimeWindowLabel        string               `json:"time_window_label,omitempty"`
	OkCount                int                  `json:"ok_count"`
	WarnCount              int                  `json:"warn_count"`
	ErrCount               int                  `json:"err_count"`
	ProviderCount          int                  `json:"provider_count"`
	UnmappedCount          int                  `json:"unmapped_count,omitempty"`
	UnmappedPhrase         string               `json:"unmapped_phrase,omitempty"`
	Catalog                []CatalogEntry       `json:"catalog"`
	ThemeTokens            ThemeTokens          `json:"theme_tokens"`
	Views                  []AccountView        `json:"views"`
	Snapshots              []core.UsageSnapshot `json:"snapshots"`
}

// ThemeTokens carries palette values for the browser dashboard.
type ThemeTokens struct {
	Name     string `json:"name"`
	Icon     string `json:"icon,omitempty"`
	Base     string `json:"base"`
	Mantle   string `json:"mantle"`
	Surface0 string `json:"surface0"`
	Surface1 string `json:"surface1"`
	Surface2 string `json:"surface2"`
	Text     string `json:"text"`
	Subtext  string `json:"subtext"`
	Dim      string `json:"dim"`
	Accent   string `json:"accent"`
	Blue     string `json:"blue"`
	Sapphire string `json:"sapphire"`
	Green    string `json:"green"`
	Yellow   string `json:"yellow"`
	Red      string `json:"red"`
	Peach    string `json:"peach"`
	Teal     string `json:"teal"`
	Lavender string `json:"lavender"`
	Mauve    string `json:"mauve"`
}

// AccountView is a TUI-projected account payload for the web dashboard.
type AccountView struct {
	Key            string           `json:"key"`
	ProviderID     string           `json:"provider_id"`
	ProviderName   string           `json:"provider_name"`
	AccountID      string           `json:"account_id"`
	Status         string           `json:"status"`
	StatusBadge    string           `json:"status_badge"`
	StatusIcon     string           `json:"status_icon"`
	AccentColor    string           `json:"accent_color"`
	Summary        string           `json:"summary"`
	Detail         string           `json:"detail,omitempty"`
	TagEmoji       string           `json:"tag_emoji,omitempty"`
	TagLabel       string           `json:"tag_label,omitempty"`
	GaugePercent   float64          `json:"gauge_percent,omitempty"`
	Message        string           `json:"message,omitempty"`
	Timestamp      time.Time        `json:"timestamp"`
	TileLines      []string         `json:"tile_lines"`
	DetailSections []DetailSection  `json:"detail_sections"`
	DetailCards    []DetailCard     `json:"detail_cards,omitempty"`
	UsageLines     []UsageLine      `json:"usage_lines,omitempty"`
	Resets         []ResetPill      `json:"resets,omitempty"`
	DailyCost      []core.TimePoint `json:"daily_cost,omitempty"`
	CycleSchedule  string           `json:"cycle_schedule,omitempty"`
	LastRefreshed  string           `json:"last_refreshed,omitempty"`
	NextReset      string           `json:"next_reset,omitempty"`
	HasGauge       bool             `json:"has_gauge,omitempty"`
	HeaderTone     string           `json:"header_tone,omitempty"`

	// HTML fragments rendered from the same TUI functions the terminal uses.
	DetailHTML  string `json:"detail_html,omitempty"`
	BadgeHTML   string `json:"badge_html,omitempty"`
	IconHTML    string `json:"icon_html,omitempty"`
	StripHTML   string `json:"strip_html,omitempty"`
	SummaryHTML string `json:"summary_html,omitempty"`
	ResetHint   string `json:"reset_hint,omitempty"`
	// FrameHTML is a full TUI dashboard frame with this account selected.
	FrameHTML string `json:"frame_html,omitempty"`
}

type DetailCard struct {
	ID    string      `json:"id"`
	Title string      `json:"title"`
	Icon  string      `json:"icon,omitempty"`
	Color string      `json:"color,omitempty"`
	Rows  []DetailRow `json:"rows"`
}

type DetailRow struct {
	Kind    string   `json:"kind"`
	Label   string   `json:"label,omitempty"`
	Value   string   `json:"value,omitempty"`
	Hint    string   `json:"hint,omitempty"`
	Percent *float64 `json:"percent,omitempty"`
	Tone    string   `json:"tone,omitempty"`
}

type DetailSection struct {
	Title string   `json:"title"`
	Icon  string   `json:"icon,omitempty"`
	Lines []string `json:"lines"`
}

type ResetPill struct {
	Label    string `json:"label"`
	Duration string `json:"duration"`
	Urgent   bool   `json:"urgent,omitempty"`
}

type UsageLine struct {
	Label   string   `json:"label"`
	Short   string   `json:"short,omitempty"`
	Percent *float64 `json:"percent,omitempty"`
	Value   string   `json:"value,omitempty"`
	Hint    string   `json:"hint,omitempty"`
	ResetIn string   `json:"reset_in,omitempty"`
	Tone    string   `json:"tone,omitempty"`
	Urgent  bool     `json:"urgent,omitempty"`
}

// Options configures a Server.
type Options struct {
	ListenAddr     string
	AuthToken      string
	Source         string // auto | direct | daemon | demo
	TimeWindow     string
	Theme          string
	UsageMode      string
	WarnThreshold  float64
	CritThreshold  float64
	RefreshSeconds int
	Version        string
	Demo           bool
	AllowPublic    bool
	BasePath       string // URL prefix, e.g. "/agentusage". Empty means "/".
	Config         *config.Config
	Collect        CollectFunc // optional override for tests
	Now            func() time.Time
}

// CollectFunc produces a snapshot envelope. Tests substitute a stub.
type CollectFunc func() (Envelope, error)
