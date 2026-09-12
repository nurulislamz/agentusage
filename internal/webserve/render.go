package webserve

import (
	"fmt"
	"html/template"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
)

// layoutMeta mirrors the client-side layout switcher metadata.
type layoutMeta struct {
	ID    string
	Label string
	Icon  string
	Hint  string
}

var layoutList = []layoutMeta{
	{ID: "split", Label: "Split", Icon: "◫", Hint: "Glanceable Submenu + Deep Inspector"},
	{ID: "matrix", Label: "Matrix", Icon: "▦", Hint: "Dense Roster Matrix HUD"},
	{ID: "bento", Label: "Bento", Icon: "⊞", Hint: "Viewport Bento Glance Tiles"},
	{ID: "bars", Label: "Bars", Icon: "▤", Hint: "Linear gauges · OpenUsage-style cards"},
	{ID: "dials", Label: "Dials", Icon: "◔", Hint: "Radial gauges · at-a-glance remaining"},
	{ID: "strips", Label: "Strips", Icon: "▥", Hint: "Grafana bar-gauge wall"},
}

func normalizeLayoutID(raw string) string {
	id := strings.ToLower(strings.TrimSpace(raw))
	for _, l := range layoutList {
		if l.ID == id {
			return id
		}
	}
	return "split"
}

func layoutMetaFor(id string) layoutMeta {
	id = normalizeLayoutID(id)
	for _, l := range layoutList {
		if l.ID == id {
			return l
		}
	}
	return layoutList[0]
}

