package webserve

import (
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/tui"
)

const (
	cookieLayout   = "au_layout"
	cookieAccount  = "au_account"
	cookieFilter   = "au_filter"
	cookieExpanded = "au_expanded"
	cookieView     = "au_view"
)

func (s *Server) cookiePath() string {
	if s.basePath == "" {
		return "/"
	}
	return s.basePath + "/"
}

func (s *Server) setUICookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		Path:     s.cookiePath(),
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
	})
}

func readUICookie(r *http.Request, name string) string {
	c, err := r.Cookie(name)
	if err != nil || c.Value == "" {
		return ""
	}
	if v, err := url.QueryUnescape(c.Value); err == nil {
		return v
	}
	return c.Value
}

func requestPort(r *http.Request) string {
	_, port, err := net.SplitHostPort(r.Host)
	if err != nil {
		return ""
	}
	return port
}

// resolveLayout resolves the active layout: explicit query, then stored choice,
// then the historical deploy-variant port heuristic.
func (s *Server) resolveLayout(r *http.Request) string {
	if q := strings.TrimSpace(r.URL.Query().Get("layout")); q != "" {
		return normalizeLayoutID(q)
	}
	if c := readUICookie(r, cookieLayout); c != "" {
		return normalizeLayoutID(c)
	}
	if port := requestPort(r); port != "" {
		if l := layoutFromPort(port); l != "" {
			return l
		}
	}
	return "split"
}

// resolveFilter resolves the filter: an explicit q parameter wins (including an
// empty value that clears the filter), otherwise the stored choice.
func (s *Server) resolveFilter(r *http.Request) string {
	if r.URL.Query().Has("q") {
		return strings.TrimSpace(r.URL.Query().Get("q"))
	}
	return strings.TrimSpace(readUICookie(r, cookieFilter))
}

func (s *Server) handleShell(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if r.Method == http.MethodHead {
		return
	}
	themeSlug := "ceramic-studio"
	themeColor := "#f5f5f7"
	env := s.envelopeOrError()
	if env.ThemeTokens.Name != "" {
		themeSlug = strings.ToLower(strings.ReplaceAll(env.ThemeTokens.Name, " ", "-"))
		if env.ThemeTokens.Base != "" {
			themeColor = env.ThemeTokens.Base
		}
	} else if s.collector != nil && s.collector.opts.Theme != "" {
		themeSlug = strings.ToLower(strings.ReplaceAll(s.collector.opts.Theme, " ", "-"))
	}
	data := shellData{
		Filter:     readUICookie(r, cookieFilter),
		ThemeSlug:  themeSlug,
		ThemeColor: themeColor,
	}
	if err := executeTemplate(w, "shell", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleAppFragment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	if !s.checkAuth(w, r) {
		return
	}
	q := r.URL.Query()
	account := strings.TrimSpace(q.Get("account"))
	if account == "" {
		account = readUICookie(r, cookieAccount)
	}
	view := readUICookie(r, cookieView)
	if account != "" && q.Has("account") {
		view = "detail"
	}
	in := renderInput{
		Layout:          s.resolveLayout(r),
		Filter:          s.resolveFilter(r),
		Account:         account,
		Dir:             strings.TrimSpace(q.Get("dir")),
		Expand:          strings.TrimSpace(q.Get("expand")),
		ExpandedAccount: readUICookie(r, cookieExpanded),
		MobileView:      view,
		Toast:           strings.TrimSpace(q.Get("toast")),
	}
	s.renderApp(w, r, in, q.Get("refresh") == "1", q.Get("focus"))
}

func (s *Server) handleInspect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	if !s.checkAuth(w, r) {
		return
	}
	account := strings.TrimSpace(r.URL.Query().Get("account"))
	if account == "" {
		account = readUICookie(r, cookieAccount)
	}
	env := s.envelopeOrError()
	model := buildRenderModel(env, renderInput{
		Layout:          s.resolveLayout(r),
		Filter:          s.resolveFilter(r),
		Account:         account,
		ExpandedAccount: readUICookie(r, cookieExpanded),
		MobileView:      "detail",
		Auth:            s.AuthEnabled(),
		Now:             time.Now(),
	})
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if model.Selected == nil {
		_, _ = w.Write([]byte(`<p class="dim">No account data.</p>`))
		return
	}
	s.setUICookie(w, cookieAccount, model.Selected.AccountID)
	if err := executeTemplate(w, "inspect", cockpitCtx{M: model, V: *model.Selected}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleThemeAction(w http.ResponseWriter, r *http.Request) {
	if !s.requireUIAction(w, r) {
		return
	}
	_ = r.ParseForm()
	theme := strings.TrimSpace(r.FormValue("theme"))
	backward := strings.EqualFold(strings.TrimSpace(r.FormValue("direction")), "backward")
	env, err := s.applyTheme(theme, backward)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.renderAppEnv(w, r, env, "Theme: "+env.ThemeTokens.Name, "", "")
}

func (s *Server) handleUsageModeAction(w http.ResponseWriter, r *http.Request) {
	if !s.requireUIAction(w, r) {
		return
	}
	_ = r.ParseForm()
	env, err := s.applyUsageMode(strings.TrimSpace(r.FormValue("usage_mode")))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.renderAppEnv(w, r, env, "Usage mode: "+usageModeLabel(env), "", "")
}

func (s *Server) handleLayoutAction(w http.ResponseWriter, r *http.Request) {
	if !s.requireUIAction(w, r) {
		return
	}
	_ = r.ParseForm()
	current := s.resolveLayout(r)
	target := strings.TrimSpace(r.FormValue("layout"))
	id := ""
	switch target {
	case "next":
		id = cycleLayoutID(current, layoutList)
	case "nextmain":
		id = cycleLayoutID(current, []layoutMeta{layoutMetaFor("split"), layoutMetaFor("matrix"), layoutMetaFor("bento")})
	case "":
		id = current
	default:
		id = normalizeLayoutID(target)
	}
	s.setUICookie(w, cookieLayout, id)
	env := s.envelopeOrError()
	s.renderAppEnv(w, r, env, "Layout: "+layoutMetaFor(id).Label, id, "")
}

func (s *Server) handleViewAction(w http.ResponseWriter, r *http.Request) {
	if !s.requireUIAction(w, r) {
		return
	}
	_ = r.ParseForm()
	view := "roster"
	if strings.EqualFold(strings.TrimSpace(r.FormValue("view")), "detail") {
		view = "detail"
	}
	s.setUICookie(w, cookieView, view)
	s.renderAppEnv(w, r, s.envelopeOrError(), "", "", view)
}

func cycleLayoutID(current string, ids []layoutMeta) string {
	idx := -1
	for i, l := range ids {
		if l.ID == normalizeLayoutID(current) {
			idx = i
			break
		}
	}
	return ids[(idx+1)%len(ids)].ID
}

func (s *Server) requireUIAction(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return false
	}
	if !s.checkAuth(w, r) {
		return false
	}
	if !isAllowedOrigin(r) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "cross-origin request forbidden"})
		return false
	}
	return true
}

