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
	Label    string
	Short    string
	Value    string
	Hint     string
	ResetIn  string
	Tone     string
	Group    string
	Pct      *float64
	Urgent   bool
	Depleted bool
	// Caption is the mode-aware percent caption ("58% left", "12% used",
	// "Limit reached") shared by every board renderer.
	Caption string
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
	Label    string
	Pct      *float64
	Value    string
	Tone     string
	Caption  string
	Depleted bool
	// Reset is the time-until-reset for this specific quota window
	// ("4h12m", "2d03h"), empty when the provider reports none.
	Reset  string
	Urgent bool
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
	Index          int
	Active         bool
	Lines          []usageLine
	Items          []usageItem
	Tables         []usageItem
	Graphs         []gaugeGroup
	Depleted       bool
	SummaryDisplay string
	ResetTitle     string
	NextDisplay    string
	Urgent         bool
	MatrixLines    []usageLine
	MatrixCells    []*usageLine
	BentoRows      []bentoRow
	Meters         []usageLine
	Cards          []DetailCard
	FirstTone      string
	TrendStats     []usageItem
	RefreshedText  string
	Expanded       bool
	Position       int
	Total          int

	// Cockpit hero strip, filled from real account data.
	HeroAccountID   string
	HeroMeta        string
	HeroPrimaryStat string
	HeroCycleStat   string

	// Ceramic Studio properties
	AvatarText  string
	AvatarColor string
	AvatarBg    string
	IsAlert     bool
	BentoSpan   string
}

// GlobalStats carries the usage-first KPI banner. Every field is derived from
// the collected snapshots; nothing is fabricated.
type GlobalStats struct {
	Health      string
	HealthTag   string
	HealthTone  string
	HealthNote  string
	AtLimit     string
	AtLimitTag  string
	AtLimitTone string
	AtLimitNote string

	Tightest     string
	TightestTag  string
	TightestNote string
	TightestTone string

	Headroom      string
	HeadroomLabel string
	HeadroomNote  string
	HeadroomTone  string
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
	// FleetNote summarises tracked accounts and quota windows for the footer.
	FleetNote string
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
	// ProviderOrder is the stored provider-id order for drag-reorder.
	ProviderOrder string
	Toast         string
	Auth          bool
	Now           time.Time
}

const (
	emptyHintText    = "Start the telemetry daemon (agentusage daemon) or configure provider credentials."
	emptyHintTimeout = "The request timed out. Check the telemetry daemon and reload."
)

var (
	ratioRe     = regexp.MustCompile(`\$?\s*([\d,.]+)\s*/\s*\$?\s*([\d,.]+)`)
	moneyRe     = regexp.MustCompile(`\$([\d,.]+)`)
	magnitudeRe = regexp.MustCompile(`(\d+(?:\.\d+)?)\s*([KMB])?`)
	resetPreRe  = regexp.MustCompile(`(?i)^(next\s+resets?:?\s*|resets?\s+(in|at|on)?:?\s*|in\s+)`)
	hoursAgoRe  = regexp.MustCompile(`(?i)(\d+)h\s+ago`)
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
	m.Groups = groupRenderViews(m.Views, in.ProviderOrder)
	m.Stats = calculateGlobalStats(m.Views, used)
	m.FleetNote = fleetNote(m.Views)
	return m
}

