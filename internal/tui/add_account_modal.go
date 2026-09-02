package tui

import (
	"fmt"
	"log"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/providers"
)

type addAccountField int

const (
	addAccountFieldProvider addAccountField = iota
	addAccountFieldAccountID
	addAccountFieldAuthType
	addAccountFieldCredential
	addAccountFieldSubmit
	addAccountFieldCount
)

type addAccountAuthMode int

const (
	addAccountAuthDirectKey addAccountAuthMode = iota
	addAccountAuthEnvVar
)

type addAccountProviderItem struct {
	ID         string
	Name       string
	Spec       core.ProviderSpec
	AuthTypes  []core.ProviderAuthType
	DocURL     string
	Quickstart []string
}

type addAccountModalState struct {
	active          bool
	field           addAccountField
	providerCursor  int
	providerList    []addAccountProviderItem
	accountID       string
	accountIDEdited bool
	authTypeIdx     int
	authModes       []core.ProviderAuthType
	authMode        addAccountAuthMode
	apiKey          string
	apiKeyEnv       string
	status          string
	validating      bool
	browserPicker   browserPickerState
}

type validateNewAccountKeyResultMsg struct {
	Account core.AccountConfig
	APIKey  string
	Valid   bool
	Error   string
}

type accountSavedMsg struct {
	Account core.AccountConfig
	Err     error
}

