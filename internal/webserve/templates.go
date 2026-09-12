package webserve

import (
	"bytes"
	"embed"
	"html/template"
	"io"
	"strconv"
	"strings"

	"github.com/nurulislamz/agentusage/internal/core"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

type cockpitCtx struct {
	M renderModel
	V renderView
}

type cardCtx struct {
	M    renderModel
	Card DetailCard
}

type shellData struct {
	Filter     string
	ThemeSlug  string
	ThemeColor string
}

var templateFuncs = template.FuncMap{
	"toneClass":    toneClass,
	"gaugeColor":   gaugeColor,
	"pillClass":    pillClass,
	"pct0":         pct0,
	"pct1":         pct1,
	"compact":      formatCompact,
	"upper":        upper,
	"subtitle":     viewSubtitle,
	"accent":       accentOr,
	"cardColor":    cardColor,
	"withIn":       withIn,
	"gaugeLeft":    gaugeLeft,
	"gaugeRight":   gaugeRight,
	"arcHub":       arcHub,
	"arcReset":     arcResetText,
	"stripOverlay": stripOverlay,
	"stripReset":   stripResetText,
	"rowPct":       rowPct,
	"rowCaption":   rowCaption,
	"sparkBars":    sparkBars,
	"sparkLine":    sparkLine,
	"cardSpark":    func(p []core.TimePoint) template.HTML { return firstHTML(sparkBars(p, 108, 26)) },
	"matrixSpark":  func(p []core.TimePoint) template.HTML { return firstHTML(sparkBars(p, 72, 18), sparkLine(p, 72, 18)) },
	"bentoSpark":   func(p []core.TimePoint) template.HTML { return firstHTML(sparkBars(p, 60, 16), sparkLine(p, 60, 16)) },
	"stripSpark":   func(p []core.TimePoint) template.HTML { return firstHTML(sparkBars(p, 72, 22), sparkLine(p, 72, 22)) },
	"dialSpark":    func(p []core.TimePoint) template.HTML { return firstHTML(sparkLine(p, 120, 32)) },
	"cockpitSpark": func(p []core.TimePoint) template.HTML {
		return firstHTML(sparkLine(p, 180, 42), sparkBars(p, 180, 42))
	},
	"themeVars":     themeVarsCSS,
	"timebandPill":  timebandPill,
	"timebandLabel": timebandLabel,
	"barTone":       barTone,
	"isCardAlert":   isCardAlert,
	"cctx":          func(m renderModel, v renderView) cockpitCtx { return cockpitCtx{M: m, V: v} },
	"cardctx":       func(m renderModel, c DetailCard) cardCtx { return cardCtx{M: m, Card: c} },
}

var pageTemplates = template.Must(template.New("agentusage").Funcs(templateFuncs).ParseFS(templateFS, "templates/*.tmpl"))

func firstHTML(candidates ...template.HTML) template.HTML {
	for _, c := range candidates {
		if c != "" {
			return c
		}
	}
	return ""
}

func executeTemplate(w io.Writer, name string, data any) error {
	var buf bytes.Buffer
	if err := pageTemplates.ExecuteTemplate(&buf, name, data); err != nil {
		return err
	}
	_, err := w.Write(buf.Bytes())
	return err
}

func viewSubtitle(v renderView) string {
	return firstNonEmpty(v.TagLabel, v.Detail, v.ProviderName, v.ProviderID)
}

func accentOr(color string) template.CSS {
	if strings.TrimSpace(color) == "" {
		return "var(--accent)"
	}
	return template.CSS(color)
}

// cardColor resolves a card accent, defaulting to the secondary text colour.
func cardColor(color string) template.CSS {
	if strings.TrimSpace(color) == "" {
		return "var(--fg-2)"
	}
	return template.CSS(color)
}

func withIn(d string) string {
	d = strings.TrimSpace(d)
	if d == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(d), "in ") || strings.HasPrefix(strings.ToLower(d), "expired") {
		return d
	}
	return "in " + d
}

func rowPct(row DetailRow) string {
	pct := 0.0
	if row.Percent != nil {
		pct = *row.Percent
	}
	return strconv.FormatFloat(pct, 'f', 1, 64)
}

func rowCaption(row DetailRow, mode string) string {
	if strings.TrimSpace(row.Hint) != "" {
		return row.Hint
	}
	return rowPct(row) + "% " + mode
}
