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
	filter := s.resolveFilter(r)
	if r.URL.Query().Has("q") {
		s.setUICookie(w, cookieFilter, filter)
	}
	if r.URL.Query().Has("layout") {
		s.setUICookie(w, cookieLayout, s.resolveLayout(r))
	}
	if r.URL.Query().Has("account") {
		s.setUICookie(w, cookieAccount, strings.TrimSpace(r.URL.Query().Get("account")))
	}
	if r.URL.Query().Has("view") {
		s.setUICookie(w, cookieView, strings.TrimSpace(r.URL.Query().Get("view")))
	}
	data := shellData{
		Filter:     filter,
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
	refresh := r.URL.Query().Get("refresh") == "1"
	env, err := s.collector.envelopeRefresh(refresh, account)
	if err != nil {
		env = Envelope{Error: err.Error()}
	}
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
	theme := strings.TrimSpace(r.PostFormValue("theme"))
	if theme == "" {
		theme = strings.TrimSpace(r.FormValue("theme"))
	}
	dir := strings.TrimSpace(r.PostFormValue("direction"))
	if dir == "" {
		dir = strings.TrimSpace(r.FormValue("direction"))
	}
	backward := strings.EqualFold(dir, "backward") || strings.EqualFold(dir, "prev") ||
		r.FormValue("backward") == "true" || r.FormValue("backward") == "1" ||
		r.FormValue("prev") == "true" || r.FormValue("prev") == "1"
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
	mode := strings.TrimSpace(r.PostFormValue("usage_mode"))
	if mode == "" {
		mode = strings.TrimSpace(r.FormValue("usage_mode"))
	}
	env, err := s.applyUsageMode(mode)
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
	current := readUICookie(r, cookieLayout)
	if current == "" {
		current = s.resolveLayout(r)
	}
	target := strings.TrimSpace(r.PostFormValue("layout"))
	if target == "" {
		target = strings.TrimSpace(r.FormValue("layout"))
	}
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
		if !tui.SetThemeByName(targetTheme) {
			cleaned := strings.ReplaceAll(targetTheme, "-", " ")
			if !tui.SetThemeByName(cleaned) {
				for _, name := range tui.AvailableThemeNames() {
					slug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
					if strings.EqualFold(slug, targetTheme) || strings.EqualFold(name, targetTheme) {
						tui.SetThemeByName(name)
						break
					}
				}
			}
		}
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

// handleProvidersFragment renders the providers/boxes dropdown list.
func (s *Server) handleProvidersFragment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	if !s.checkAuth(w, r) {
		return
	}
	env := s.envelopeOrError()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	data := providerPanelData{Groups: s.providerPanelGroups(env)}
	if err := executeTemplate(w, "providers", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleProviderAction hides, shows, or removes one account and re-renders the
// dashboard plus an out-of-band refresh of the open dropdown.
func (s *Server) handleProviderAction(w http.ResponseWriter, r *http.Request) {
	if !s.requireUIAction(w, r) {
		return
	}
	_ = r.ParseForm()
	op := strings.ToLower(strings.TrimSpace(r.PostFormValue("op")))
	if op == "" {
		op = strings.ToLower(strings.TrimSpace(r.FormValue("op")))
	}
	accountID := strings.TrimSpace(r.PostFormValue("account_id"))
	if accountID == "" {
		accountID = strings.TrimSpace(r.FormValue("account_id"))
	}
	if accountID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "account_id is required"})
		return
	}

	var err error
	toast := ""
	switch op {
	case "hide":
		err = s.setAccountVisibility(accountID, false)
		toast = "Hidden: " + accountID
	case "show":
		err = s.setAccountVisibility(accountID, true)
		toast = "Visible: " + accountID
	case "delete", "remove":
		err = s.deleteAccount(accountID)
		toast = "Removed: " + accountID
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unknown op " + op})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	env := s.envelopeOrError()
	in := renderInput{
		Layout:          s.resolveLayout(r),
		Filter:          s.resolveFilter(r),
		Account:         readUICookie(r, cookieAccount),
		ExpandedAccount: readUICookie(r, cookieExpanded),
		MobileView:      readUICookie(r, cookieView),
		Toast:           toast,
	}
	model := s.appModel(env, in)
	s.setAppCookies(w, model)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if err := executeTemplate(w, "app", model); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := providerPanelData{Groups: s.providerPanelGroups(env)}
	if err := executeTemplate(w, "providers-oob", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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
	model := s.appModel(env, in)
	s.setAppCookies(w, model)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if err := executeTemplate(w, "app", model); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// appModel builds the dashboard render model for one request.
func (s *Server) appModel(env Envelope, in renderInput) renderModel {
	in.Auth = s.AuthEnabled()
	if in.Now.IsZero() {
		in.Now = time.Now()
	}
	return buildRenderModel(env, in)
}

// setAppCookies persists the dashboard UI state the request resolved to.
func (s *Server) setAppCookies(w http.ResponseWriter, model renderModel) {
	s.setUICookie(w, cookieLayout, model.Layout.ID)
	if model.Selected != nil {
		s.setUICookie(w, cookieAccount, model.Selected.AccountID)
	} else {
		s.setUICookie(w, cookieAccount, "")
	}
	s.setUICookie(w, cookieFilter, model.Filter)
	s.setUICookie(w, cookieExpanded, model.ExpandedKey)
	s.setUICookie(w, cookieView, model.MobileView)
}