func allAddAccountProviders() []addAccountProviderItem {
	all := providers.AllProviders()
	items := make([]addAccountProviderItem, 0, len(all))

	for _, p := range all {
		spec := p.Spec()
		id := spec.ID
		if id == "" {
			id = p.ID()
		}
		name := spec.Info.Name
		if name == "" {
			name = id
		}

		authTypes := make([]core.ProviderAuthType, 0, 2)
		primary := spec.Auth.Type
		if primary == "" && spec.Auth.APIKeyEnv != "" {
			primary = core.ProviderAuthTypeAPIKey
		}
		if primary != "" {
			authTypes = append(authTypes, primary)
		}
		for _, sup := range spec.Auth.SupplementalTypes {
			if sup != "" && !containsAuthType(authTypes, sup) {
				authTypes = append(authTypes, sup)
			}
		}
		if len(authTypes) == 0 {
			authTypes = append(authTypes, core.ProviderAuthTypeAPIKey)
		}

		docURL := spec.Setup.DocsURL
		if docURL == "" {
			docURL = spec.Info.DocURL
		}

		items = append(items, addAccountProviderItem{
			ID:         id,
			Name:       name,
			Spec:       spec,
			AuthTypes:  authTypes,
			DocURL:     docURL,
			Quickstart: spec.Setup.Quickstart,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})

	return items
}

func containsAuthType(list []core.ProviderAuthType, item core.ProviderAuthType) bool {
	for _, a := range list {
		if a == item {
			return true
		}
	}
	return false
}

func (m Model) defaultAccountIDForProvider(item addAccountProviderItem) string {
	base := strings.TrimSpace(item.Spec.Auth.DefaultAccountID)
	if base == "" {
		base = item.ID
	}
	if !m.accountExists(base) {
		return base
	}
	for i := 2; i <= 100; i++ {
		candidate := fmt.Sprintf("%s-%d", base, i)
		if !m.accountExists(candidate) {
			return candidate
		}
	}
	return base + "-new"
}

func (m Model) accountExists(id string) bool {
	id = strings.TrimSpace(id)
	if id == "" {
		return false
	}
	if _, ok := m.accountProviders[id]; ok {
		return true
	}
	for _, p := range m.providerOrder {
		if p == id {
			return true
		}
	}
	return false
}

func (m *Model) openAddAccountModal() {
	list := allAddAccountProviders()
	cursor := 0

	currentProvider := ""
	if m.settings.show && len(m.providerOrder) > 0 {
		selID := m.providerOrder[clamp(m.settings.cursor, 0, len(m.providerOrder)-1)]
		currentProvider = m.accountProviders[selID]
	} else if len(m.sortedIDs) > 0 {
		selID := m.sortedIDs[clamp(m.cursor, 0, len(m.sortedIDs)-1)]
		currentProvider = m.accountProviders[selID]
	}

	if currentProvider != "" {
		for i, item := range list {
			if item.ID == currentProvider {
				cursor = i
				break
			}
		}
	}

	m.settings.addAccount = addAccountModalState{
		active:          true,
		field:           addAccountFieldProvider,
		providerCursor:  cursor,
		providerList:    list,
		accountIDEdited: false,
	}
	m.syncAddAccountProviderState()
}

func (m *Model) closeAddAccountModal() {
	m.settings.addAccount = addAccountModalState{}
}

func (m *Model) syncAddAccountProviderState() {
	state := &m.settings.addAccount
	if len(state.providerList) == 0 {
		return
	}
	state.providerCursor = clamp(state.providerCursor, 0, len(state.providerList)-1)
	item := state.providerList[state.providerCursor]

	if !state.accountIDEdited {
		state.accountID = m.defaultAccountIDForProvider(item)
	}
	state.authModes = item.AuthTypes
	state.authTypeIdx = 0
	state.apiKeyEnv = item.Spec.Auth.APIKeyEnv
	state.apiKey = ""
	state.status = ""
	state.validating = false
}

func (m Model) currentAddAccountAuthType() core.ProviderAuthType {
	state := m.settings.addAccount
	if len(state.authModes) == 0 {
		return core.ProviderAuthTypeAPIKey
	}
	idx := clamp(state.authTypeIdx, 0, len(state.authModes)-1)
	return state.authModes[idx]
}

func (m Model) handleAddAccountKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	state := &m.settings.addAccount
	if state.browserPicker.active {
		return m.handleAddAccountBrowserPickerKey(msg)
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.closeAddAccountModal()
		return m, nil
	case "tab":
		m.advanceAddAccountField(1)
		return m, nil
	case "shift+tab":
		m.advanceAddAccountField(-1)
		return m, nil
	}

	switch state.field {
	case addAccountFieldProvider:
		switch msg.String() {
		case "up", "k", "left", "h":
			if state.providerCursor > 0 {
				state.providerCursor--
			} else {
				state.providerCursor = len(state.providerList) - 1
			}
			m.syncAddAccountProviderState()
			return m, nil
		case "down", "j", "right", "l":
			if state.providerCursor < len(state.providerList)-1 {
				state.providerCursor++
			} else {
				state.providerCursor = 0
			}
			m.syncAddAccountProviderState()
			return m, nil
		case "enter":
			state.field = addAccountFieldAccountID
			return m, nil
		}

	case addAccountFieldAccountID:
		switch msg.String() {
		case "backspace":
			if len(state.accountID) > 0 {
				state.accountID = state.accountID[:len(state.accountID)-1]
				state.accountIDEdited = true
				state.status = ""
			}
			return m, nil
		case "enter":
			if len(state.authModes) > 1 {
				state.field = addAccountFieldAuthType
			} else {
				state.field = addAccountFieldCredential
			}
			return m, nil
		default:
			if msg.Type == tea.KeyRunes {
				state.accountID += string(msg.Runes)
				state.accountIDEdited = true
				state.status = ""
			}
			return m, nil
		}

	case addAccountFieldAuthType:
		switch msg.String() {
		case "left", "h", "up", "k":
			if state.authTypeIdx > 0 {
				state.authTypeIdx--
			} else {
				state.authTypeIdx = len(state.authModes) - 1
			}
			state.status = ""
			return m, nil
		case "right", "l", "down", "j", " ":
			if state.authTypeIdx < len(state.authModes)-1 {
				state.authTypeIdx++
			} else {
				state.authTypeIdx = 0
			}
			state.status = ""
			return m, nil
		case "enter":
			state.field = addAccountFieldCredential
			return m, nil
		}

	case addAccountFieldCredential:
		authType := m.currentAddAccountAuthType()
		if authType == core.ProviderAuthTypeAPIKey {
			switch msg.String() {
			case "left", "right":
				if state.authMode == addAccountAuthDirectKey {
					state.authMode = addAccountAuthEnvVar
				} else {
					state.authMode = addAccountAuthDirectKey
				}
				state.status = ""
				return m, nil
			case "backspace":
				if state.authMode == addAccountAuthDirectKey {
					if len(state.apiKey) > 0 {
						state.apiKey = state.apiKey[:len(state.apiKey)-1]
						state.status = ""
					}
				} else {
					if len(state.apiKeyEnv) > 0 {
						state.apiKeyEnv = state.apiKeyEnv[:len(state.apiKeyEnv)-1]
						state.status = ""
					}
				}
				return m, nil
			case "enter":
				return m.submitAddAccount()
			default:
				if msg.Type == tea.KeyRunes {
					if state.authMode == addAccountAuthDirectKey {
						state.apiKey += string(msg.Runes)
					} else {
						state.apiKeyEnv += string(msg.Runes)
					}
					state.status = ""
				}
				return m, nil
			}
		} else if authType == core.ProviderAuthTypeBrowserSession {
			switch msg.String() {
			case "enter", " ":
				return m.submitAddAccount()
			}
		} else {
			switch msg.String() {
			case "enter", " ":
				return m.submitAddAccount()
			}
		}

	case addAccountFieldSubmit:
		switch msg.String() {
		case "enter", " ":
			return m.submitAddAccount()
		}
	}

	return m, nil
}