func (s *Server) envelopeOrError() Envelope {
	env, err := s.collector.envelope()
	if err != nil {
		return Envelope{Error: err.Error()}
	}
	return env
}

// applyTheme cycles or sets the active theme, persists the choice, and returns
// a freshly projected envelope. Shared by the JSON and htmx entry points.
func (s *Server) applyTheme(target string, backward bool) (Envelope, error) {
	targetTheme := strings.TrimSpace(target)
	if targetTheme == "" {
		if backward {
			targetTheme = tui.CycleThemeBackward()
		} else {
			targetTheme = tui.CycleTheme()
		}
	} else {
		_ = tui.SetThemeByName(targetTheme)
	}
	if activeName := tui.ActiveTheme().Name; activeName != "" {
		targetTheme = activeName
	}
	if !s.collector.demo {
		if err := config.SaveTheme(targetTheme); err != nil {
			return Envelope{}, err
		}
	}
	s.collector.setTheme(targetTheme)
	return s.collector.envelope()
}

// applyUsageMode toggles or sets the usage mode, persists it, and returns a
// freshly projected envelope.
func (s *Server) applyUsageMode(requested string) (Envelope, error) {
	mode := normalizeUsageMode(requested)
	if strings.TrimSpace(requested) == "" {
		current := s.collector.opts.UsageMode
		if current == "" && s.collector.opts.Config != nil {
			current = s.collector.opts.Config.Dashboard.UsageMode
		}
		if normalizeUsageMode(current) == config.UsageModeUsed {
			mode = config.UsageModeRemaining
		} else {
			mode = config.UsageModeUsed
		}
	}
	if !s.collector.demo {
		if err := config.SaveDashboardUsageMode(mode); err != nil {
			return Envelope{}, err
		}
	}
	s.collector.setUsageMode(mode)
	return s.collector.envelope()
}

func (s *Server) renderApp(w http.ResponseWriter, r *http.Request, in renderInput, refresh bool, focus string) {
	env, err := s.collector.envelopeRefresh(refresh, focus)
	if err != nil {
		env = Envelope{Error: err.Error()}
	}
	s.writeApp(w, r, env, in)
}

func (s *Server) renderAppEnv(w http.ResponseWriter, r *http.Request, env Envelope, toast, layoutOverride, viewOverride string) {
	account := readUICookie(r, cookieAccount)
	view := viewOverride
	if view == "" {
		view = readUICookie(r, cookieView)
	}
	in := renderInput{
		Layout:          layoutOverride,
		Filter:          s.resolveFilter(r),
		Account:         account,
		ExpandedAccount: readUICookie(r, cookieExpanded),
		MobileView:      view,
		Toast:           toast,
	}
	if in.Layout == "" {
		in.Layout = s.resolveLayout(r)
	}
	s.writeApp(w, r, env, in)
}

func (s *Server) writeApp(w http.ResponseWriter, r *http.Request, env Envelope, in renderInput) {
	in.Auth = s.AuthEnabled()
	if in.Now.IsZero() {
		in.Now = time.Now()
	}
	model := buildRenderModel(env, in)

	s.setUICookie(w, cookieLayout, model.Layout.ID)
	if model.Selected != nil {
		s.setUICookie(w, cookieAccount, model.Selected.AccountID)
	} else {
		s.setUICookie(w, cookieAccount, "")
	}
	s.setUICookie(w, cookieFilter, model.Filter)
	s.setUICookie(w, cookieExpanded, model.ExpandedKey)
	s.setUICookie(w, cookieView, model.MobileView)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if err := executeTemplate(w, "app", model); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