// fleetNote summarises the tracked fleet for the footer dock: account count,
// live quota windows and any accounts pinned at their limit.
func fleetNote(views []renderView) string {
	windows := 0
	atLimit := 0
	for i := range views {
		if views[i].Depleted {
			atLimit++
		}
		for _, l := range views[i].Lines {
			if l.Pct != nil {
				windows++
			}
		}
	}
	note := fmt.Sprintf("%d accounts · %d quota windows", len(views), windows)
	if atLimit > 0 {
		note += fmt.Sprintf(" · %d at limit", atLimit)
	}
	return note
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
	if strings.EqualFold(strings.TrimSpace(v.TagLabel), "usage") {
		v.TagLabel = ""
	}
	fullLines := buildUsageLines(v, used)
	for i := range fullLines {
		if fullLines[i].Pct != nil {
			fullLines[i].Depleted = isDepleted(*fullLines[i].Pct, used)
			if fullLines[i].Caption == "" {
				fullLines[i].Caption = percentCaption(*fullLines[i].Pct, used)
			}
		}
	}
	// Board projections share one collapsed set: one bar per quota window so
	// bars/dials/strips/matrix/bento never duplicate a window.
	lines := collapseUsageLines(v, fullLines, used)
	items := buildUsageItems(v, lines, used)
	rv := renderView{
		AccountView:   v,
		Index:         index,
		Lines:         fullLines,
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
		if dur != "" {
			rv.NextDisplay = dur
		} else if next != "" {
			rv.NextDisplay = next
		}
		if title != "" {
			rv.ResetTitle = title
		} else if dur != "" {
			rv.ResetTitle = "Gemini Weekly resets in " + dur
		}
	}
	rv.SummaryDisplay = summaryDisplay(v.Summary)
	rv.MatrixLines = matrixLines(v, lines, used)
	rv.MatrixCells = matrixCells(rv.MatrixLines)
	rv.BentoRows = bentoRows(lines, used)
	rv.Cards = extraCards(v)
	rv.Meters = firstN(lines, 2)
	rv.TrendStats = trendStats(v.DailyCost)
	worstTone := "ok"
	for _, l := range lines {
		if l.Pct != nil {
			if isDepleted(*l.Pct, used) {
				rv.Depleted = true
			}
		}
		switch l.Tone {
		case "crit":
			worstTone = "crit"
		case "warn", "peach":
			if worstTone != "crit" {
				worstTone = l.Tone
			}
		}
	}
	if !rv.Depleted && (strings.Contains(strings.ToUpper(v.StatusBadge), "LIMIT") || strings.ToUpper(v.Status) == "LIMITED") {
		rv.Depleted = true
	}
	if rv.Depleted {
		worstTone = "crit"
	}
	if len(lines) > 0 {
		rv.FirstTone = worstTone
	}
	rv.AvatarText, rv.AvatarColor, rv.AvatarBg = avatarFor(v.ProviderID, v.AccountID)
	rv.IsAlert = isCardAlert(v.StatusBadge, v.Status) || rv.Depleted
	rv.BentoSpan = bentoSpanFor(v, rv.BentoRows, rv.IsAlert)
	buildCockpitHero(&rv, v)
	return rv
}

func groupRenderViews(views []renderView, providerOrder string) []renderGroup {
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
		balanceBentoGroupSpans(groups[i].Items)
	}
	return sortProviderGroups(groups, providerOrder)
}

// sortProviderGroups sorts provider groups by the stored drag order; ids
// absent from the stored order keep first-seen order after the known ones.
func sortProviderGroups(groups []renderGroup, providerOrder string) []renderGroup {
	order := strings.TrimSpace(providerOrder)
	if order == "" {
		return groups
	}
	rank := map[string]int{}
	for _, pid := range strings.Split(order, "|") {
		pid = strings.TrimSpace(pid)
		if pid != "" {
			rank[strings.ToLower(pid)] = len(rank)
		}
	}
	if len(rank) == 0 {
		return groups
	}
	sort.SliceStable(groups, func(i, j int) bool {
		ri, oki := rank[strings.ToLower(groups[i].ProviderID)]
		rj, okj := rank[strings.ToLower(groups[j].ProviderID)]
		// Unknown ids sort after known ones, keeping first-seen order.
		if !oki && !okj {
			return false
		}
		if !oki {
			return false
		}
		if !okj {
			return true
		}
		return ri < rj
	})
	return groups
}