func (m *Model) advanceAddAccountField(delta int) {
	state := &m.settings.addAccount
	fields := []addAccountField{
		addAccountFieldProvider,
		addAccountFieldAccountID,
	}
	if len(state.authModes) > 1 {
		fields = append(fields, addAccountFieldAuthType)
	}
	fields = append(fields, addAccountFieldCredential, addAccountFieldSubmit)

	currentIdx := 0
	for i, f := range fields {
		if f == state.field {
			currentIdx = i
			break
		}
	}

	nextIdx := (currentIdx + delta + len(fields)) % len(fields)
	state.field = fields[nextIdx]
}

func (m Model) submitAddAccount() (tea.Model, tea.Cmd) {
	state := &m.settings.addAccount
	if state.validating {
		return m, nil
	}

	accountID := strings.TrimSpace(state.accountID)
	if accountID == "" {
		state.status = "account ID cannot be empty"
		state.field = addAccountFieldAccountID
		return m, nil
	}

	if len(state.providerList) == 0 {
		state.status = "no provider selected"
		return m, nil
	}
	item := state.providerList[state.providerCursor]
	providerID := item.ID
	authType := m.currentAddAccountAuthType()

	acct := core.AccountConfig{
		ID:       accountID,
		Provider: providerID,
		Auth:     string(authType),
	}

	switch authType {
	case core.ProviderAuthTypeAPIKey:
		apiKey := strings.TrimSpace(state.apiKey)
		apiKeyEnv := strings.TrimSpace(state.apiKeyEnv)
		acct.APIKeyEnv = apiKeyEnv

		if state.authMode == addAccountAuthDirectKey && apiKey != "" {
			state.validating = true
			state.status = "validating API key..."
			return m, m.validateNewAccountKeyCmd(acct, apiKey)
		}

		if apiKeyEnv == "" && apiKey == "" {
			state.status = "API key or env var name required"
			state.field = addAccountFieldCredential
			return m, nil
		}

		state.status = "saving account..."
		return m, m.saveNewAccountCmd(acct, apiKey)

	case core.ProviderAuthTypeBrowserSession:
		domain, cookieName, _ := browserCookieRefForProvider(providerID)
		if domain == "" {
			domain = item.Spec.Auth.BrowserCookieDomain
		}
		if cookieName == "" {
			cookieName = item.Spec.Auth.BrowserCookieName
		}
		acct.BrowserCookie = &core.BrowserCookieRef{
			Domain:     domain,
			CookieName: cookieName,
		}
		if m.services != nil {
			_ = m.services.SaveAccount(acct)
		}
		m.applyAddedAccount(acct)
		m.settings.addAccount.browserPicker = browserPickerState{
			active:     true,
			accountID:  accountID,
			domain:     domain,
			cookieName: cookieName,
			loading:    true,
			status:     "looking for installed browsers...",
		}
		return m, m.loadAvailableBrowsersCmd(accountID)

	default:
		state.status = "saving account..."
		return m, m.saveNewAccountCmd(acct, "")
	}
}