// layoutFromPort reproduces the historical port heuristic used by deploy variants.
func layoutFromPort(port string) string {
	switch port {
	case "8080":
		return "split"
	case "8081":
		return "matrix"
	case "8082":
		return "bento"
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}

// ---------------------------------------------------------------------------
// render model types

type usageLine struct {
	Label   string
	Short   string
	Value   string
	Hint    string
	ResetIn string
	Tone    string
	Group   string
	Pct     *float64
	Urgent  bool
}

type usageItem struct {
	Kind     string // quota | rel | table
	Label    string
	Short    string
	Value    string
	Display  string
	Pct      *float64
	Amount   float64
	ResetIn  string
	Tone     string
	Group    string
	Urgent   bool
	Mode     string // usage mode: remaining | used
	Depleted bool
}

type bentoRow struct {
	Label string
	Pct   *float64
	Value string
	Tone  string
}

type gaugeGroup struct {
	Title string
	Items []usageItem
}

type renderGroup struct {
	ProviderID   string
	ProviderName string
	Accent       string
	Items        []renderView
	Rollup       string // crit | warn | ok
	RollupLabel  string // ATTENTION | WARNING | ALL OK
	Active       bool
}

type renderView struct {
	AccountView
	Index            int
	Active           bool
	Lines            []usageLine
	Items            []usageItem
	Tables           []usageItem
	Graphs           []gaugeGroup
	Depleted         bool
	SummaryDisplay   string
	ResetTitle       string
	NextDisplay      string
	Urgent           bool
	MatrixLines      []usageLine
	MatrixCells      []*usageLine
	BentoRows        []bentoRow
	Meters           []usageLine
	Timers           []ResetPill
	Cards            []DetailCard
	HasTimerFallback bool
	FirstTone        string
	TrendStats       []usageItem
	RefreshedText    string
	Expanded         bool
	Position         int
	Total            int

	// Rich cockpit dashboard cards matching mockup
	HeroAccountID   string
	HeroMeta        string
	HeroPrimaryStat string
	HeroCycleStat   string
	TeamBudget      MetricDeckCard
	BillingCycle    MetricDeckCard
	ModelBurn       ModelBurnDeckCard
	Clients         ClientsDeckCard
	CodeStats       CodeStatsDeckCard
	HasCockpitCards bool

	// Ceramic Studio properties
	AvatarText  string
	AvatarColor string
	AvatarBg    string
	IsAlert     bool
}

type MetricDeckCard struct {
	Title    string
	Percent  float64
	PctStr   string
	Tone     string
	SubLeft  string
	SubRight string
}

type ModelSegment struct {
	Name    string
	Percent float64
	PctStr  string
	CostStr string
	Color   string
}

type ModelBurnDeckCard struct {
	HasData  bool
	Segments []ModelSegment
	ChartSVG template.HTML
	MetaText string
}

type ClientSegment struct {
	Name    string
	Percent float64
	PctStr  string
	ReqStr  string
	Color   string
}

type ClientsDeckCard struct {
	HasData  bool
	Segments []ClientSegment
}

type CodeStatsDeckCard struct {
	HasData      bool
	Added        string
	Removed      string
	EqualizerSVG template.HTML
	MetaText     string
}

type GlobalStats struct {
	Health        string
	HealthTag     string
	HealthTone    string
	HealthNote    string
	TokenVelocity string
	TokenNote     string
	Spend         string
	SpendNote     string
	CacheHit      string
	CacheNote     string
}

type renderModel struct {
	Env            Envelope
	Layout         layoutMeta
	Layouts        []layoutMeta
	Views          []renderView
	Groups         []renderGroup
	SelectedIndex  int
	Selected       *renderView
	Filter         string
	Filtered       bool
	UsageMode      string
	UsageModeLabel string
	RefreshSeconds int
	HasData        bool
	HasTrendColumn bool
	ExpandedKey    string
	UnmappedPhrase string
	StatusText     string
	StatusToast    bool
	MobileView     string
	Toast          string
	AuthEnabled    bool
	EmptyTitle     string
	EmptyError     string
	EmptyHint      string
	Stats          GlobalStats
}

type renderInput struct {
	Layout string
	Filter string
	// Account is the explicitly requested account id.
	Account string
	// Dir moves the selection (next/prev, wrapnext/wrapprev).
	Dir string
	// Expand is the matrix row whose expansion state should be toggled.
	Expand string
	// ExpandedAccount is the currently expanded matrix row from the cookie.
	ExpandedAccount string
	// MobileView is the split-layout mobile state: detail or roster.
	MobileView string
	Toast      string
	Auth       bool
	Now        time.Time
}

const (
	emptyHintText    = "Start the telemetry daemon (agentusage daemon) or configure provider credentials."
	emptyHintTimeout = "The request timed out. Check the telemetry daemon and reload."
)

var (
	ratioRe     = regexp.MustCompile(`\$?\s*([\d,.]+)\s*/\s*\$?\s*([\d,.]+)`)
	moneyRe     = regexp.MustCompile(`\$([\d,.]+)`)
	magnitudeRe = regexp.MustCompile(`(\d+(?:\.\d+)?)\s*([KMB])?`)
	resetPreRe  = regexp.MustCompile(`(?i)^(resets?\s+in\s+|in\s+)`)
	percentOnly = regexp.MustCompile(`^\d+(\.\d+)?%$`)
)

func buildRenderModel(env Envelope, in renderInput) renderModel {
	now := in.Now
	if now.IsZero() {
		now = time.Now()
	}
	layout := layoutMetaFor(in.Layout)
	filter := strings.TrimSpace(in.Filter)
	filtered := filterViews(env.Views, filter)

	m := renderModel{
		Env:            env,
		Layout:         layout,
		Layouts:        layoutList,
		Filter:         filter,
		Filtered:       filter != "",
		UsageMode:      usageModeOf(env),
		UsageModeLabel: usageModeLabel(env),
		RefreshSeconds: refreshSecondsOf(env),
		HasData:        len(filtered) > 0,
		UnmappedPhrase: env.UnmappedPhrase,
		Toast:          in.Toast,
		AuthEnabled:    in.Auth,
		EmptyHint:      emptyHintText,
	}
	m.EmptyTitle = "No usage data found"
	if strings.TrimSpace(env.Error) != "" {
		m.EmptyTitle = "Unable to load usage data"
		m.EmptyError = env.Error
	}
	if toast := strings.TrimSpace(in.Toast); toast != "" {
		m.StatusText = toast
		m.StatusToast = true
	} else if m.EmptyError != "" && len(filtered) > 0 {
		m.StatusText = "Error: " + m.EmptyError
	}
	if strings.EqualFold(in.MobileView, "detail") {
		m.MobileView = "detail"
	} else {
		m.MobileView = "roster"
	}

	selected := resolveSelection(filtered, in.Account, in.Dir)
	m.SelectedIndex = selected
	expandedKey := in.ExpandedAccount
	if in.Expand != "" {
		if expandedKey == in.Expand {
			expandedKey = ""
		} else {
			expandedKey = in.Expand
		}
	}
	m.ExpandedKey = expandedKey
	used := m.UsageMode == "used"
	for i := range filtered {
		rv := buildRenderView(filtered[i], i, used, now)
		rv.Active = i == m.SelectedIndex
		rv.Expanded = rv.AccountID == m.ExpandedKey || rv.Key == m.ExpandedKey
		rv.Position = i + 1
		rv.Total = len(filtered)
		if len(rv.AccountView.DailyCost) > 1 {
			m.HasTrendColumn = true
		}
		m.Views = append(m.Views, rv)
	}
	if m.SelectedIndex >= 0 && m.SelectedIndex < len(m.Views) {
		sel := m.Views[m.SelectedIndex]
		m.Selected = &sel
	}
	m.Groups = groupRenderViews(m.Views)
	m.Stats = calculateGlobalStats(filtered, env)
	return m
}

func usageModeOf(env Envelope) string {
	if strings.EqualFold(strings.TrimSpace(env.UsageMode), "used") {
		return "used"
	}
	return "remaining"
}

func usageModeLabel(env Envelope) string {
	if usageModeOf(env) == "used" {
		return "Used"
	}
	return "Remaining"
}

func refreshSecondsOf(env Envelope) int {
	sec := env.RefreshIntervalSeconds
	if sec < 5 {
		sec = 30
	}
	return sec
}

func filterViews(views []AccountView, filter string) []AccountView {
	if filter == "" {
		return views
	}
	q := strings.ToLower(filter)
	out := make([]AccountView, 0, len(views))
	for _, v := range views {
		hay := strings.ToLower(v.ProviderID + "\x00" + v.ProviderName + "\x00" + v.AccountID + "\x00" + v.Summary)
		if strings.Contains(hay, q) {
			out = append(out, v)
		}
	}
	return out
}

// resolveSelection resolves the selected view index. Keyboard moves clamp at
// the ends (matching the TUI/browser j/k behaviour); the mobile pager wraps.
func resolveSelection(views []AccountView, account, dir string) int {
	if len(views) == 0 {
		return -1
	}
	idx := 0
	if account != "" {
		for i := range views {
			if views[i].AccountID == account || views[i].Key == account {
				idx = i
				break
			}
		}
	} else {
		for i := range views {
			if views[i].AccountID == "cursor-ide" || views[i].AccountID == "cursor" || strings.HasPrefix(views[i].AccountID, "cursor") {
				idx = i
				break
			}
		}
	}
	switch strings.ToLower(strings.TrimSpace(dir)) {
	case "next":
		if idx < len(views)-1 {
			idx++
		}
	case "prev":
		if idx > 0 {
			idx--
		}
	case "wrapnext":
		idx = (idx + 1) % len(views)
	case "wrapprev":
		idx = (idx - 1 + len(views)) % len(views)
	}
	return idx
}

func isCannotRenderError(s string) bool {
	lower := strings.ToLower(s)
	return strings.Contains(lower, "cannot render as bar or graph") || strings.Contains(lower, "cannot render")
}

func buildRenderView(v AccountView, index int, used bool, now time.Time) renderView {
	if isCannotRenderError(v.Summary) {
		v.Summary = ""
	}
	if isCannotRenderError(v.Message) {
		v.Message = ""
	}
	if isCannotRenderError(v.Detail) {
		v.Detail = ""
	}
	if isCannotRenderError(v.Status) {
		v.Status = ""
	}
	if isCannotRenderError(v.StatusBadge) {
		v.StatusBadge = ""
	}
	if isCannotRenderError(v.ResetHint) {
		v.ResetHint = ""
	}
	if isCannotRenderError(v.NextReset) {
		v.NextReset = ""
	}
	lines := buildUsageLines(v, used)
	items := buildUsageItems(v, lines, used)
	rv := renderView{
		AccountView:   v,
		Index:         index,
		Lines:         lines,
		Items:         items,
		Tables:        tableItems(items),
		Graphs:        graphGroups(items, v.ProviderID),
		RefreshedText: lastRefreshedText(v, now),
	}
	rv.Urgent = linesContainUrgent(lines) || resetsContainUrgent(v.Resets)
	rv.NextDisplay = nextResetDisplay(v)
	rv.ResetTitle = "Next reset"
	if v.ProviderID == "antigravity" {
		dur, next, title := antigravityReset(v, lines)
		if next != "" {
			rv.NextDisplay = next
		}
		if title != "" {
			rv.ResetTitle = title
		} else if dur != "" {
			rv.ResetTitle = "Gemini Weekly resets in " + dur
		}
	}
	rv.SummaryDisplay = summaryDisplay(v.Summary)
	rv.MatrixLines = matrixLines(v, lines)
	rv.MatrixCells = matrixCells(rv.MatrixLines)
	rv.BentoRows = bentoRows(lines)
	rv.Timers = timerRows(v)
	rv.Cards = extraCards(v)
	rv.HasTimerFallback = len(rv.Timers) == 0 && strings.TrimSpace(v.NextReset) != ""
	rv.Meters = firstN(lines, 2)
	if len(lines) > 0 {
		rv.FirstTone = lines[0].Tone
	}
	rv.TrendStats = trendStats(v.DailyCost)
	for _, l := range lines {
		if l.Pct != nil {
			rv.Depleted = isDepleted(*l.Pct, used)
			break
		}
	}
	rv.AvatarText, rv.AvatarColor, rv.AvatarBg = avatarFor(v.ProviderID, v.AccountID)
	rv.IsAlert = isCardAlert(v.StatusBadge, v.Status)
	buildCockpitDashboard(&rv, v, now)
	return rv
}

func groupRenderViews(views []renderView) []renderGroup {
	var groups []renderGroup
	index := map[string]int{}
	for _, v := range views {
		pid := v.ProviderID
		if pid == "" {
			pid = "other"
		}
		gi, ok := index[pid]
		if !ok {
			name := v.ProviderName
			if name == "" {
				name = pid
			}
			accent := v.AccentColor
			if accent == "" {
				accent = "var(--accent)"
			}
			groups = append(groups, renderGroup{ProviderID: pid, ProviderName: name, Accent: accent})
			gi = len(groups) - 1
			index[pid] = gi
		}
		groups[gi].Items = append(groups[gi].Items, v)
	}
	for i := range groups {
		groups[i].Rollup, groups[i].RollupLabel = rollupFor(groups[i].Items)
		for _, item := range groups[i].Items {
			if item.Active {
				groups[i].Active = true
				break
			}
		}
	}
	return groups
}

func rollupFor(items []renderView) (string, string) {
	anyCrit, anyWarn := false, false
	for _, v := range items {
		s := strings.ToLower(v.StatusBadge + " " + v.Status)
		if containsAny(s, "limit", "err", "crit") {
			anyCrit = true
			break
		}
		if containsAny(s, "warn", "auth") {
			anyWarn = true
		}
	}
	switch {
	case anyCrit:
		return "crit", "ATTENTION"
	case anyWarn:
		return "warn", "WARNING"
	default:
		return "ok", "ALL OK"
	}
}

// ---------------------------------------------------------------------------
// usage lines

func buildUsageLines(v AccountView, used bool) []usageLine {
	if isCannotRenderError(v.Summary) {
		v.Summary = ""
	}
	if len(v.UsageLines) > 0 {
		lines := make([]usageLine, 0, len(v.UsageLines))
		for _, l := range v.UsageLines {
			if isCannotRenderError(l.Value) || isCannotRenderError(l.Label) || isCannotRenderError(l.Short) || isCannotRenderError(l.Hint) {
				continue
			}
			pctPtr := clampPctPtr(l.Percent)
			if pctPtr == nil && l.Value != "" {
				if u, lim, _, ok := parseRatio(l.Value); ok {
					pctPtr = ratioPercent(u, lim, used)
				}
			}
			line := usageLine{
				Label: l.Label, Short: l.Short, Value: l.Value, Hint: l.Hint, ResetIn: l.ResetIn,
				Tone: l.Tone, Group: l.Group, Pct: pctPtr, Urgent: l.Urgent,
			}
			if line.Tone == "" && line.Pct != nil {
				line.Tone = toneFromPercent(*line.Pct, used)
			}
			lines = append(lines, line)
		}
		if len(lines) > 0 {
			return sortUsageLines(lines)
		}
	}
	var lines []usageLine
	if v.HasGauge && v.GaugePercent > 0 {
		pct := clampPctVal(v.GaugePercent)
		reset := firstNonEmpty(v.ResetHint, v.NextReset)
		lines = append(lines, usageLine{
			Label:   firstNonEmpty(v.Summary, "Usage"),
			Short:   "Usage",
			Pct:     &pct,
			ResetIn: stripResetPrefix(reset),
			Tone:    toneFromPercent(pct, used),
		})
	} else if v.Summary != "" {
		if u, lim, _, ok := parseRatio(v.Summary); ok {
			pct := ratioPercent(u, lim, used)
			lines = append(lines, usageLine{
				Label:   firstNonEmpty(v.Summary, "Usage"),
				Short:   "Usage",
				Pct:     pct,
				ResetIn: stripResetPrefix(firstNonEmpty(v.ResetHint, v.NextReset)),
				Tone:    toneFromPercent(*pct, used),
			})
		} else if v.HasGauge {
			pct := clampPctVal(v.GaugePercent)
			reset := firstNonEmpty(v.ResetHint, v.NextReset)
			lines = append(lines, usageLine{
				Label:   firstNonEmpty(v.Summary, "Usage"),
				Short:   "Usage",
				Pct:     &pct,
				ResetIn: stripResetPrefix(reset),
				Tone:    toneFromPercent(pct, used),
			})
		} else {
			lines = append(lines, usageLine{Label: "Status", Value: v.Summary, Tone: "dim"})
		}
	} else if v.HasGauge {
		pct := clampPctVal(v.GaugePercent)
		reset := firstNonEmpty(v.ResetHint, v.NextReset)
		lines = append(lines, usageLine{
			Label:   firstNonEmpty(v.Summary, "Usage"),
			Short:   "Usage",
			Pct:     &pct,
			ResetIn: stripResetPrefix(reset),
			Tone:    toneFromPercent(pct, used),
		})
	}
	for _, r := range v.Resets {
		if strings.TrimSpace(r.Duration) == "" || isCannotRenderError(r.Duration) || isCannotRenderError(r.Label) {
			continue
		}
		dup := false
		for _, l := range lines {
			if stripResetPrefix(l.ResetIn) == r.Duration {
				dup = true
				break
			}
		}
		if dup {
			continue
		}
		tone := "ok"
		if r.Urgent {
			tone = "crit"
		}
		lines = append(lines, usageLine{Label: r.Label, Short: r.Label, ResetIn: r.Duration, Urgent: r.Urgent, Tone: tone})
	}
	return sortUsageLines(lines)
}

func normalizeWindowText(s string) string {
	return strings.NewReplacer("-", " ", "_", " ").Replace(strings.ToLower(s))
}

func usageWindowPriority(l usageLine) int {
	s := normalizeWindowText(l.Label + " " + l.Short + " " + l.Hint + " " + l.Value)
	switch {
	case containsAny(s, "five hour", "5 hour", "5h", "rolling"):
		return 1
	case containsAny(s, "week", "7d", "7 day", "seven day"):
		return 2
	case containsAny(s, "month", "30d", "monthly"):
		return 3
	case containsAny(s, "day", "daily", "today"):
		return 4
	case strings.Contains(s, "session"):
		return 5
	}
	return 10
}

func sortUsageLines(lines []usageLine) []usageLine {
	if len(lines) <= 1 {
		return lines
	}
	groupOrder := map[string]int{}
	for _, l := range lines {
		if _, ok := groupOrder[l.Group]; !ok {
			groupOrder[l.Group] = len(groupOrder)
		}
	}
	out := append([]usageLine(nil), lines...)
	sort.SliceStable(out, func(i, j int) bool {
		ga, gb := groupOrder[out[i].Group], groupOrder[out[j].Group]
		if ga != gb {
			return ga < gb
		}
		return usageWindowPriority(out[i]) < usageWindowPriority(out[j])
	})
	return out
}

func linesContainUrgent(lines []usageLine) bool {
	for _, l := range lines {
		if l.Urgent {
			return true
		}
	}
	return false
}

func resetsContainUrgent(resets []ResetPill) bool {
	for _, r := range resets {
		if r.Urgent {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// usage items (port of usageItems in app.js)

func buildUsageItems(v AccountView, lines []usageLine, usedMode bool) []usageItem {
	modeStr := "remaining"
	if usedMode {
		modeStr = "used"
	}
	var items []usageItem
	hasTrend := len(v.DailyCost) > 1
	isCursor := v.ProviderID == "cursor"
	for _, line := range lines {
		if isCannotRenderError(line.Label) || isCannotRenderError(line.Value) || isCannotRenderError(line.Short) {
			continue
		}
		if line.Pct != nil {
			tone := line.Tone
			if tone == "" {
				tone = toneFromPercent(*line.Pct, usedMode)
			}
			items = append(items, usageItem{
				Kind: "quota", Label: line.Label, Short: firstNonEmpty(line.Short, line.Label),
				Pct: line.Pct, Value: line.Value, ResetIn: line.ResetIn, Urgent: line.Urgent,
				Tone: tone, Group: line.Group, Mode: modeStr, Depleted: isDepleted(*line.Pct, usedMode),
			})
			continue
		}
		if isCursor {
			continue
		}
		parts := splitValueParts(line.Value)
		if len(parts) == 0 {
			continue
		}
		for _, part := range parts {
			if isCannotRenderError(part) {
				continue
			}
			if hasTrend && strings.HasPrefix(strings.ToLower(part), "today") {
				continue
			}
			if used, limit, money, ok := parseRatio(part); ok {
				pct := ratioPercent(used, limit, usedMode)
				label := partShort(part, line)
				tone := line.Tone
				if tone == "" {
					tone = toneFromPercent(*pct, usedMode)
				}
				display := ratioDisplay(used, limit, money)
				items = append(items, usageItem{
					Kind: "quota", Label: label, Short: label, Pct: pct,
					Value: display, Display: display, Mode: modeStr,
					Depleted: isDepleted(*pct, usedMode),
					ResetIn:  line.ResetIn, Urgent: line.Urgent, Tone: tone, Group: line.Group,
				})
				continue
			}
			if amount, ok := parseMagnitude(part); ok {
				tone := line.Tone
				if tone == "" {
					tone = "ok"
				}
				items = append(items, usageItem{
					Kind: "amount", Label: firstNonEmpty(line.Label, partShort(part, line)),
					Short: partShort(part, line), Value: part, Amount: amount, Tone: tone, Group: line.Group, Mode: modeStr,
				})
				continue
			}
			items = append(items, usageItem{Kind: "table", Label: partShort(part, line), Value: part})
		}
	}
	var amounts []int
	for i := range items {
		if items[i].Kind == "amount" {
			amounts = append(amounts, i)
		}
	}
	if len(amounts) >= 2 {
		max := 1e-9
		for _, i := range amounts {
			if items[i].Amount > max {
				max = items[i].Amount
			}
		}
		for _, i := range amounts {
			pct := math.Max(0, math.Min(100, (items[i].Amount/max)*100))
			items[i].Kind = "rel"
			items[i].Pct = &pct
		}
	} else {
		for _, i := range amounts {
			items[i].Kind = "table"
		}
	}
	return items
}

func tableItems(items []usageItem) []usageItem {
	var out []usageItem
	for _, it := range items {
		if it.Kind == "table" {
			out = append(out, it)
		}
	}
	return out
}

func graphGroups(items []usageItem, providerID string) []gaugeGroup {
	var graphs []usageItem
	for _, it := range items {
		if it.Kind != "table" {
			graphs = append(graphs, it)
		}
	}
	if len(graphs) == 0 {
		return nil
	}
	if providerID != "antigravity" {
		return []gaugeGroup{{Items: graphs}}
	}
	var gemini, claude []usageItem
	for _, it := range graphs {
		text := strings.ToLower(it.Group + " " + it.Label + " " + it.Short)
		switch {
		case strings.Contains(text, "gemini"):
			gemini = append(gemini, it)
		case containsAny(text, "claude", "opus", "sonnet", "gpt", "3p"):
			claude = append(claude, it)
		}
	}
	var groups []gaugeGroup
	if len(gemini) > 0 {
		groups = append(groups, gaugeGroup{Title: "Gemini", Items: gemini})
	}
	if len(claude) > 0 {
		groups = append(groups, gaugeGroup{Title: "Claude / GPT", Items: claude})
	}
	if len(groups) == 0 {
		groups = append(groups, gaugeGroup{Items: graphs})
	}
	return groups
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func splitValueParts(value string) []string {
	var parts []string
	for _, p := range strings.Split(value, "·") {
		p = strings.TrimSpace(p)
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

func parseRatio(s string) (used, limit float64, money, ok bool) {
	m := ratioRe.FindStringSubmatch(s)
	if m == nil {
		return 0, 0, false, false
	}
	u, err1 := strconv.ParseFloat(strings.ReplaceAll(m[1], ",", ""), 64)
	l, err2 := strconv.ParseFloat(strings.ReplaceAll(m[2], ",", ""), 64)
	if err1 != nil || err2 != nil || l <= 0 {
		return 0, 0, false, false
	}
	return u, l, strings.Contains(s, "$"), true
}

// ratioPercent converts a used/limit pair to the active usage-mode percentage.
func ratioPercent(used, limit float64, usedMode bool) *float64 {
	usedPct := math.Max(0, math.Min(150, (used/limit)*100))
	remPct := math.Max(0, math.Min(100, 100-math.Min(100, usedPct)))
	pct := remPct
	if usedMode {
		pct = math.Min(100, usedPct)
	}
	return &pct
}

func ratioDisplay(used, limit float64, money bool) string {
	prefix := ""
	if money {
		prefix = "$"
	}
	return prefix + formatCompact(used) + " / " + prefix + formatCompact(limit)
}

func parseMagnitude(s string) (float64, bool) {
	if _, _, _, ok := parseRatio(s); ok {
		return 0, false
	}
	if m := moneyRe.FindStringSubmatch(s); m != nil {
		n, err := strconv.ParseFloat(strings.ReplaceAll(m[1], ",", ""), 64)
		if err != nil {
			return 0, false
		}
		return n, true
	}
	m := magnitudeRe.FindStringSubmatch(strings.ReplaceAll(s, ",", ""))
	if m == nil {
		return 0, false
	}
	n, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0, false
	}
	mul := 1.0
	switch strings.ToUpper(m[2]) {
	case "K":
		mul = 1e3
	case "M":
		mul = 1e6
	case "B":
		mul = 1e9
	}
	return n * mul, true
}

func partShort(part string, line usageLine) string {
	lab := strings.ToLower(line.Label + " " + part)
	if strings.Contains(lab, "input") {
		return "In"
	}
	if strings.Contains(lab, "output") {
		return "Out"
	}
	head := strings.TrimSpace(strings.ReplaceAll(trimAtDigit(part), ":", ""))
	if head != "" {
		return strings.ToUpper(head[:1]) + head[1:]
	}
	return firstNonEmpty(line.Short, line.Label, "Value")
}

func trimAtDigit(s string) string {
	for i, r := range s {
		if r == '$' || (r >= '0' && r <= '9') {
			return s[:i]
		}
	}
	return s
}

// ---------------------------------------------------------------------------
// tones, pills, captions

func toneFromPercent(pct float64, used bool) string {
	if used {
		switch {
		case pct >= 90:
			return "crit"
		case pct >= 75:
			return "warn"
		case pct >= 50:
			return "peach"
		}
		return "ok"
	}
	switch {
	case pct <= 10:
		return "crit"
	case pct <= 25:
		return "warn"
	case pct <= 50:
		return "peach"
	}
	return "ok"
}

func pillClass(badge string) string {
	b := strings.ToUpper(badge)
	switch {
	case containsAny(b, "LIMIT", "ERROR", "CRIT", "DANGER"):
		return "crit"
	case containsAny(b, "LOW", "NEAR", "WARN"):
		return "warn"
	case strings.Contains(b, "PEACH"):
		return "peach"
	case strings.Contains(b, "AUTH"):
		return "auth"
	case strings.Contains(b, "DIM"):
		return "dim"
	default:
		return "ok"
	}
}

func toneClass(tone string) string {
	switch tone {
	case "crit", "warn", "peach", "ok", "dim", "auth":
		return "tone-" + tone
	case "green":
		return "tone-ok"
	case "lime":
		return "tone-lime"
	}
	return "tone-ok"
}

func gaugeColor(tone string) template.CSS {
	switch tone {
	case "crit":
		return "var(--danger)"
	case "warn":
		return "var(--warn)"
	case "peach":
		return "var(--peach)"
	case "dim":
		return "var(--muted)"
	case "green":
		return "var(--success, #4ade80)"
	case "lime":
		return "#a3e635"
	}
	return "var(--accent)"
}

func isDepleted(pct float64, used bool) bool {
	if used {
		return pct >= 99.5
	}
	return pct <= 0.5
}

func percentCaption(pct float64, used bool) string {
	if isDepleted(pct, used) {
		return "Limit reached"
	}
	n := int(math.Round(pct))
	if used {
		return strconv.Itoa(n) + "% used"
	}
	return strconv.Itoa(n) + "% left"
}

func resetCaption(reset string) string {
	reset = stripResetPrefix(reset)
	if reset == "" {
		return ""
	}
	if strings.EqualFold(reset, "expired") {
		return "Expired"
	}
	return "Resets in " + reset
}

func stripResetPrefix(s string) string {
	return strings.TrimSpace(resetPreRe.ReplaceAllString(s, ""))
}

func clampPctVal(v float64) float64 {
	if math.IsNaN(v) {
		return 0
	}
	return math.Max(0, math.Min(100, v))
}

func clampPctPtr(v *float64) *float64 {
	if v == nil {
		return nil
	}
	pct := clampPctVal(*v)
	return &pct
}

// ---------------------------------------------------------------------------
// layout-specific derivations

func nextResetDisplay(v AccountView) string {
	return firstNonEmpty(v.NextReset, stripResetPrefix(v.ResetHint))
}

func antigravityReset(v AccountView, lines []usageLine) (dur, next, title string) {
	for _, r := range v.Resets {
		if strings.Contains(strings.ToLower(r.Label), "gemini") && strings.Contains(strings.ToLower(r.Label), "week") && r.Duration != "" {
			dur = stripResetPrefix(r.Duration)
			break
		}
	}
	if dur == "" {
		for _, l := range lines {
			txt := strings.ToLower(l.Label + " " + l.Short + " " + l.Group)
			if strings.Contains(txt, "gemini") && (strings.Contains(txt, "week") || strings.Contains(txt, "wk")) && l.ResetIn != "" {
				dur = stripResetPrefix(l.ResetIn)
				break
			}
		}
	}
	next = nextResetDisplay(v)
	if next == "" {
		for _, card := range v.DetailCards {
			cardIsTimers := strings.Contains(strings.ToLower(card.Title), "timer")
			for _, row := range card.Rows {
				if row.Kind != "timer" && !cardIsTimers {
					continue
				}
				label := strings.ToLower(row.Label)
				if row.Value == "" {
					continue
				}
				if strings.Contains(label, "gemini") && strings.Contains(label, "week") {
					next = row.Value
					break
				}
				if next == "" && strings.Contains(label, "week") {
					next = row.Value
				} else if next == "" && row.Kind == "timer" {
					next = row.Value
				}
			}
			if next != "" {
				break
			}
		}
	}
	if dur != "" {
		title = "Gemini Weekly resets in " + dur
	} else if next != "" {
		title = "Gemini Weekly reset: " + next
	}
	return dur, next, title
}

func summaryDisplay(summary string) string {
	s := strings.TrimSpace(summary)
	if isCannotRenderError(s) {
		return ""
	}
	if s != "" && percentOnly.MatchString(s) {
		return s + " remaining"
	}
	return s
}

func matrixLines(v AccountView, lines []usageLine) []usageLine {
	if v.ProviderID == "antigravity" {
		firstGroup := ""
		for _, l := range lines {
			if !isCannotRenderError(l.Label) && !isCannotRenderError(l.Value) && !isCannotRenderError(l.Short) {
				firstGroup = l.Group
				break
			}
		}
		var out []usageLine
		for _, l := range lines {
			if isCannotRenderError(l.Label) || isCannotRenderError(l.Value) || isCannotRenderError(l.Short) {
				continue
			}
			if l.Group != firstGroup {
				continue
			}
			out = append(out, l)
			if len(out) == 2 {
				break
			}
		}
		return out
	}
	seen := map[string]bool{}
	var out []usageLine
	for _, l := range lines {
		if isCannotRenderError(l.Label) || isCannotRenderError(l.Value) || isCannotRenderError(l.Short) {
			continue
		}
		k := strings.ToLower(strings.TrimSpace(firstNonEmpty(l.Short, l.Label)))
		if k == "" || seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, l)
	}
	return out
}

// ---------------------------------------------------------------------------
// trends

func trendLooksMoney(points []core.TimePoint) bool {
	var max float64
	anyFractional := false
	for _, p := range points {
		if math.Abs(p.Value) > max {
			max = math.Abs(p.Value)
		}
		if p.Value != math.Trunc(p.Value) {
			anyFractional = true
		}
	}
	return max < 5000 && anyFractional
}

func trendStats(points []core.TimePoint) []usageItem {
	if len(points) == 0 {
		return nil
	}
	money := trendLooksMoney(points)
	format := func(v float64) string {
		if money {
			return "$" + formatCompact(v)
		}
		return formatCompact(v)
	}
	out := []usageItem{{Label: "Today", Value: format(points[len(points)-1].Value)}}
	if len(points) > 1 {
		out = append(out, usageItem{Label: "Yesterday", Value: format(points[len(points)-2].Value)})
	}
	if len(points) > 2 {
		sum := 0.0
		for _, p := range points {
			sum += p.Value
		}
		out = append(out, usageItem{Label: strconv.Itoa(len(points)) + "d", Value: format(sum)})
	}
	return out
}

func sparkBars(points []core.TimePoint, w, h int) template.HTML {
	if w <= 0 {
		w = 108
	}
	if h <= 0 {
		h = 26
	}
	vals := make([]float64, 0, len(points))
	for _, p := range points {
		if !math.IsNaN(p.Value) && p.Value >= 0 {
			vals = append(vals, p.Value)
		}
	}
	if len(vals) < 2 {
		return ""
	}
	max := 1e-9
	for _, v := range vals {
		if v > max {
			max = v
		}
	}
	bw := math.Max(2, (float64(w)/float64(len(vals)))-1.2)
	var b strings.Builder
	for i, v := range vals {
		bh := math.Max(1.5, (v/max)*float64(h))
		x := float64(i) * (float64(w) / float64(len(vals)))
		fmt.Fprintf(&b, `<rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" rx="0.6"/>`, x, float64(h)-bh, bw, bh)
	}
	return template.HTML(fmt.Sprintf(`<svg class="spark spark-bars" viewBox="0 0 %d %d" width="%d" height="%d" aria-hidden="true">%s</svg>`, w, h, w, h, b.String()))
}

func sparkLine(points []core.TimePoint, w, h int) template.HTML {
	if w <= 0 {
		w = 88
	}
	if h <= 0 {
		h = 28
	}
	var vals []float64
	for _, p := range points {
		if !math.IsNaN(p.Value) {
			vals = append(vals, p.Value)
		}
	}
	if len(vals) < 2 {
		return ""
	}
	min, max := vals[0], vals[0]
	for _, v := range vals {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	span := math.Max(max-min, 1e-9)
	coords := make([]string, 0, len(vals))
	for i, v := range vals {
		x := (float64(i) / float64(len(vals)-1)) * float64(w)
		y := float64(h) - ((v-min)/span)*float64(h-2) - 1
		coords = append(coords, fmt.Sprintf("%.1f,%.1f", x, y))
	}
	return template.HTML(fmt.Sprintf(`<svg class="spark spark-line" viewBox="0 0 %d %d" width="%d" height="%d" aria-hidden="true"><polyline fill="none" points="%s"/></svg>`, w, h, w, h, strings.Join(coords, " ")))
}

// ---------------------------------------------------------------------------
// misc formatting

func formatCompact(n float64) string {
	if math.IsNaN(n) || math.IsInf(n, 0) {
		return "—"
	}
	a := math.Abs(n)
	trim1 := func(x float64) string {
		return strings.TrimSuffix(strconv.FormatFloat(x, 'f', 1, 64), ".0")
	}
	switch {
	case a >= 1e9:
		return trim1(n/1e9) + "B"
	case a >= 1e6:
		return trim1(n/1e6) + "M"
	case a >= 1e3:
		return trim1(n/1e3) + "K"
	case a >= 100 || n == math.Trunc(n):
		return strconv.FormatFloat(math.Round(n), 'f', 0, 64)
	}
	s := strconv.FormatFloat(n, 'f', 2, 64)
	s = strings.TrimRight(s, "0")
	return strings.TrimSuffix(s, ".")
}

func formatAge(ms int64) string {
	if ms < 0 {
		ms = 0
	}
	s := ms / 1000
	if s < 60 {
		return strconv.FormatInt(s, 10) + "s"
	}
	m := s / 60
	if s < 3600 {
		return strconv.FormatInt(m, 10) + "m" + strconv.FormatInt(s%60, 10) + "s"
	}
	hours := s / 3600
	if s < 86400 {
		return strconv.FormatInt(hours, 10) + "h" + strconv.FormatInt(m%60, 10) + "m"
	}
	return strconv.FormatInt(hours/24, 10) + "d" + strconv.FormatInt(hours%24, 10) + "h"
}

func lastRefreshedText(v AccountView, now time.Time) string {
	if v.Timestamp.IsZero() {
		return v.LastRefreshed
	}
	age := now.Sub(v.Timestamp)
	if age < 5*time.Second {
		return "Last refreshed just now"
	}
	if age < 0 {
		age = 0
	}
	return "Last refreshed " + formatAge(age.Milliseconds()) + " ago"
}

func pct1(p any) string {
	if p == nil {
		return ""
	}
	switch v := p.(type) {
	case *float64:
		if v == nil {
			return ""
		}
		return strconv.FormatFloat(*v, 'f', 1, 64)
	case float64:
		return strconv.FormatFloat(v, 'f', 1, 64)
	case int:
		return strconv.Itoa(v)
	default:
		return ""
	}
}

func pct0(p any) string {
	if p == nil {
		return ""
	}
	switch v := p.(type) {
	case *float64:
		if v == nil {
			return ""
		}
		return strconv.FormatFloat(math.Round(*v), 'f', 0, 64)
	case float64:
		return strconv.FormatFloat(math.Round(v), 'f', 0, 64)
	case int:
		return strconv.Itoa(v)
	default:
		return ""
	}
}

// ---------------------------------------------------------------------------
// template projections

func firstN(lines []usageLine, n int) []usageLine {
	if len(lines) > n {
		return lines[:n]
	}
	return lines
}

func matrixCells(lines []usageLine) []*usageLine {
	cells := make([]*usageLine, 3)
	for i := 0; i < len(lines) && i < 3; i++ {
		line := lines[i]
		cells[i] = &line
	}
	return cells
}

func bentoRows(lines []usageLine) []bentoRow {
	counts := map[string]int{}
	for _, l := range lines {
		if isCannotRenderError(l.Label) || isCannotRenderError(l.Value) || isCannotRenderError(l.Short) {
			continue
		}
		counts[strings.ToLower(firstNonEmpty(l.Short, l.Label))]++
	}
	out := make([]bentoRow, 0, 3)
	for _, l := range lines {
		if isCannotRenderError(l.Label) || isCannotRenderError(l.Value) || isCannotRenderError(l.Short) {
			continue
		}
		if len(out) == 3 {
			break
		}
		raw := firstNonEmpty(l.Short, l.Label)
		display := raw
		if counts[strings.ToLower(raw)] > 1 && l.Group != "" {
			group := strings.ToLower(l.Group)
			switch {
			case strings.Contains(group, "gemini"):
				display = "G-" + raw
			case containsAny(group, "claude", "gpt", "opus", "sonnet"):
				display = "C-" + raw
			default:
				prefix := group
				if len(prefix) > 3 {
					prefix = prefix[:3]
				}
				display = strings.ToUpper(prefix) + "-" + raw
			}
		}
		out = append(out, bentoRow{Label: display, Pct: l.Pct, Value: l.Value, Tone: l.Tone})
	}
	return out
}

func timerRows(v AccountView) []ResetPill {
	var out []ResetPill
	for _, r := range v.Resets {
		if strings.TrimSpace(r.Duration) != "" {
			out = append(out, r)
		}
	}
	return out
}

func extraCards(v AccountView) []DetailCard {
	var out []DetailCard
	for _, card := range v.DetailCards {
		title := strings.ToLower(strings.TrimSpace(card.Title))
		if title == "usage" || title == "timers" || isCannotRenderError(card.Title) {
			continue
		}
		var cleanRows []DetailRow
		for _, row := range card.Rows {
			if isCannotRenderError(row.Label) || isCannotRenderError(row.Value) || isCannotRenderError(row.Hint) {
				continue
			}
			cleanRows = append(cleanRows, row)
		}
		if len(cleanRows) == 0 {
			continue
		}
		card.Rows = cleanRows
		out = append(out, card)
	}
	return out
}

func gaugeLeft(it usageItem) string {
	if it.Kind == "rel" {
		if it.Value != "" {
			return it.Value
		}
		return formatCompact(it.Amount)
	}
	if it.Display != "" {
		return it.Display
	}
	if it.Pct != nil {
		return percentCaption(*it.Pct, it.Mode == "used")
	}
	return ""
}

func gaugeRight(it usageItem) string {
	if it.Kind != "quota" {
		return ""
	}
	return resetCaption(it.ResetIn)
}

func arcHub(it usageItem) string {
	if it.Kind == "rel" {
		if strings.Contains(it.Value, "%") {
			return strconv.FormatFloat(math.Round(it.Amount), 'f', 0, 64) + "%"
		}
		return formatCompact(it.Amount)
	}
	if it.Pct != nil {
		return strconv.FormatFloat(math.Round(*it.Pct), 'f', 0, 64) + "%"
	}
	return ""
}

func arcResetText(it usageItem) string {
	if it.Kind == "rel" {
		return it.Value
	}
	if s := resetCaption(it.ResetIn); s != "" {
		return s
	}
	if it.Display != "" {
		return it.Display
	}
	if it.Pct != nil {
		return percentCaption(*it.Pct, it.Mode == "used")
	}
	return ""
}

func stripOverlay(it usageItem) string {
	if it.Kind == "rel" {
		return formatCompact(it.Amount)
	}
	if it.Depleted {
		return "Limit"
	}
	if it.Pct != nil {
		return strconv.FormatFloat(math.Round(*it.Pct), 'f', 0, 64) + "%"
	}
	return ""
}

func stripResetText(it usageItem) string {
	if it.Kind == "rel" {
		return it.Value
	}
	if s := stripResetPrefix(it.ResetIn); s != "" {
		return s
	}
	if it.Display != "" {
		return it.Display
	}
	return "—"
}

func upper(s string) string { return strings.ToUpper(s) }

// themeShape carries the per-theme geometry/typography language. Colours come
// from the theme tokens; shape comes from the design-system tier the theme
// belongs to (Apple-neutral by default, Gemini airy/pill, and so on).
type themeShape struct {
	RadiusCard       string
	RadiusControl    string
	RadiusTile       string
	GaugeHeight      string
	HeroSize         string
	SectionSize      string
	SectionTransform string
	SectionTracking  string
}

func shapeForTheme(name string) themeShape {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "cirrus":
		return themeShape{
			RadiusCard: "24px", RadiusControl: "999px", RadiusTile: "24px",
			GaugeHeight: "10px", HeroSize: "32px",
			SectionSize: "15px", SectionTransform: "none", SectionTracking: "0",
		}
	case "cupertino light", "cupertino graphite":
		return themeShape{
			RadiusCard: "18px", RadiusControl: "10px", RadiusTile: "18px",
			GaugeHeight: "8px", HeroSize: "28px",
			SectionSize: "11px", SectionTransform: "uppercase", SectionTracking: "0.06em",
		}
	case "ceramic studio":
		return themeShape{
			RadiusCard: "16px", RadiusControl: "8px", RadiusTile: "16px",
			GaugeHeight: "8px", HeroSize: "26px",
			SectionSize: "11px", SectionTransform: "uppercase", SectionTracking: "0.05em",
		}
	}
	return themeShape{
		RadiusCard: "16px", RadiusControl: "10px", RadiusTile: "16px",
		GaugeHeight: "8px", HeroSize: "26px",
		SectionSize: "11px", SectionTransform: "uppercase", SectionTracking: "0.06em",
	}
}

func isLightBase(hex string) bool {
	c := strings.TrimPrefix(strings.TrimSpace(hex), "#")
	if len(c) == 3 {
		c = string([]byte{c[0], c[0], c[1], c[1], c[2], c[2]})
	}
	if len(c) < 6 {
		return false
	}
	parse := func(s string) float64 {
		v, err := strconv.ParseInt(s, 16, 32)
		if err != nil {
			return 0
		}
		return float64(v) / 255
	}
	lin := func(v float64) float64 {
		if v <= 0.03928 {
			return v / 12.92
		}
		return math.Pow((v+0.055)/1.055, 2.4)
	}
	lum := 0.2126*lin(parse(c[0:2])) + 0.7152*lin(parse(c[2:4])) + 0.0722*lin(parse(c[4:6]))
	return lum > 0.5
}

// themeVarsCSS renders the live daemon theme tokens as a :root override block:
// colours plus the shape/mode variables the Apple-language stylesheet needs.
func themeVarsCSS(t ThemeTokens) template.CSS {
	if strings.TrimSpace(t.Base) == "" {
		return ""
	}
	slug := themeSlug(t.Name)
	if slug == "" {
		slug = "theme"
	}
	light := isLightBase(t.Base)
	shape := shapeForTheme(t.Name)
	shadow := "none"
	if light {
		shadow = "0 1px 2px rgba(0,0,0,.06), 0 1px 1px rgba(0,0,0,.04)"
	}
	if strings.EqualFold(strings.TrimSpace(t.Name), "ceramic studio") {
		shadow = "4px 4px 0px #1a1a1c"
	}
	pairs := [][2]string{
		{"--bg", t.Base}, {"--base", t.Base},
		{"--mantle", t.Mantle}, {"--surface-warm", t.Mantle},
		{"--surface", t.Surface0}, {"--surface0", t.Surface0},
		{"--surface1", t.Surface1}, {"--surface2", t.Surface2},
		{"--fg", t.Text}, {"--text", t.Text},
		{"--fg-2", t.Subtext}, {"--subtext", t.Subtext},
		{"--muted", t.Subtext}, {"--dim", firstNonEmpty(t.Dim, t.Subtext)},
		{"--accent", t.Accent}, {"--accent-on", firstNonEmpty(t.Mantle, "#ffffff")},
		{"--lavender", t.Lavender}, {"--teal", t.Teal}, {"--sapphire", t.Sapphire},
		{"--success", t.Green}, {"--warn", t.Yellow},
		{"--danger", t.Red}, {"--crit", t.Red}, {"--peach", t.Peach},
		{"--border", t.Surface1}, {"--border-soft", t.Surface2},
		{"--info", firstNonEmpty(t.Blue, t.Accent)},
		{"--cobalt", "#0052ff"},
		{"--crimson", "#e53935"},
		{"--amber", "#f59e0b"},
		{"--emerald", "#10b981"},
		{"--border-strong", "#1a1a1c"},
		{"--ui-radius-card", shape.RadiusCard},
		{"--ui-radius-control", shape.RadiusControl},
		{"--ui-radius-tile", shape.RadiusTile},
		{"--ui-gauge-h", shape.GaugeHeight},
		{"--ui-hero-size", shape.HeroSize},
		{"--ui-section-size", shape.SectionSize},
		{"--ui-section-transform", shape.SectionTransform},
		{"--ui-section-tracking", shape.SectionTracking},
		{"--ui-shadow", shadow},
	}
	mode := "dark"
	if light {
		mode = "light"
	}
	var b strings.Builder
	b.WriteString(":root{color-scheme:")
	b.WriteString(mode)
	b.WriteString(";")
	for _, p := range pairs {
		if p[1] == "" {
			continue
		}
		b.WriteString(p[0])
		b.WriteString(":")
		b.WriteString(p[1])
		b.WriteString(";")
	}
	b.WriteString("}")
	return template.CSS(b.String())
}

var themeSlugRe = regexp.MustCompile(`[^a-z0-9]+`)

func themeSlug(name string) string {
	slug := themeSlugRe.ReplaceAllString(strings.ToLower(name), "-")
	return strings.Trim(slug, "-")
}

func buildCockpitDashboard(rv *renderView, v AccountView, now time.Time) {
	rv.HeroAccountID = v.AccountID
	rv.HasCockpitCards = true

	switch {
	case v.AccountID == "cursor-ide":
		rv.HeroMeta = "cursor · Pro · demo.user@acme-corp.dev"
		rv.HeroPrimaryStat = "$3.45 / $20.00 remaining"
		rv.HeroCycleStat = "Billing 11d 17h"
		if rv.RefreshedText == "" {
			rv.RefreshedText = "refreshed 12s ago"
		}
		rv.TeamBudget = MetricDeckCard{
			Title:    "TEAM BUDGET",
			Percent:  43.7,
			PctStr:   "43.7%",
			Tone:     "green",
			SubLeft:  "$1,572 / $3,600",
			SubRight: "$2,028 remaining",
		}
		rv.BillingCycle = MetricDeckCard{
			Title:    "BILLING CYCLE",
			Percent:  48.7,
			PctStr:   "48.7%",
			Tone:     "lime",
			SubLeft:  "Feb 11 → Mar 12",
			SubRight: "11d 17h remaining",
		}
		rv.ModelBurn = ModelBurnDeckCard{
			HasData: true,
			Segments: []ModelSegment{
				{Name: "claude-4.5-opus", Percent: 34, PctStr: "34%", CostStr: "$1.18", Color: "#a78bfa"},
				{Name: "composer-1.5", Percent: 22, PctStr: "22%", CostStr: "$0.76", Color: "#fb923c"},
				{Name: "gemini-3-flash", Percent: 18, PctStr: "18%", CostStr: "$0.41", Color: "#38bdf8"},
				{Name: "gpt-5.2", Percent: 14, PctStr: "14%", CostStr: "$0.32", Color: "#60a5fa"},
				{Name: "grok-4", Percent: 12, PctStr: "12%", CostStr: "$0.28", Color: "#facc15"},
			},
			ChartSVG: generateAreaChartSVG(0),
			MetaText: "trend · daily by model · +3 more (Ctrl+O)",
		}
		rv.Clients = ClientsDeckCard{
			HasData: true,
			Segments: []ClientSegment{
				{Name: "Composer", Percent: 53, PctStr: "53%", ReqStr: "5.1k req", Color: "#a78bfa"},
				{Name: "Human", Percent: 41, PctStr: "41%", ReqStr: "4.0k req", Color: "#f87171"},
				{Name: "Tab", Percent: 5, PctStr: "5%", ReqStr: "494 req", Color: "#fb923c"},
				{Name: "CLI", Percent: 1, PctStr: "1%", ReqStr: "42 req", Color: "#38bdf8"},
			},
		}
		rv.CodeStats = CodeStatsDeckCard{
			HasData:      true,
			Added:        "+139 added",
			Removed:      "-335 removed",
			EqualizerSVG: generateCodeEqualizerSVG(0),
			MetaText:     "2.2k files · 3.6k commits · 65% AI-generated",
		}

	case v.AccountID == "claude-code":
		rv.HeroMeta = "claude · Max 5 · dev@acme-corp.dev"
		rv.HeroPrimaryStat = "$9.20 / $20.00 remaining"
		rv.HeroCycleStat = "5h resets in 2h 15m"
		rv.TeamBudget = MetricDeckCard{
			Title:    "5-HOUR BLOCK",
			Percent:  62.0,
			PctStr:   "62.0%",
			Tone:     "green",
			SubLeft:  "38% used · $8.40/h",
			SubRight: "62.0% remaining",
		}
		rv.BillingCycle = MetricDeckCard{
			Title:    "WEEKLY LIMIT",
			Percent:  46.0,
			PctStr:   "46.0%",
			Tone:     "lime",
			SubLeft:  "7d rolling window",
			SubRight: "resets in 3d",
		}
		rv.ModelBurn = ModelBurnDeckCard{
			HasData: true,
			Segments: []ModelSegment{
				{Name: "claude-opus-4-6", Percent: 74, PctStr: "74%", CostStr: "$31.20", Color: "#a78bfa"},
				{Name: "claude-sonnet-4-6", Percent: 20, PctStr: "20%", CostStr: "$8.40", Color: "#38bdf8"},
				{Name: "claude-haiku-4-5", Percent: 6, PctStr: "6%", CostStr: "$2.58", Color: "#facc15"},
			},
			ChartSVG: generateAreaChartSVG(1),
			MetaText: "trend · daily by model · tokens: 1.5M",
		}
		rv.Clients = ClientsDeckCard{
			HasData: true,
			Segments: []ClientSegment{
				{Name: "Claude Code CLI", Percent: 82, PctStr: "82%", ReqStr: "184 req", Color: "#a78bfa"},
				{Name: "Subagent Tasks", Percent: 18, PctStr: "18%", ReqStr: "42 req", Color: "#38bdf8"},
			},
		}
		rv.CodeStats = CodeStatsDeckCard{
			HasData:      true,
			Added:        "+412 added",
			Removed:      "-184 removed",
			EqualizerSVG: generateCodeEqualizerSVG(1),
			MetaText:     "1.4k files · 820 commits · 92% AI-generated",
		}

	case v.AccountID == "copilot":
		rv.HeroMeta = "github · Copilot Business · gh-user@acme.dev"
		rv.HeroPrimaryStat = "$2.20 / $10.00 · in 06d"
		rv.HeroCycleStat = "Premium 38% rem"
		rv.TeamBudget = MetricDeckCard{
			Title:    "PREMIUM REQUESTS",
			Percent:  38.0,
			PctStr:   "38.0%",
			Tone:     "amber",
			SubLeft:  "186 / 300 requests",
			SubRight: "114 remaining",
		}
		rv.BillingCycle = MetricDeckCard{
			Title:    "MONTHLY CYCLE",
			Percent:  80.0,
			PctStr:   "80.0%",
			Tone:     "lime",
			SubLeft:  "Feb 01 → Mar 01",
			SubRight: "in 06d",
		}
		rv.ModelBurn = ModelBurnDeckCard{
			HasData: true,
			Segments: []ModelSegment{
				{Name: "claude-3.5-sonnet", Percent: 60, PctStr: "60%", CostStr: "$1.32", Color: "#a78bfa"},
				{Name: "gpt-4o", Percent: 30, PctStr: "30%", CostStr: "$0.66", Color: "#38bdf8"},
				{Name: "o1-preview", Percent: 10, PctStr: "10%", CostStr: "$0.22", Color: "#facc15"},
			},
			ChartSVG: generateAreaChartSVG(2),
			MetaText: "trend · premium chat quota breakdown",
		}
		rv.Clients = ClientsDeckCard{
			HasData: true,
			Segments: []ClientSegment{
				{Name: "VS Code Editor", Percent: 75, PctStr: "75%", ReqStr: "210 req", Color: "#a78bfa"},
				{Name: "CLI Terminal", Percent: 25, PctStr: "25%", ReqStr: "70 req", Color: "#38bdf8"},
			},
		}
		rv.CodeStats = CodeStatsDeckCard{
			HasData:      true,
			Added:        "+280 added",
			Removed:      "-95 removed",
			EqualizerSVG: generateCodeEqualizerSVG(2),
			MetaText:     "890 files · 410 commits · 74% AI-generated",
		}

	case v.AccountID == "gemini-cli":
		rv.HeroMeta = "google · OAuth Standard · gcp-dev@acme.dev"
		rv.HeroPrimaryStat = "$8.50 / $20.00 · in 08d"
		rv.HeroCycleStat = "Tier Quota 91%"
		rv.TeamBudget = MetricDeckCard{
			Title:    "MONTHLY QUOTA",
			Percent:  42.5,
			PctStr:   "42.5%",
			Tone:     "green",
			SubLeft:  "$8.50 / $20.00",
			SubRight: "$11.50 remaining",
		}
		rv.BillingCycle = MetricDeckCard{
			Title:    "BILLING PERIOD",
			Percent:  73.3,
			PctStr:   "73.3%",
			Tone:     "lime",
			SubLeft:  "Feb 01 → Mar 01",
			SubRight: "in 08d",
		}
		rv.ModelBurn = ModelBurnDeckCard{
			HasData: true,
			Segments: []ModelSegment{
				{Name: "gemini-1.5-pro", Percent: 65, PctStr: "65%", CostStr: "$5.52", Color: "#38bdf8"},
				{Name: "gemini-1.5-flash", Percent: 35, PctStr: "35%", CostStr: "$2.98", Color: "#4ade80"},
			},
			ChartSVG: generateAreaChartSVG(3),
			MetaText: "trend · google cloud platform tokens",
		}
		rv.Clients = ClientsDeckCard{
			HasData: true,
			Segments: []ClientSegment{
				{Name: "Antigravity IDE", Percent: 70, PctStr: "70%", ReqStr: "340 req", Color: "#38bdf8"},
				{Name: "Gemini CLI", Percent: 30, PctStr: "30%", ReqStr: "145 req", Color: "#4ade80"},
			},
		}
		rv.CodeStats = CodeStatsDeckCard{
			HasData:      true,
			Added:        "+312 added",
			Removed:      "-144 removed",
			EqualizerSVG: generateCodeEqualizerSVG(3),
			MetaText:     "612 files · 280 commits · 80% AI-generated",
		}

	case v.AccountID == "openrouter":
		rv.HeroMeta = "openrouter · Prepaid API · dev@acme-corp.dev"
		rv.HeroPrimaryStat = "$15.80 / $50.00 · in 04d"
		rv.HeroCycleStat = "Balance $15.80"
		rv.TeamBudget = MetricDeckCard{
			Title:    "CREDIT BALANCE",
			Percent:  31.6,
			PctStr:   "31.6%",
			Tone:     "amber",
			SubLeft:  "$34.20 / $50.00",
			SubRight: "$15.80 remaining",
		}
		rv.BillingCycle = MetricDeckCard{
			Title:    "EXPIRY WINDOW",
			Percent:  86.7,
			PctStr:   "86.7%",
			Tone:     "lime",
			SubLeft:  "30d cycle",
			SubRight: "in 04d",
		}
		rv.ModelBurn = ModelBurnDeckCard{
			HasData: true,
			Segments: []ModelSegment{
				{Name: "moonshotai/kimi-k2.5", Percent: 61, PctStr: "61%", CostStr: "$3.76", Color: "#facc15"},
				{Name: "qwen/qwen3-coder-flash", Percent: 39, PctStr: "39%", CostStr: "$2.44", Color: "#38bdf8"},
			},
			ChartSVG: generateAreaChartSVG(4),
			MetaText: "trend · openrouter multi-provider API",
		}
		rv.Clients = ClientsDeckCard{
			HasData: true,
			Segments: []ClientSegment{
				{Name: "Coding Agent", Percent: 85, PctStr: "85%", ReqStr: "180 req", Color: "#facc15"},
				{Name: "CLI Tool", Percent: 15, PctStr: "15%", ReqStr: "32 req", Color: "#38bdf8"},
			},
		}
		rv.CodeStats = CodeStatsDeckCard{
			HasData:      true,
			Added:        "+195 added",
			Removed:      "-88 removed",
			EqualizerSVG: generateCodeEqualizerSVG(4),
			MetaText:     "410 files · 190 commits · 68% AI-generated",
		}

	case v.AccountID == "ollama" || v.AccountID == "ollama-local":
		rv.HeroMeta = "ollama · Local Runtime · http://127.0.0.1:11434"
		rv.HeroPrimaryStat = "$10.50 / $30.00 · in 06d"
		rv.HeroCycleStat = "Local 4 models"
		rv.TeamBudget = MetricDeckCard{
			Title:    "LOCAL ACTIVITY",
			Percent:  35.0,
			PctStr:   "35.0%",
			Tone:     "green",
			SubLeft:  "38 / 100 req/day",
			SubRight: "62 remaining",
		}
		rv.BillingCycle = MetricDeckCard{
			Title:    "HARDWARE GPU LOAD",
			Percent:  48.0,
			PctStr:   "48.0%",
			Tone:     "lime",
			SubLeft:  "VRAM 11.5 / 24 GB",
			SubRight: "in 06d",
		}
		rv.ModelBurn = ModelBurnDeckCard{
			HasData: true,
			Segments: []ModelSegment{
				{Name: "qwen2.5-coder:32b", Percent: 55, PctStr: "55%", CostStr: "0.00 USD", Color: "#a78bfa"},
				{Name: "deepseek-coder-v2:16b", Percent: 30, PctStr: "30%", CostStr: "0.00 USD", Color: "#38bdf8"},
				{Name: "llama3.1:8b", Percent: 15, PctStr: "15%", CostStr: "0.00 USD", Color: "#4ade80"},
			},
			ChartSVG: generateAreaChartSVG(5),
			MetaText: "trend · local hardware compute",
		}
		rv.Clients = ClientsDeckCard{
			HasData: true,
			Segments: []ClientSegment{
				{Name: "Local Antigravity", Percent: 65, PctStr: "65%", ReqStr: "25 req", Color: "#a78bfa"},
				{Name: "Terminal CLI", Percent: 35, PctStr: "35%", ReqStr: "13 req", Color: "#38bdf8"},
			},
		}
		rv.CodeStats = CodeStatsDeckCard{
			HasData:      true,
			Added:        "+145 added",
			Removed:      "-42 removed",
			EqualizerSVG: generateCodeEqualizerSVG(5),
			MetaText:     "180 files · 95 commits · 88% AI-generated",
		}

	case strings.HasPrefix(v.AccountID, "cursor"):
		rv.HeroMeta = fmt.Sprintf("cursor · %s · %s", firstNonEmpty(v.Detail, "Pro"), v.AccountID)
		rv.HeroPrimaryStat = firstNonEmpty(v.Summary, "$0.00 remaining")
		rv.HeroCycleStat = firstNonEmpty(v.CycleSchedule, rv.NextDisplay)
		pct := 100.0
		if v.GaugePercent >= 0 {
			pct = v.GaugePercent
		}
		tone := "green"
		if pct <= 5.0 {
			tone = "crit"
		} else if pct < 50.0 {
			tone = "warn"
		}
		rv.TeamBudget = MetricDeckCard{
			Title:    "QUOTA REMAINING",
			Percent:  pct,
			PctStr:   fmt.Sprintf("%.1f%%", pct),
			Tone:     tone,
			SubLeft:  firstNonEmpty(v.Summary, "Active plan"),
			SubRight: firstNonEmpty(rv.NextDisplay, "current cycle"),
		}
		rv.BillingCycle = MetricDeckCard{
			Title:    "BILLING CYCLE",
			Percent:  math.Max(5.0, 100.0-pct),
			PctStr:   fmt.Sprintf("%.1f%%", math.Max(5.0, 100.0-pct)),
			Tone:     "lime",
			SubLeft:  firstNonEmpty(v.CycleSchedule, "Monthly quota"),
			SubRight: firstNonEmpty(rv.NextDisplay, "in cycle"),
		}
		rv.ModelBurn = ModelBurnDeckCard{
			HasData: true,
			Segments: []ModelSegment{
				{Name: "claude-4.5-opus", Percent: 50, PctStr: "50%", CostStr: "Active", Color: "#a78bfa"},
				{Name: "composer-1.5", Percent: 30, PctStr: "30%", CostStr: "Active", Color: "#fb923c"},
				{Name: "gpt-5.2", Percent: 20, PctStr: "20%", CostStr: "Active", Color: "#38bdf8"},
			},
			ChartSVG: generateAreaChartSVG(0),
			MetaText: "trend · Cursor IDE telemetry",
		}
		rv.Clients = ClientsDeckCard{
			HasData: true,
			Segments: []ClientSegment{
				{Name: "Composer", Percent: 65, PctStr: "65%", ReqStr: "Active", Color: "#a78bfa"},
				{Name: "Tab Autocomplete", Percent: 35, PctStr: "35%", ReqStr: "Active", Color: "#38bdf8"},
			},
		}
		rv.CodeStats = CodeStatsDeckCard{
			HasData:      true,
			Added:        "+220 added",
			Removed:      "-85 removed",
			EqualizerSVG: generateCodeEqualizerSVG(0),
			MetaText:     "Cursor project worktree",
		}

	case strings.HasPrefix(v.AccountID, "opencode"):
		rv.HeroMeta = fmt.Sprintf("opencode · %s · %s", firstNonEmpty(v.Detail, "OpenCode Go"), v.AccountID)
		rv.HeroPrimaryStat = firstNonEmpty(v.Summary, "OpenCode Go (70 models)")
		rv.HeroCycleStat = firstNonEmpty(v.CycleSchedule, rv.NextDisplay)
		pct := 100.0
		if v.GaugePercent >= 0 {
			pct = v.GaugePercent
		}
		rv.TeamBudget = MetricDeckCard{
			Title:    "5-HOUR LIMIT",
			Percent:  pct,
			PctStr:   fmt.Sprintf("%.1f%%", pct),
			Tone:     "green",
			SubLeft:  firstNonEmpty(v.Summary, "Active quota"),
			SubRight: firstNonEmpty(rv.NextDisplay, "rolling window"),
		}
		rv.BillingCycle = MetricDeckCard{
			Title:    "WEEKLY / MONTHLY",
			Percent:  math.Max(10.0, math.Min(100.0, pct*0.85)),
			PctStr:   fmt.Sprintf("%.1f%%", math.Max(10.0, math.Min(100.0, pct*0.85))),
			Tone:     "lime",
			SubLeft:  firstNonEmpty(v.CycleSchedule, "OpenCode Go"),
			SubRight: firstNonEmpty(rv.NextDisplay, "7d reset"),
		}
		rv.ModelBurn = ModelBurnDeckCard{
			HasData: true,
			Segments: []ModelSegment{
				{Name: "claude-fable-5", Percent: 55, PctStr: "55%", CostStr: "Included", Color: "#a78bfa"},
				{Name: "claude-fable-5-1", Percent: 30, PctStr: "30%", CostStr: "Included", Color: "#38bdf8"},
				{Name: "claude-opus-4-6", Percent: 15, PctStr: "15%", CostStr: "Included", Color: "#4ade80"},
			},
			ChartSVG: generateAreaChartSVG(1),
			MetaText: "trend · OpenCode Go model burn",
		}
		rv.Clients = ClientsDeckCard{
			HasData: true,
			Segments: []ClientSegment{
				{Name: "OpenCode CLI", Percent: 85, PctStr: "85%", ReqStr: "Active", Color: "#a78bfa"},
				{Name: "Subagents", Percent: 15, PctStr: "15%", ReqStr: "Tasks", Color: "#38bdf8"},
			},
		}
		rv.CodeStats = CodeStatsDeckCard{
			HasData:      true,
			Added:        "+310 added",
			Removed:      "-140 removed",
			EqualizerSVG: generateCodeEqualizerSVG(1),
			MetaText:     "OpenCode sessions",
		}

	case strings.HasPrefix(v.AccountID, "antigravity"):
		rv.HeroMeta = fmt.Sprintf("antigravity · %s · Google DeepMind", v.AccountID)
		rv.HeroPrimaryStat = firstNonEmpty(v.Summary, v.Message, "CLI Runtime Active")
		rv.HeroCycleStat = firstNonEmpty(v.CycleSchedule, "Local Container Fleet")
		tone := "green"
		if v.StatusBadge == "AUTH" {
			tone = "amber"
		}
		rv.TeamBudget = MetricDeckCard{
			Title:    "SESSION STATUS",
			Percent:  100.0,
			PctStr:   "100.0%",
			Tone:     tone,
			SubLeft:  v.AccountID,
			SubRight: v.Status,
		}
		rv.BillingCycle = MetricDeckCard{
			Title:    "CONTAINER FLEET",
			Percent:  80.0,
			PctStr:   "80.0%",
			Tone:     "lime",
			SubLeft:  "box-orchestrator",
			SubRight: "tmux ready",
		}
		rv.ModelBurn = ModelBurnDeckCard{
			HasData: true,
			Segments: []ModelSegment{
				{Name: "gemini-3.8-flash", Percent: 65, PctStr: "65%", CostStr: "Active", Color: "#38bdf8"},
				{Name: "gemini-3.8-pro", Percent: 35, PctStr: "35%", CostStr: "Active", Color: "#a78bfa"},
			},
			ChartSVG: generateAreaChartSVG(2),
			MetaText: "trend · Antigravity AI agent compute",
		}
		rv.Clients = ClientsDeckCard{
			HasData: true,
			Segments: []ClientSegment{
				{Name: "Antigravity CLI", Percent: 75, PctStr: "75%", ReqStr: "Fleet Pane", Color: "#38bdf8"},
				{Name: "IDE Extension", Percent: 25, PctStr: "25%", ReqStr: "Desktop", Color: "#a78bfa"},
			},
		}
		rv.CodeStats = CodeStatsDeckCard{
			HasData:      true,
			Added:        "+420 added",
			Removed:      "-150 removed",
			EqualizerSVG: generateCodeEqualizerSVG(2),
			MetaText:     "Autonomous agent work",
		}

	case strings.HasPrefix(v.AccountID, "codex"):
		rv.HeroMeta = fmt.Sprintf("openai · %s · dev@acme-corp.dev", firstNonEmpty(v.Detail, "OpenAI Codex CLI"))
		rv.HeroPrimaryStat = firstNonEmpty(v.Summary, "$11.40 today")
		nextDisp := strings.TrimSpace(rv.NextDisplay)
		if v.CycleSchedule != "" {
			rv.HeroCycleStat = v.CycleSchedule
		} else if nextDisp != "" && nextDisp != "—" && nextDisp != "–" && nextDisp != "-" {
			rv.HeroCycleStat = "Resets in " + nextDisp
		} else {
			rv.HeroCycleStat = "Daily rolling window"
		}
		rv.TeamBudget = MetricDeckCard{
			Title:    "DAILY SPEND",
			Percent:  22.8,
			PctStr:   "22.8%",
			Tone:     "green",
			SubLeft:  firstNonEmpty(v.Summary, "$11.40 today"),
			SubRight: "7d $48.20 spend",
		}
		rv.BillingCycle = MetricDeckCard{
			Title:    "TOKEN VELOCITY",
			Percent:  51.6,
			PctStr:   "516k tok",
			Tone:     "lime",
			SubLeft:  "420k in / 96k out",
			SubRight: "today",
		}
		rv.ModelBurn = ModelBurnDeckCard{
			HasData: true,
			Segments: []ModelSegment{
				{Name: "gpt-5.1-codex", Percent: 100, PctStr: "100%", CostStr: "$11.40", Color: "#10a37f"},
			},
			ChartSVG: generateAreaChartSVG(3),
			MetaText: "trend · daily by model · OpenAI Codex",
		}
		rv.Clients = ClientsDeckCard{
			HasData: true,
			Segments: []ClientSegment{
				{Name: "Codex CLI", Percent: 80, PctStr: "80%", ReqStr: "Active", Color: "#10a37f"},
				{Name: "API Runner", Percent: 20, PctStr: "20%", ReqStr: "Active", Color: "#38bdf8"},
			},
		}
		rv.CodeStats = CodeStatsDeckCard{
			HasData:      true,
			Added:        "+260 added",
			Removed:      "-75 removed",
			EqualizerSVG: generateCodeEqualizerSVG(3),
			MetaText:     "Codex code synthesis sessions",
		}

	case strings.HasPrefix(v.AccountID, "command"):
		plan := "GOAT"
		if strings.TrimSpace(v.Detail) != "" && !strings.HasPrefix(strings.TrimSpace(v.Detail), "$") {
			plan = v.Detail
		}
		rv.HeroMeta = fmt.Sprintf("command-code · %s · dev@acme-corp.dev", plan)
		rv.HeroPrimaryStat = firstNonEmpty(v.Summary, "$34.16 / $70.00 remaining")
		nextDisp := strings.TrimSpace(rv.NextDisplay)
		if v.CycleSchedule != "" {
			rv.HeroCycleStat = v.CycleSchedule
		} else if nextDisp != "" && nextDisp != "—" && nextDisp != "–" && nextDisp != "-" {
			rv.HeroCycleStat = "Resets in " + nextDisp
		} else {
			rv.HeroCycleStat = "Weekly 80.0% rem"
		}
		rv.TeamBudget = MetricDeckCard{
			Title:    "MONTHLY CREDITS",
			Percent:  48.8,
			PctStr:   "48.8%",
			Tone:     "green",
			SubLeft:  "$35.84 / $70.00",
			SubRight: "$34.16 remaining",
		}
		rv.BillingCycle = MetricDeckCard{
			Title:    "WEEKLY ALLOWANCE",
			Percent:  80.0,
			PctStr:   "80.0%",
			Tone:     "lime",
			SubLeft:  "7d rolling window",
			SubRight: "80.0% remaining",
		}
		rv.ModelBurn = ModelBurnDeckCard{
			HasData: true,
			Segments: []ModelSegment{
				{Name: "command-r-plus", Percent: 65, PctStr: "65%", CostStr: "$23.30", Color: "#a78bfa"},
				{Name: "command-r", Percent: 35, PctStr: "35%", CostStr: "$12.54", Color: "#38bdf8"},
			},
			ChartSVG: generateAreaChartSVG(4),
			MetaText: "trend · Command Code (GOAT) compute",
		}
		rv.Clients = ClientsDeckCard{
			HasData: true,
			Segments: []ClientSegment{
				{Name: "Command CLI", Percent: 85, PctStr: "85%", ReqStr: "Fleet", Color: "#a78bfa"},
				{Name: "Subagents", Percent: 15, PctStr: "15%", ReqStr: "Tasks", Color: "#38bdf8"},
			},
		}
		rv.CodeStats = CodeStatsDeckCard{
			HasData:      true,
			Added:        "+340 added",
			Removed:      "-110 removed",
			EqualizerSVG: generateCodeEqualizerSVG(4),
			MetaText:     "1.4M tokens · Command Code workspace",
		}

	default:
		rv.HasCockpitCards = true
		rv.HeroPrimaryStat = firstNonEmpty(v.Summary, "$0.00 remaining")
		cycleStat := strings.TrimSpace(v.CycleSchedule)
		if cycleStat == "" {
			nextDisp := strings.TrimSpace(rv.NextDisplay)
			if nextDisp != "" && nextDisp != "—" && nextDisp != "–" && nextDisp != "-" {
				cycleStat = "Resets in " + nextDisp
			} else {
				cycleStat = "Rolling window"
			}
		}
		rv.HeroCycleStat = cycleStat
		pct := 50.0
		if v.GaugePercent >= 0 {
			pct = v.GaugePercent
		}
		subRight := strings.TrimSpace(rv.NextDisplay)
		if subRight == "—" || subRight == "–" || subRight == "-" {
			subRight = ""
		}
		rv.TeamBudget = MetricDeckCard{
			Title:    "USAGE QUOTA",
			Percent:  pct,
			PctStr:   fmt.Sprintf("%.1f%%", pct),
			Tone:     "green",
			SubLeft:  "Current window",
			SubRight: subRight,
		}
		rv.BillingCycle = MetricDeckCard{
			Title:    "CYCLE STATUS",
			Percent:  math.Min(100, pct*1.1),
			PctStr:   fmt.Sprintf("%.1f%%", math.Min(100, pct*1.1)),
			Tone:     "lime",
			SubLeft:  firstNonEmpty(v.CycleSchedule, "Active tier"),
			SubRight: subRight,
		}
		rv.ModelBurn = ModelBurnDeckCard{
			HasData: true,
			Segments: []ModelSegment{
				{Name: "Primary Model", Percent: 70, PctStr: "70%", CostStr: "$1.00", Color: "#a78bfa"},
				{Name: "Secondary Model", Percent: 30, PctStr: "30%", CostStr: "$0.45", Color: "#38bdf8"},
			},
			ChartSVG: generateAreaChartSVG(0),
			MetaText: "trend · telemetry series",
		}
		rv.Clients = ClientsDeckCard{
			HasData: true,
			Segments: []ClientSegment{
				{Name: "Primary Client", Percent: 80, PctStr: "80%", ReqStr: "100 req", Color: "#a78bfa"},
				{Name: "Secondary Client", Percent: 20, PctStr: "20%", ReqStr: "25 req", Color: "#38bdf8"},
			},
		}
		rv.CodeStats = CodeStatsDeckCard{
			HasData:      true,
			Added:        "+100 added",
			Removed:      "-50 removed",
			EqualizerSVG: generateCodeEqualizerSVG(0),
			MetaText:     "Active repository files",
		}
	}
}

func generateAreaChartSVG(seed int) template.HTML {
	points := []float64{
		12, 14, 18, 15, 20, 26, 18, 14, 15, 18, 16, 15, 14, 16, 17, 16,
		15, 14, 15, 17, 18, 16, 15, 16, 18, 22, 28, 35, 30, 24, 28, 32,
		25, 22, 20, 18, 20, 22, 21, 23, 22, 24, 28, 32, 28, 24, 26, 28,
		22, 20, 22, 24, 23, 21, 20, 19, 18, 17, 18, 18,
	}
	if seed > 0 {
		for i := range points {
			shift := math.Sin(float64(i+seed)*0.5) * 4
			points[i] = math.Max(8, points[i]+shift)
		}
	}
	w := 600
	h := 70
	maxVal := 45.0
	step := float64(w) / float64(len(points)-1)

	var pathD strings.Builder
	var areaD strings.Builder

	for i, v := range points {
		x := float64(i) * step
		y := float64(h) - (v/maxVal)*float64(h-12) - 4
		if i == 0 {
			pathD.WriteString(fmt.Sprintf("M %.1f %.1f", x, y))
			areaD.WriteString(fmt.Sprintf("M %.1f %.1f L %.1f %.1f", x, float64(h), x, y))
		} else {
			pathD.WriteString(fmt.Sprintf(" L %.1f %.1f", x, y))
			areaD.WriteString(fmt.Sprintf(" L %.1f %.1f", x, y))
		}
	}
	areaD.WriteString(fmt.Sprintf(" L %.1f %.1f Z", float64(w), float64(h)))

	gradID := fmt.Sprintf("burnGrad_%d", seed)
	svg := fmt.Sprintf(`<svg viewBox="0 0 %d %d" class="burn-svg-chart" preserveAspectRatio="none">
		<defs>
			<linearGradient id="%s" x1="0" y1="0" x2="0" y2="1">
				<stop offset="0%%" stop-color="#38bdf8" stop-opacity="0.35"/>
				<stop offset="100%%" stop-color="#38bdf8" stop-opacity="0.0"/>
			</linearGradient>
		</defs>
		<path d="%s" fill="url(#%s)"/>
		<path d="%s" fill="none" stroke="#38bdf8" stroke-width="1.8" stroke-linejoin="round" stroke-linecap="round"/>
	</svg>`, w, h, gradID, areaD.String(), gradID, pathD.String())

	return template.HTML(svg)
}

func generateCodeEqualizerSVG(seed int) template.HTML {
	w := 340
	h := 60
	axisY := 32

	addHeights := []int{
		6, 12, 4, 8, 16, 22, 14, 8, 4, 10, 12, 6, 10, 6, 8, 14, 18, 10, 4, 6, 8, 12, 10, 14, 10, 6, 8, 12, 10, 16, 20, 10, 4, 6,
	}
	delHeights := []int{
		4, 8, 6, 12, 16, 10, 6, 14, 10, 4, 12, 16, 6, 10, 14, 8, 6, 10, 16, 12, 6, 4, 8, 12, 14, 10, 6, 12, 8, 4, 10, 14, 8, 4,
	}

	numBars := len(addHeights)
	barW := 3
	gap := 9
	startX := 12

	var bars strings.Builder
	bars.WriteString(fmt.Sprintf(`<line x1="8" y1="%d" x2="%d" y2="%d" stroke="#334155" stroke-width="1"/>`, axisY, w-8, axisY))

	for i := 0; i < numBars; i++ {
		x := startX + i*gap
		addH := addHeights[i]
		delH := delHeights[i]
		if seed > 0 {
			addH = int(math.Max(2, float64(addH)+(math.Sin(float64(i+seed))*3)))
			delH = int(math.Max(2, float64(delH)+(math.Cos(float64(i+seed))*3)))
		}

		addY := axisY - addH - 1
		bars.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" rx="1" fill="#4ade80"/>`, x, addY, barW, addH))

		delY := axisY + 2
		bars.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" rx="1" fill="#f87171"/>`, x, delY, barW, delH))
	}

	svg := fmt.Sprintf(`<svg viewBox="0 0 %d %d" class="equalizer-svg-chart" preserveAspectRatio="none">%s</svg>`, w, h, bars.String())
	return template.HTML(svg)
}

func calculateGlobalStats(views []AccountView, env Envelope) GlobalStats {
	total := len(views)
	if total == 0 {
		return GlobalStats{
			Health:        "100%",
			HealthTag:     "Ready",
			HealthTone:    "ok",
			HealthNote:    "Waiting for agents",
			TokenVelocity: "0",
			TokenNote:     "No active streams",
			Spend:         "$0.00",
			SpendNote:     "Active billing cycle",
			CacheHit:      "0%",
			CacheNote:     "No requests recorded",
		}
	}

	healthy := 0
	for _, v := range views {
		badge := strings.ToUpper(v.StatusBadge)
		stat := strings.ToLower(v.Status)
		if !strings.Contains(badge, "CRIT") && !strings.Contains(badge, "WARN") && !strings.Contains(stat, "err") && !strings.Contains(stat, "limit") {
			healthy++
		}
	}
	healthPct := float64(healthy) / float64(total) * 100
	healthTone := "ok"
	healthTag := "Optimal"
	if healthPct < 70 {
		healthTone = "crit"
		healthTag = "Degraded"
	} else if healthPct < 90 {
		healthTone = "warn"
		healthTag = "Attention"
	}

	return GlobalStats{
		Health:        fmt.Sprintf("%.1f%%", healthPct),
		HealthTag:     healthTag,
		HealthTone:    healthTone,
		HealthNote:    fmt.Sprintf("%d of %d agents well below burst limits", healthy, total),
		TokenVelocity: "3.48M",
		TokenNote:     "+14.2% vs yesterday · Peak: 14:00 UTC",
		Spend:         "$242.30",
		SpendNote:     "of $500.00 ceiling · 18 days left",
		CacheHit:      "88.4%",
		CacheNote:     "Prompt caching saved $68.40",
	}
}

func avatarFor(providerID, accountID string) (text, color, bg string) {
	pid := strings.ToLower(providerID)
	acc := strings.ToLower(accountID)
	switch {
	case strings.Contains(pid, "opencode") || strings.Contains(acc, "opencode"):
		return "OC", "#2563eb", "#eff6ff"
	case strings.Contains(pid, "gemini") || strings.Contains(acc, "gemini"):
		return "G", "#dc2626", "#fef2f2"
	case strings.Contains(pid, "cursor") || strings.Contains(acc, "cursor"):
		return "CU", "#0f172a", "#f8fafc"
	case strings.Contains(pid, "claude") || strings.Contains(acc, "claude"):
		return "CC", "#ca8a04", "#fefce8"
	case strings.Contains(pid, "codex") || strings.Contains(acc, "codex") || strings.Contains(pid, "openai") || strings.Contains(acc, "openai"):
		return "OX", "#16a34a", "#f0fdf4"
	case strings.Contains(pid, "antigravity") || strings.Contains(acc, "antigravity"):
		return "AG", "#7c3aed", "#f5f3ff"
	case strings.Contains(pid, "ollama") || strings.Contains(acc, "ollama"):
		return "OL", "#0284c7", "#f0f9ff"
	case strings.Contains(pid, "copilot") || strings.Contains(acc, "copilot"):
		return "GH", "#1e293b", "#f1f5f9"
	case strings.Contains(pid, "anthropic") || strings.Contains(acc, "anthropic"):
		return "AN", "#d97706", "#fffbeb"
	case strings.Contains(pid, "deepseek") || strings.Contains(acc, "deepseek"):
		return "DS", "#0284c7", "#f0f9ff"
	case strings.Contains(pid, "xai") || strings.Contains(acc, "xai"):
		return "X", "#0f172a", "#f8fafc"
	case strings.Contains(pid, "groq") || strings.Contains(acc, "groq"):
		return "GQ", "#ea580c", "#fff7ed"
	default:
		s := accountID
		if len(s) >= 2 {
			return strings.ToUpper(s[:2]), "#1a1a1c", "#f4f4f5"
		} else if len(s) == 1 {
			return strings.ToUpper(s), "#1a1a1c", "#f4f4f5"
		}
		return "AI", "#1a1a1c", "#f4f4f5"
	}
}

func isCardAlert(badge, status string) bool {
	s := strings.ToLower(badge + " " + status)
	return strings.Contains(s, "warn") || strings.Contains(s, "crit") || strings.Contains(s, "alert") || strings.Contains(s, "limit")
}

func timebandPill(it usageItem) string {
	s := strings.ToLower(it.Label + " " + it.Short + " " + it.Group)
	switch {
	case strings.Contains(s, "5h") || strings.Contains(s, "5 hour") || strings.Contains(s, "5-hour"):
		return "5h"
	case strings.Contains(s, "24h") || strings.Contains(s, "daily") || strings.Contains(s, "today"):
		return "24h"
	case strings.Contains(s, "7d") || strings.Contains(s, "weekly") || strings.Contains(s, "7 day"):
		return "7d"
	case strings.Contains(s, "30d") || strings.Contains(s, "month"):
		return "30d"
	case strings.Contains(s, "fast"):
		return "Fast"
	case strings.Contains(s, "burst"):
		return "Burst"
	case strings.Contains(s, "rpm") || strings.Contains(s, "minute"):
		return "RPM"
	case strings.Contains(s, "spend") || strings.Contains(s, "cost") || strings.Contains(s, "$"):
		return "Spend"
	case strings.Contains(s, "token") || strings.Contains(s, "tok"):
		return "Tokens"
	case it.Short != "" && len(it.Short) <= 6:
		return it.Short
	default:
		return "Quota"
	}
}

func timebandLabel(label string) string {
	s := strings.ToLower(label)
	switch {
	case strings.Contains(s, "5h") || strings.Contains(s, "5 hour") || strings.Contains(s, "5-hour"):
		return "5h"
	case strings.Contains(s, "24h") || strings.Contains(s, "daily") || strings.Contains(s, "today"):
		return "24h"
	case strings.Contains(s, "7d") || strings.Contains(s, "weekly") || strings.Contains(s, "7 day"):
		return "7d"
	case strings.Contains(s, "30d") || strings.Contains(s, "month"):
		return "30d"
	case strings.Contains(s, "fast"):
		return "Fast"
	case strings.Contains(s, "burst"):
		return "Burst"
	case strings.Contains(s, "rpm") || strings.Contains(s, "minute"):
		return "RPM"
	case strings.Contains(s, "spend") || strings.Contains(s, "cost") || strings.Contains(s, "$"):
		return "Spend"
	case strings.Contains(s, "token") || strings.Contains(s, "tok"):
		return "Tokens"
	default:
		return "Quota"
	}
}

func barTone(tone string) string {
	switch tone {
	case "crit", "danger", "red":
		return "bar-red"
	case "warn", "warning", "yellow", "amber":
		return "bar-amber"
	default:
		return "bar-blue"
	}
}

func cleanQuotaLabel(pill, label string) string {
	p := strings.TrimSpace(pill)
	l := strings.TrimSpace(label)
	if l == "" {
		l = p
	}

	// If label duplicates pill (e.g. pill="5h", label="5h"), replace with human-readable name.
	if strings.EqualFold(p, l) {
		return humanReadableQuotaName(p)
	}

	// If label starts with pill prefix, strip it cleanly.
	pLower := strings.ToLower(p)
	lLower := strings.ToLower(l)
	if pLower != "" && strings.HasPrefix(lLower, pLower) {
		remainder := l[len(p):]
		remainder = strings.TrimLeft(remainder, " -_:·/")
		remainder = strings.TrimSpace(remainder)
		if remainder != "" {
			return remainder
		}
		return humanReadableQuotaName(p)
	}

	return l
}

func humanReadableQuotaName(pill string) string {
	switch strings.ToLower(strings.TrimSpace(pill)) {
	case "5h":
		return "5-Hour Limit"
	case "24h", "daily", "today", "day":
		return "Daily Quota"
	case "7d", "weekly", "week":
		return "Weekly Quota"
	case "30d", "monthly", "month":
		return "Monthly Quota"
	case "spend", "cost", "$":
		return "Monthly Spend"
	case "fast":
		return "Fast Quota"
	case "burst":
		return "Burst Quota"
	case "rpm", "minute":
		return "Request Limit"
	case "token", "tokens", "tok":
		return "Token Quota"
	default:
		p := strings.TrimSpace(pill)
		if p == "" || strings.EqualFold(p, "quota") {
			return "Usage Quota"
		}
		return p + " Quota"
	}
}

func cleanQuotaTitle(pill string, it usageItem) string {
	raw := it.Short
	if strings.TrimSpace(raw) == "" || strings.EqualFold(strings.TrimSpace(raw), strings.TrimSpace(pill)) {
		if strings.TrimSpace(it.Label) != "" {
			raw = it.Label
		}
	}
	if strings.TrimSpace(raw) == "" {
		raw = pill
	}
	return cleanQuotaLabel(pill, raw)
}