func balanceBentoGroupSpans(items []renderView) {
	n := len(items)
	if n == 0 {
		return
	}
	if n == 1 {
		if items[0].BentoSpan != "compact" {
			items[0].BentoSpan = "wide"
		}
		return
	}
	if n == 2 {
		if items[0].BentoSpan == "compact" && items[1].BentoSpan == "compact" {
			return
		}
		if items[0].IsAlert || len(items[0].BentoRows) >= len(items[1].BentoRows) {
			items[0].BentoSpan = "wide"
			if items[1].BentoSpan != "compact" {
				items[1].BentoSpan = "standard"
			}
		} else {
			items[1].BentoSpan = "wide"
			if items[0].BentoSpan != "compact" {
				items[0].BentoSpan = "standard"
			}
		}
		return
	}
	if n == 3 {
		for i := range items {
			if items[i].BentoSpan != "compact" {
				items[i].BentoSpan = "standard"
			}
		}
		return
	}
	for i := range items {
		if items[i].BentoSpan != "compact" {
			items[i].BentoSpan = "standard"
		}
	}
	rem := n % 3
	if rem == 2 {
		items[n-2].BentoSpan = "wide"
	} else if rem == 1 {
		if items[n-1].BentoSpan != "compact" {
			items[n-1].BentoSpan = "wide"
		}
	}
}