func (m Model) validateNewAccountKeyCmd(acct core.AccountConfig, apiKey string) tea.Cmd {
	return func() tea.Msg {
		if m.services == nil {
			return validateNewAccountKeyResultMsg{
				Account: acct,
				APIKey:  apiKey,
				Valid:   false,
				Error:   "validation service unavailable",
			}
		}
		valid, errMsg := m.services.ValidateAPIKey(acct.ID, acct.Provider, apiKey)
		return validateNewAccountKeyResultMsg{
			Account: acct,
			APIKey:  apiKey,
			Valid:   valid,
			Error:   errMsg,
		}
	}
}

func (m Model) saveNewAccountCmd(acct core.AccountConfig, apiKey string) tea.Cmd {
	return func() tea.Msg {
		if m.services != nil {
			if apiKey != "" {
				if err := m.services.SaveCredential(acct.ID, apiKey); err != nil {
					return accountSavedMsg{Account: acct, Err: fmt.Errorf("save credential: %w", err)}
				}
			}
			if err := m.services.SaveAccount(acct); err != nil {
				return accountSavedMsg{Account: acct, Err: fmt.Errorf("save account: %w", err)}
			}
		}
		return accountSavedMsg{Account: acct, Err: nil}
	}
}

func (m Model) saveAccountCmd(acct core.AccountConfig) tea.Cmd {
	return func() tea.Msg {
		if m.services == nil {
			return accountSavedMsg{Account: acct, Err: fmt.Errorf("account service unavailable")}
		}
		err := m.services.SaveAccount(acct)
		if err != nil {
			log.Printf("save account (%s): %v", acct.ID, err)
		}
		return accountSavedMsg{Account: acct, Err: err}
	}
}

func (m *Model) applyAddedAccount(acct core.AccountConfig) {
	m.ensureProviderTracking()
	accountID := acct.ID
	providerID := acct.Provider

	m.accountProviders[accountID] = providerID
	if m.providerOrderIndex(accountID) < 0 {
		m.providerOrder = append(m.providerOrder, accountID)
		m.providerEnabled[accountID] = true
	}
	if m.onAddAccount != nil {
		m.onAddAccount(acct)
	}
	m.rebuildSortedIDs()
	m.invalidateRenderCaches()
}

func (m Model) handleAddAccountBrowserPickerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	picker := &m.settings.addAccount.browserPicker
	switch msg.String() {
	case "esc", "q":
		m.settings.addAccount.browserPicker = browserPickerState{}
		m.closeAddAccountModal()
		m.settings.status = "browser connection cancelled"
		return m, nil
	case "up", "k":
		if picker.cursor > 0 {
			picker.cursor--
		}
		return m, nil
	case "down", "j":
		if picker.cursor < len(picker.browsers)-1 {
			picker.cursor++
		}
		return m, nil
	case "enter", " ":
		if picker.loading || len(picker.browsers) == 0 {
			return m, nil
		}
		choice := picker.browsers[clamp(picker.cursor, 0, len(picker.browsers)-1)]
		account := picker.accountID
		domain := picker.domain
		cookieName := picker.cookieName
		m.settings.addAccount.browserPicker = browserPickerState{}
		m.closeAddAccountModal()
		m.settings.status = fmt.Sprintf("reading cookie from %s...", choice)
		return m, m.connectBrowserSessionCmd(account, domain, cookieName, choice)
	}
	return m, nil
}

func (m Model) renderAddAccountModalOverlay() string {
	state := m.settings.addAccount
	if state.browserPicker.active {
		return m.renderBrowserPicker(m.width-20, 18)
	}

	contentW := m.width - 20
	if contentW < 64 {
		contentW = 64
	}
	if contentW > 86 {
		contentW = 86
	}
	innerW := contentW - 6

	title := lipgloss.NewStyle().Bold(true).Foreground(colorRosewater).Render("✦ Add Provider Account")
	sub := dimStyle.Render("Configure a new provider account · Esc: Cancel")

	var lines []string
	lines = append(lines, title, sub, settingsBodyRule(innerW+2), "")

	if len(state.providerList) == 0 {
		lines = append(lines, dimStyle.Render("No providers registered."))
		panel := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Background(colorBase).
			Padding(1, 2).
			Width(contentW).
			Render(strings.Join(lines, "\n"))
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, panel)
	}

	item := state.providerList[state.providerCursor]
	authType := m.currentAddAccountAuthType()

	// Field 1: Provider Picker
	f1Prefix := "  "
	f1LabelStyle := lipgloss.NewStyle().Foreground(colorSubtext)
	if state.field == addAccountFieldProvider {
		f1Prefix = lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render("➤ ")
		f1LabelStyle = lipgloss.NewStyle().Bold(true).Foreground(colorLavender)
	}
	pColor := providerThemeColor(item.ID)
	pNameStyled := lipgloss.NewStyle().Bold(true).Foreground(pColor).Render(item.Name)
	authBadge := dimStyle.Render(fmt.Sprintf("[%s]", string(authType)))
	pNav := fmt.Sprintf("◀ %s (%s) %s ▶", pNameStyled, item.ID, authBadge)
	lines = append(lines, fmt.Sprintf("%s%-14s %s", f1Prefix, f1LabelStyle.Render("Provider:"), pNav))
	if item.DocURL != "" && contentW > 50 {
		lines = append(lines, fmt.Sprintf("    %-12s %s", "", dimStyle.Render(truncateToWidth(item.DocURL, innerW-16))))
	}
	lines = append(lines, "")

	// Field 2: Account ID
	f2Prefix := "  "
	f2LabelStyle := lipgloss.NewStyle().Foreground(colorSubtext)
	if state.field == addAccountFieldAccountID {
		f2Prefix = lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render("➤ ")
		f2LabelStyle = lipgloss.NewStyle().Bold(true).Foreground(colorLavender)
	}
	idInput := state.accountID
	cursorChar := ""
	if state.field == addAccountFieldAccountID {
		cursorChar = PulseChar("█", "▌", m.animFrame)
	}
	idInputStyled := lipgloss.NewStyle().Foreground(colorSapphire).Bold(true).Render(idInput + cursorChar)
	idStatus := ""
	if strings.TrimSpace(state.accountID) == "" {
		idStatus = lipgloss.NewStyle().Foreground(colorRed).Render(" (required)")
	} else if m.accountExists(state.accountID) {
		idStatus = lipgloss.NewStyle().Foreground(colorYellow).Render(" (exists · will update)")
	} else {
		idStatus = lipgloss.NewStyle().Foreground(colorGreen).Render(" (new ✓)")
	}
	lines = append(lines, fmt.Sprintf("%s%-14s [%-24s]%s", f2Prefix, f2LabelStyle.Render("Account ID:"), idInputStyled, idStatus), "")

	// Field 3: Auth Type (if >1)
	if len(state.authModes) > 1 {
		f3Prefix := "  "
		f3LabelStyle := lipgloss.NewStyle().Foreground(colorSubtext)
		if state.field == addAccountFieldAuthType {
			f3Prefix = lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render("➤ ")
			f3LabelStyle = lipgloss.NewStyle().Bold(true).Foreground(colorLavender)
		}
		var typePills []string
		for i, at := range state.authModes {
			label := string(at)
			if i == state.authTypeIdx {
				typePills = append(typePills, lipgloss.NewStyle().Bold(true).Foreground(colorMantle).Background(colorAccent).Render(" "+label+" "))
			} else {
				typePills = append(typePills, dimStyle.Render(" "+label+" "))
			}
		}
		lines = append(lines, fmt.Sprintf("%s%-14s %s", f3Prefix, f3LabelStyle.Render("Auth Method:"), strings.Join(typePills, " ")), "")
	}

	// Field 4: Credential / Config
	f4Prefix := "  "
	f4LabelStyle := lipgloss.NewStyle().Foreground(colorSubtext)
	if state.field == addAccountFieldCredential {
		f4Prefix = lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render("➤ ")
		f4LabelStyle = lipgloss.NewStyle().Bold(true).Foreground(colorLavender)
	}

	if authType == core.ProviderAuthTypeAPIKey {
		keyTab := "Direct Key"
		envTab := "Env Var"
		if state.authMode == addAccountAuthDirectKey {
			keyTab = lipgloss.NewStyle().Bold(true).Foreground(colorMantle).Background(colorAccent).Render(" " + keyTab + " ")
			envTab = dimStyle.Render(" " + envTab + " ")
		} else {
			keyTab = dimStyle.Render(" " + keyTab + " ")
			envTab = lipgloss.NewStyle().Bold(true).Foreground(colorMantle).Background(colorAccent).Render(" " + envTab + " ")
		}
		lines = append(lines, fmt.Sprintf("%s%-14s %s │ %s  %s", f4Prefix, f4LabelStyle.Render("Credential:"), keyTab, envTab, dimStyle.Render("(◄ ► to switch)")))

		if state.authMode == addAccountAuthDirectKey {
			keyInput := maskAPIKey(state.apiKey)
			cChar := ""
			if state.field == addAccountFieldCredential {
				cChar = PulseChar("█", "▌", m.animFrame)
			}
			keyStyled := lipgloss.NewStyle().Foreground(colorSapphire).Render(keyInput + cChar)
			lines = append(lines, fmt.Sprintf("    %-12s [%-32s]", f4LabelStyle.Render("API Key:"), keyStyled))
		} else {
			envInput := state.apiKeyEnv
			cChar := ""
			if state.field == addAccountFieldCredential {
				cChar = PulseChar("█", "▌", m.animFrame)
			}
			envStyled := lipgloss.NewStyle().Foreground(colorSapphire).Render(envInput + cChar)
			lines = append(lines, fmt.Sprintf("    %-12s [%-32s]", f4LabelStyle.Render("Env Var:"), envStyled))
		}
	} else if authType == core.ProviderAuthTypeBrowserSession {
		domain, cookieName, _ := browserCookieRefForProvider(item.ID)
		if domain == "" {
			domain = item.Spec.Auth.BrowserCookieDomain
		}
		if cookieName == "" {
			cookieName = item.Spec.Auth.BrowserCookieName
		}
		lines = append(lines,
			fmt.Sprintf("%s%-14s %s · %s", f4Prefix, f4LabelStyle.Render("Browser Auth:"), lipgloss.NewStyle().Foreground(colorTeal).Render(domain), dimStyle.Render(cookieName)),
			fmt.Sprintf("    %-12s %s", "", dimStyle.Render("Will connect to your browser session on save")),
		)
	} else {
		lines = append(lines,
			fmt.Sprintf("%s%-14s %s", f4Prefix, f4LabelStyle.Render("Environment:"), lipgloss.NewStyle().Foreground(colorTeal).Render(string(authType))),
		)
		if len(item.Quickstart) > 0 {
			lines = append(lines, fmt.Sprintf("    %-12s %s", "", dimStyle.Render(truncateToWidth(item.Quickstart[0], innerW-16))))
		}
	}
	lines = append(lines, "")

	// Field 5: Submit Button
	f5Prefix := "  "
	btnText := "  Save Account  "
	if authType == core.ProviderAuthTypeAPIKey && state.authMode == addAccountAuthDirectKey && state.apiKey != "" {
		btnText = "  Validate & Save Account  "
	} else if authType == core.ProviderAuthTypeBrowserSession {
		btnText = "  Connect & Save Account  "
	}

	btnStyle := lipgloss.NewStyle().Bold(true).Foreground(colorSubtext).Border(lipgloss.RoundedBorder()).BorderForeground(colorLine)
	if state.field == addAccountFieldSubmit {
		f5Prefix = lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render("➤ ")
		btnStyle = lipgloss.NewStyle().Bold(true).Foreground(colorMantle).Background(colorGreen).Padding(0, 1)
	}
	lines = append(lines, fmt.Sprintf("%s%s", f5Prefix, btnStyle.Render(btnText)))

	// Status Line
	if state.status != "" {
		sColor := colorSapphire
		if strings.HasPrefix(state.status, "invalid") || strings.Contains(state.status, "error") || strings.Contains(state.status, "required") || strings.Contains(state.status, "cannot") {
			sColor = colorRed
		} else if strings.Contains(state.status, "✓") || strings.Contains(state.status, "saved") {
			sColor = colorGreen
		}
		lines = append(lines, "", "  "+lipgloss.NewStyle().Foreground(sColor).Render(state.status))
	} else {
		lines = append(lines, "")
	}

	lines = append(lines, settingsBodyRule(innerW+2),
		dimStyle.Render("Tab/Shift+Tab: Navigate fields · Enter: Confirm · Esc: Cancel"),
	)

	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorAccent).
		Background(colorBase).
		Padding(1, 2).
		Width(contentW).
		Render(strings.Join(lines, "\n"))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, panel)
}