func rollupFor(items []renderView) (string, string) {
	anyCrit, anyWarn := false, false
	for _, v := range items {
		if v.Depleted {
			anyCrit = true
			break
		}
		s := strings.ToLower(v.StatusBadge + " " + v.Status)
		if containsAny(s, "limit", "err", "crit") {
			anyCrit = true
			break
		}
		if v.Urgent || containsAny(s, "warn", "auth") {
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
	hasActiveGauge := v.HasGauge && (v.GaugePercent > 0 || strings.Contains(strings.ToUpper(v.StatusBadge), "LIMIT") || strings.ToUpper(v.Status) == "LIMITED")
	if hasActiveGauge && v.GaugePercent > 0 {
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
		} else if hasActiveGauge {
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
	} else if hasActiveGauge {
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
			trimmedPart := strings.TrimSpace(part)
			if resetPreRe.MatchString(trimmedPart) || strings.HasPrefix(strings.ToLower(trimmedPart), "in ") || strings.HasPrefix(strings.ToLower(trimmedPart), "resets in") {
				if line.ResetIn == "" {
					line.ResetIn = stripResetPrefix(trimmedPart)
				}
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
			if line.Tone != "dim" && (trimAtDigit(part) != part || strings.Contains(part, "$")) {
				items = append(items, usageItem{Kind: "table", Label: partShort(part, line), Value: part})
			}
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
	if head != "" && !strings.EqualFold(head, "in") && !strings.EqualFold(head, "resets in") {
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
	res := strings.TrimSpace(resetPreRe.ReplaceAllString(s, ""))
	res = strings.TrimLeft(res, ":- ")
	if res == ":" || res == "-" {
		return ""
	}
	return res
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
	return stripResetPrefix(firstNonEmpty(v.NextReset, v.ResetHint))
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

// collapseUsageLines reduces usage lines to one row per quota window. It is
// the single collapse for every board projection (matrix, bento, bars,
// dials, strips): antigravity's Gemini + Claude "5-Hour Limit" siblings
// collapse to the tightest window; distinct windows stay separate.
// Cursor keeps its three usage metrics (Included / Auto / API) because they
// are distinct buckets, but they share one reset timer.
func collapseUsageLines(v AccountView, lines []usageLine, used bool) []usageLine {
	order := []string{}
	byWindow := map[string]int{}
	for _, l := range lines {
		if isCannotRenderError(l.Label) || isCannotRenderError(l.Value) || isCannotRenderError(l.Short) {
			continue
		}
		key := bentoWindowKey(l, v.ProviderID)
		if _, ok := byWindow[key]; !ok {
			byWindow[key] = len(order)
			order = append(order, key)
		}
	}
	rows := make([]usageLine, len(order))
	resets := make([]string, len(order))
	urgents := make([]bool, len(order))
	taken := make([]bool, len(order))
	for _, l := range lines {
		if isCannotRenderError(l.Label) || isCannotRenderError(l.Value) || isCannotRenderError(l.Short) {
			continue
		}
		i := byWindow[bentoWindowKey(l, v.ProviderID)]
		reset := stripResetPrefix(l.ResetIn)
		if !taken[i] {
			rows[i], resets[i], urgents[i], taken[i] = l, reset, l.Urgent, true
			continue
		}
		// Collapse duplicates: keep the tighter bar (mode-aware) but merge the
		// reset/urgent signal so the per-window chip still renders.
		if cur := rows[i]; cur.Pct != nil && l.Pct != nil && isTighter(*l.Pct, *cur.Pct, used) {
			rows[i] = l
		}
		resets[i] = firstNonEmpty(resets[i], reset)
		urgents[i] = urgents[i] || l.Urgent
	}
	out := make([]usageLine, 0, len(order))
	for i := range order {
		if !taken[i] {
			continue
		}
		l := rows[i]
		if l.ResetIn == "" {
			l.ResetIn = resets[i]
		}
		l.Urgent = urgents[i]
		if strings.EqualFold(strings.TrimSpace(v.ProviderID), "cursor") && l.ResetIn == "" {
			// Cursor's one billing reset applies to every usage bucket.
			if next := nextResetDisplay(v); next != "" {
				l.ResetIn = next
			}
		}
		out = append(out, l)
	}
	return out
}

func matrixLines(v AccountView, lines []usageLine, used bool) []usageLine {
	return firstN(collapseUsageLines(v, lines, used), 3)
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
		raw := strings.TrimSpace(v.LastRefreshed)
		if raw == "" {
			return ""
		}
		lower := strings.ToLower(raw)
		if strings.HasPrefix(lower, "last refreshed ") {
			raw = strings.TrimSpace(raw[len("last refreshed "):])
		} else if strings.HasPrefix(lower, "refreshed ") {
			raw = strings.TrimSpace(raw[len("refreshed "):])
		}
		if m := hoursAgoRe.FindStringSubmatch(raw); len(m) > 1 {
			if h, err := strconv.Atoi(m[1]); err == nil && h >= 24 {
				d := h / 24
				remH := h % 24
				formatted := fmt.Sprintf("%dd%dh ago", d, remH)
				raw = hoursAgoRe.ReplaceAllString(raw, formatted)
			}
		}
		return raw
	}
	age := now.Sub(v.Timestamp)
	if age < 5*time.Second {
		return "just now"
	}
	if age < 0 {
		age = 0
	}
	return formatAge(age.Milliseconds()) + " ago"
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

// bentoRows projects the collapsed board lines into bento tile rows; the
// window collapse itself lives in collapseUsageLines.
func bentoRows(lines []usageLine, used bool) []bentoRow {
	collapsed := collapseUsageLines(AccountView{}, lines, used)
	out := make([]bentoRow, 0, 3)
	for _, l := range collapsed {
		if len(out) == 3 {
			break
		}
		out = append(out, bentoRow{
			Label:    l.Label,
			Pct:      l.Pct,
			Value:    l.Value,
			Tone:     l.Tone,
			Caption:  l.Caption,
			Depleted: l.Depleted,
			Reset:    stripResetPrefix(l.ResetIn),
			Urgent:   l.Urgent,
		})
	}
	return out
}

// bentoWindowKey buckets a usage line by quota window so duplicate bars from
// sibling model families collapse into a single tile row. Cursor has a
// single reset timer but three distinct usage metrics, so its plan buckets
// key on identity like every other provider.
func bentoWindowKey(l usageLine, providerID string) string {
	// Short + Label only: Hint carries reset countdowns ("Resets in 15h 52m")
	// whose durations would falsely match window keywords ("15h" → "5h").
	s := normalizeWindowText(l.Short + " " + l.Label)
	switch {
	case containsAny(s, "five hour", "5 hour", "5h", "rolling"):
		return "5h"
	case containsAny(s, "week", "7d", "seven day"):
		return "week"
	case containsAny(s, "month", "30d", "monthly"):
		return "30d"
	case containsAny(s, "day", "daily", "today"):
		return "day"
	case strings.Contains(s, "session"):
		return "session"
	}
	// Distinct buckets (Included/Auto/API, model names) key on the identity.
	return strings.ToLower(strings.TrimSpace(firstNonEmpty(l.Short, l.Label)))
}

func isInternalTelemetryRow(row DetailRow) bool {
	l := strings.ToLower(strings.TrimSpace(row.Label))
	v := strings.ToLower(strings.TrimSpace(row.Value))
	h := strings.ToLower(strings.TrimSpace(row.Hint))
	all := l + " " + v + " " + h
	if strings.Contains(all, "telemetry") ||
		strings.Contains(all, "unmapped") ||
		strings.Contains(all, "canonical") ||
		strings.HasPrefix(l, "providers") ||
		strings.HasPrefix(l, "workspace") ||
		strings.Contains(all, "workspace /") {
		return true
	}
	telemetryKeys := []string{
		"link hint", "status file", "installation id", "session id",
		"ineligible reasons", "oauth scope", "sessions dirs", "mapped_target_missing",
		"limit_snapshot",
	}
	for _, k := range telemetryKeys {
		if strings.Contains(all, k) {
			return true
		}
	}
	return false
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
			if isInternalTelemetryRow(row) {
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
	if it.Pct != nil {
		return percentCaption(*it.Pct, it.Mode == "used")
	}
	if it.Display != "" {
		return it.Display
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
	// The reset slot shows only reset countdowns; usage captions live in the
	// dial hub, and accounts without reset data show an em dash.
	return resetCaption(it.ResetIn)
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
	borderStrong := "#1a1a1c"
	if !light {
		borderStrong = firstNonEmpty(t.Surface2, "color-mix(in srgb, var(--fg) 28%, var(--border))")
	}
	accentOn := "#ffffff"
	if isLightBase(t.Accent) {
		accentOn = firstNonEmpty(t.Base, "#141416")
	}
	tintFill := "color-mix(in oklab, var(--fg) 9%, transparent)"
	if light {
		tintFill = "color-mix(in srgb, var(--fg) 7%, transparent)"
	}
	pairs := [][2]string{
		{"--bg", t.Base}, {"--base", t.Base},
		{"--mantle", t.Mantle}, {"--surface-warm", t.Mantle},
		{"--surface", t.Surface0}, {"--surface0", t.Surface0},
		{"--surface1", t.Surface1}, {"--surface2", t.Surface2},
		{"--fg", t.Text}, {"--text", t.Text},
		{"--fg-2", t.Subtext}, {"--subtext", t.Subtext},
		{"--muted", t.Subtext}, {"--dim", firstNonEmpty(t.Dim, t.Subtext)},
		{"--accent", t.Accent}, {"--accent-on", accentOn},
		{"--lavender", t.Lavender}, {"--teal", t.Teal}, {"--sapphire", t.Sapphire},
		{"--success", t.Green}, {"--warn", t.Yellow},
		{"--danger", t.Red}, {"--crit", t.Red}, {"--peach", t.Peach},
		{"--border", t.Surface1}, {"--border-soft", t.Surface2},
		{"--info", firstNonEmpty(t.Blue, t.Accent)},
		{"--cobalt", "#0052ff"},
		{"--crimson", "#e53935"},
		{"--amber", "#f59e0b"},
		{"--emerald", "#10b981"},
		{"--border-strong", borderStrong},
		{"--tint-fill", tintFill},
		{"--tone", "var(--accent)"},
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

// buildCockpitHero fills the cockpit hero strip from real account data. The
// cockpit body renders the projected usage lines directly; the hero only
// carries identity, status and the active window summary.
func buildCockpitHero(rv *renderView, v AccountView) {
	rv.HeroAccountID = v.AccountID

	provider := firstNonEmpty(strings.TrimSpace(v.ProviderName), strings.TrimSpace(v.ProviderID))
	if detail := strings.TrimSpace(v.Detail); detail != "" && !strings.EqualFold(detail, provider) {
		if provider == "" {
			provider = detail
		} else {
			provider = provider + " · " + detail
		}
	}
	rv.HeroMeta = provider

	rv.HeroPrimaryStat = rv.SummaryDisplay
	if rv.HeroPrimaryStat == "" {
		rv.HeroPrimaryStat = strings.TrimSpace(v.Summary)
	}

	next := strings.TrimSpace(rv.NextDisplay)
	switch next {
	case "", "—", "–", "-":
		rv.HeroCycleStat = strings.TrimSpace(v.CycleSchedule)
	default:
		rv.HeroCycleStat = "Resets in " + next
	}
}

// calculateGlobalStats derives the KPI banner from concrete quota windows:
// fleet health, accounts at limit, the tightest window and average headroom.
// Every value comes from the collected snapshots; nothing is fabricated.
func calculateGlobalStats(views []renderView, used bool) GlobalStats {
	if len(views) == 0 {
		return GlobalStats{
			Health: "—", HealthTag: "Idle", HealthTone: "ok", HealthNote: "No accounts configured",
			AtLimit: "0", AtLimitTag: "All clear", AtLimitTone: "ok", AtLimitNote: "No quota windows tracked",
			Tightest: "—", TightestTag: "Idle", TightestTone: "ok", TightestNote: "No quota windows tracked",
			Headroom: "—", HeadroomLabel: headroomLabel(used), HeadroomNote: "No quota windows tracked", HeadroomTone: "ok",
		}
	}

	total := len(views)
	healthy := 0
	atLimit := 0
	var limitNames []string
	for i := range views {
		v := &views[i]
		badge := strings.ToUpper(v.StatusBadge)
		stat := strings.ToLower(v.Status)
		isLimit := v.Depleted || strings.Contains(badge, "LIMIT") || strings.Contains(stat, "limit")
		isDegraded := strings.Contains(badge, "CRIT") || strings.Contains(badge, "WARN") || strings.Contains(stat, "err") || isLimit
		if !isDegraded {
			healthy++
		}
		if isLimit {
			atLimit++
			if len(limitNames) < 3 {
				limitNames = append(limitNames, v.AccountID)
			}
		}
	}

	windowCount := 0
	sum := 0.0
	tightest := 0.0
	tightestName := ""
	for i := range views {
		v := &views[i]
		for _, l := range v.Lines {
			if l.Pct == nil {
				continue
			}
			pct := clampPctVal(*l.Pct)
			if windowCount == 0 || isTighter(pct, tightest, used) {
				tightest = pct
				tightestName = v.AccountID
				if win := firstNonEmpty(l.Short, l.Label); win != "" {
					tightestName += " · " + win
				}
			}
			windowCount++
			sum += pct
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

	stats := GlobalStats{
		Health:        fmt.Sprintf("%.1f%%", healthPct),
		HealthTag:     healthTag,
		HealthTone:    healthTone,
		HealthNote:    fmt.Sprintf("%d of %d agents well below burst limits", healthy, total),
		AtLimit:       strconv.Itoa(atLimit),
		AtLimitTone:   "ok",
		AtLimitTag:    "All clear",
		AtLimitNote:   "All accounts have headroom",
		Tightest:      "—",
		TightestTag:   "Idle",
		TightestTone:  "ok",
		TightestNote:  "No quota windows tracked",
		Headroom:      "—",
		HeadroomLabel: headroomLabel(used),
		HeadroomNote:  "No quota windows tracked",
		HeadroomTone:  "ok",
	}

	if atLimit > 0 {
		stats.AtLimitTone = "crit"
		stats.AtLimitTag = "At limit"
		note := strings.Join(limitNames, ", ")
		if atLimit > len(limitNames) {
			note = fmt.Sprintf("%s + %d more", note, atLimit-len(limitNames))
		}
		stats.AtLimitNote = note
	}

	if windowCount > 0 {
		stats.Tightest = fmt.Sprintf("%d%%", int(math.Round(tightest)))
		stats.TightestNote = tightestName
		stats.TightestTone = toneFromPercent(tightest, used)
		stats.TightestTag = stats.TightestTone

		avg := sum / float64(windowCount)
		stats.Headroom = fmt.Sprintf("%d%%", int(math.Round(avg)))
		stats.HeadroomNote = fmt.Sprintf("across %d quota windows", windowCount)
		stats.HeadroomTone = toneFromPercent(avg, used)
	}

	return stats
}

// isTighter reports whether pct is worse than current: lower in remaining
// mode, higher in used mode.
func isTighter(pct, current float64, used bool) bool {
	if used {
		return pct > current
	}
	return pct < current
}

func headroomLabel(used bool) string {
	if used {
		return "Average Burn"
	}
	return "Average Headroom"
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
	return strings.Contains(s, "warn") || strings.Contains(s, "crit") || strings.Contains(s, "alert") || strings.Contains(s, "limit") || strings.Contains(s, "auth") || strings.Contains(s, "err")
}

func bentoSpanFor(v AccountView, rows []bentoRow, isAlert bool) string {
	hasGauges := false
	for _, r := range rows {
		if r.Pct != nil {
			hasGauges = true
			break
		}
	}
	if len(rows) == 0 || !hasGauges {
		return "compact"
	}
	if isAlert || len(rows) >= 3 {
		return "wide"
	}
	return "standard"
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
	case strings.Contains(s, "5h") || strings.Contains(s, "5 hour") || strings.Contains(s, "5-hour") || strings.Contains(s, "five hour"):
		return "5h"
	case strings.Contains(s, "24h") || strings.Contains(s, "daily") || strings.Contains(s, "today"):
		return "24h"
	case strings.Contains(s, "7d") || strings.Contains(s, "weekly") || strings.Contains(s, "7 day") || strings.Contains(s, "week"):
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

	// Strip trailing "remaining" or "used" (e.g. "Five Hour Limit Remaining")
	lLower := strings.ToLower(l)
	for _, suff := range []string{" remaining", " used"} {
		if strings.HasSuffix(lLower, suff) {
			l = strings.TrimSpace(l[:len(l)-len(suff)])
			lLower = strings.ToLower(l)
			break
		}
	}

	// If label duplicates pill (e.g. pill="5h", label="5h"), replace with human-readable name.
	if strings.EqualFold(p, l) {
		return humanReadableQuotaName(p)
	}

	// If label starts with pill prefix or window-word prefix corresponding to the pill,
	// strip it cleanly to eliminate stutter like [5H] 5-Hour Limit or [5H] Five Hour Limit.
	pLower := strings.ToLower(p)
	prefixes := []string{pLower}
	switch pLower {
	case "5h":
		prefixes = append(prefixes, "5-hour", "5 hour", "5-hr", "5 hr")
	case "24h":
		prefixes = append(prefixes, "24-hour", "24 hour", "24-hr", "24 hr")
	case "7d":
		prefixes = append(prefixes, "7-day", "7 day")
	case "30d":
		prefixes = append(prefixes, "30-day", "30 day")
	}

	for _, pfx := range prefixes {
		if pfx != "" && strings.HasPrefix(lLower, pfx) {
			remainder := l[len(pfx):]
			remainder = strings.TrimLeft(remainder, " -_:·/")
			remainder = strings.TrimSpace(remainder)
			if remainder != "" {
				return remainder
			}
			return humanReadableQuotaName(p)
		}
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
