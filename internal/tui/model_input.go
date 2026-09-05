package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/nurulislamz/agentusage/internal/core"
)

// applyPersisted is the shared handler for the seven simple "save settings"
// persisted-message types. Each msg type carries only an err; the only
// thing that varies is the status label. Set m.settings.status to either
// failureLabel or successLabel and return the updated model.
func (m Model) applyPersisted(err error, failureLabel, successLabel string) Model {
	if err != nil {
		m.settings.status = failureLabel
	} else {
		m.settings.status = successLabel
	}
	return m
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		return m.handleTickMsg(msg)

	case autoRefreshMsg:
		return m.handleAutoRefresh()

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.invalidateRenderCaches()
		return m, nil

	case DaemonStatusMsg:
		m.daemon.status = msg.Status
		m.daemon.message = msg.Message
		if msg.Status == DaemonRunning {
			m.daemon.installing = false
		}
		return m, m.restartTickIfNeeded()

	case AppUpdateMsg:
		m.daemon.appUpdateCurrent = strings.TrimSpace(msg.CurrentVersion)
		m.daemon.appUpdateLatest = strings.TrimSpace(msg.LatestVersion)
		m.daemon.appUpdateHint = strings.TrimSpace(msg.UpgradeHint)
		return m, nil

	case daemonInstallResultMsg:
		return m.handleDaemonInstallResultMsg(msg)

	case SnapshotsMsg:
		return m.handleSnapshotsMsg(msg)

	case dashboardPrefsPersistedMsg:
		return m.applyPersisted(msg.err, "save failed", "saved"), nil
	case dashboardProviderHideCostsPersistedMsg:
		return m.applyPersisted(msg.err, "hide_costs save failed", "hide_costs saved"), nil
	case dashboardViewPersistedMsg:
		return m.applyPersisted(msg.err, "view save failed", "view saved"), nil
	case dashboardUsageModePersistedMsg:
		return m.applyPersisted(msg.err, "usage mode save failed", "usage mode saved"), nil
	case dashboardWidgetSectionsPersistedMsg:
		return m.applyPersisted(msg.err, "section save failed", "sections saved"), nil
	case detailWidgetSectionsPersistedMsg:
		return m.applyPersisted(msg.err, "detail section save failed", "detail sections saved"), nil
	case dashboardHideSectionsWithNoDataPersistedMsg:
		return m.applyPersisted(msg.err, "empty-state save failed", "empty-state saved"), nil
	case themePersistedMsg:
		return m.applyPersisted(msg.err, "theme save failed", "theme saved"), nil
	case timeWindowPersistedMsg:
		return m.applyPersisted(msg.err, "time window save failed", "time window saved"), nil

	case providerLinkPersistedMsg:
		if msg.err != nil {
			m.settings.providerLinkPicker.status = "save failed: " + msg.err.Error()
			m.settings.status = "provider link save failed"
			return m, nil
		}
		m.settings.providerLinkPicker = providerLinkPickerState{}
		m.settings.status = fmt.Sprintf("mapped %s → %s", msg.source, msg.target)
		m = m.requestRefreshAll()
		return m, nil

	case providerLinkDeletedMsg:
		if msg.err != nil {
			m.settings.providerLinkPicker.status = "clear failed: " + msg.err.Error()
			m.settings.status = "provider link clear failed"
			return m, nil
		}
		m.settings.providerLinkPicker = providerLinkPickerState{}
		m.settings.status = fmt.Sprintf("cleared mapping for %s", msg.source)
		m = m.requestRefreshAll()
		return m, nil

	case validateNewAccountKeyResultMsg:
		if !msg.Valid {
			m.settings.addAccount.validating = false
			errMsg := msg.Error
			if errMsg == "" {
				errMsg = "invalid API key"
			}
			m.settings.addAccount.status = "invalid: " + errMsg
			return m, nil
		}
		m.settings.addAccount.validating = false
		m.applyAddedAccount(msg.Account)
		m.closeAddAccountModal()
		m.settings.status = fmt.Sprintf("account %s added ✓", msg.Account.ID)
		m = m.requestRefreshAll()
		var cmds []tea.Cmd
		if m.services != nil {
			cmds = append(cmds, m.saveCredentialCmd(msg.Account.ID, msg.APIKey), m.saveAccountCmd(msg.Account))
		}
		return m, tea.Batch(cmds...)

	case accountSavedMsg:
		if msg.Err != nil {
			m.settings.addAccount.status = "save failed: " + msg.Err.Error()
			return m, nil
		}
		m.applyAddedAccount(msg.Account)
		m.closeAddAccountModal()
		m.settings.status = fmt.Sprintf("account %s added ✓", msg.Account.ID)
		m = m.requestRefreshAll()
		return m, nil

	case availableBrowsersLoadedMsg:
		// Picker may have been dismissed (esc) before the scan finished —
		// or a fresh open replaced it for a different account. In either
		// case, drop this stale result on the floor.
		if !m.settings.browserPicker.active || m.settings.browserPicker.accountID != msg.AccountID {
			return m, nil
		}
		picker := &m.settings.browserPicker
		picker.loading = false
		if msg.Err != nil {
			picker.status = "could not enumerate browsers: " + msg.Err.Error()
			picker.browsers = nil
			return m, nil
		}
		picker.browsers = msg.Browsers
		picker.cursor = 0
		switch len(msg.Browsers) {
		case 0:
			picker.status = "no supported browsers found on this machine"
		case 1:
			picker.status = "found 1 browser — Enter to read its cookie"
		default:
			picker.status = fmt.Sprintf("found %d browsers — pick the one you log in with", len(msg.Browsers))
		}
		return m, nil

	case browserSessionConnectedMsg:
		if msg.Err != nil {
			m.settings.apiKeyStatus = "connect failed: " + msg.Err.Error()
			m.settings.status = "browser session connect failed for " + msg.AccountID
			return m, nil
		}
		m.ensureProviderTracking()
		providerID := providerForAccountID(msg.AccountID, m.accountProviders)
		if providerID != "" {
			authType := "browser_session"
			acct := core.AccountConfig{
				ID:       msg.AccountID,
				Provider: providerID,
				BrowserCookie: &core.BrowserCookieRef{
					Domain:        msg.Info.Domain,
					CookieName:    msg.Info.CookieName,
					SourceBrowser: msg.Info.SourceBrowser,
				},
			}
			if domain, cookieName, _ := browserCookieRefForProvider(providerID); domain != "" || cookieName != "" {
				acct.BrowserCookie.Domain = core.FirstNonEmpty(domain, acct.BrowserCookie.Domain)
				acct.BrowserCookie.CookieName = core.FirstNonEmpty(cookieName, acct.BrowserCookie.CookieName)
			}
			if isAPIKeyProvider(providerID) {
				authType = "api_key"
				acct.APIKeyEnv = resolvedAPIKeyEnvForProvider(providerID)
			}
			acct.Auth = authType
			if m.onAddAccount != nil {
				m.onAddAccount(acct)
			}
			m.accountProviders[msg.AccountID] = providerID
			if m.providerOrderIndex(msg.AccountID) < 0 {
				m.providerOrder = append(m.providerOrder, msg.AccountID)
				m.providerEnabled[msg.AccountID] = true
			}
		}
		m.settings.apiKeyStatus = fmt.Sprintf("connected via %s", msg.Info.SourceBrowser)
		m.settings.status = fmt.Sprintf("browser session connected for %s", msg.AccountID)
		// Trigger a fresh poll so the tile picks up the new auth path.
		m = m.requestRefreshAll()
		return m, nil

	case browserSessionDisconnectedMsg:
		if msg.Err != nil {
			m.settings.apiKeyStatus = "disconnect failed: " + msg.Err.Error()
			m.settings.status = "browser session disconnect failed for " + msg.AccountID
			return m, nil
		}
		m.settings.apiKeyStatus = "disconnected"
		m.settings.status = fmt.Sprintf("browser session removed for %s", msg.AccountID)
		m = m.requestRefreshAll()
		return m, nil

	case antigravityBoxesLoadedMsg:
		m.settings.boxes.loading = false
		if msg.Err != nil {
			m.settings.boxes.status = "error loading boxes: " + msg.Err.Error()
			return m, nil
		}
		m.settings.boxes.boxes = msg.Boxes
		m.settings.boxes.cursor = clamp(m.settings.boxes.cursor, 0, len(msg.Boxes)-1)
		return m, nil

	case antigravityBoxCreatedMsg:
		m.settings.boxes.loading = false
		if msg.Err != nil {
			m.settings.boxes.status = "create box failed: " + msg.Err.Error()
			return m, nil
		}
		m.settings.boxes.status = fmt.Sprintf("created box %q", msg.Name)
		acctID := "antigravity-" + msg.Name
		if m.providerOrderIndex(acctID) < 0 {
			m.providerOrder = append(m.providerOrder, acctID)
			m.providerEnabled[acctID] = true
			m.accountProviders[acctID] = "antigravity"
		}
		m = m.requestRefreshAll()
		return m, m.loadAntigravityBoxesCmd()

	case antigravityBoxDeletedMsg:
		m.settings.boxes.loading = false
		if msg.Err != nil {
			m.settings.boxes.status = "delete box failed: " + msg.Err.Error()
			return m, nil
		}
		m.settings.boxes.status = fmt.Sprintf("deleted box %q", msg.Name)
		acctID := "antigravity-" + msg.Name
		m.providerEnabled[acctID] = false
		m = m.requestRefreshAll()
		return m, m.loadAntigravityBoxesCmd()

	case antigravityBoxOAuthURLMsg:
		m.settings.boxes.status = fmt.Sprintf("OAuth URL detected for %s: opening browser...", msg.BoxName)
		return m, nil

	case antigravityBoxLoggedInMsg:
		m.settings.boxes.loggingIn = false
		m.settings.boxes.loginTarget = ""
		if msg.Err != nil {
			m.settings.boxes.status = "login failed: " + msg.Err.Error()
			return m, nil
		}
		m.settings.boxes.status = fmt.Sprintf("logged in box %q!", msg.BoxName)
		acctID := "antigravity-" + msg.BoxName
		if m.providerOrderIndex(acctID) < 0 {
			m.providerOrder = append(m.providerOrder, acctID)
			m.providerEnabled[acctID] = true
			m.accountProviders[acctID] = "antigravity"
		}
		m = m.requestRefreshAll()
		return m, m.loadAntigravityBoxesCmd()

	case providerConsoleOpenedMsg:
		if msg.Err != nil {
			m.settings.apiKeyStatus = "open browser failed: " + msg.Err.Error()
			return m, nil
		}
		m.settings.apiKeyStatus = "opened browser — log in, then press Enter to read cookie"
		return m, nil

	case validateKeyResultMsg:
		return m.handleValidateKeyResultMsg(msg)

	case credentialSavedMsg:
		return m.handleCredentialSavedMsg(msg)

	case credentialDeletedMsg:
		if msg.Err != nil {
			m.settings.status = "delete failed"
		} else {
			m.settings.status = "key deleted"
		}
		return m, nil

	case integrationInstallResultMsg:
		return m.handleIntegrationInstallResultMsg(msg)

	case tea.KeyMsg:
		m.lastInteraction = time.Now()
		cmd := m.restartTickIfNeeded()
		if m.settings.show {
			mdl, keyCmd := m.handleSettingsModalKey(msg)
			return mdl, tea.Batch(cmd, keyCmd)
		}
		if !m.hasData {
			mdl, keyCmd := m.handleSplashKey(msg)
			return mdl, tea.Batch(cmd, keyCmd)
		}
		mdl, keyCmd := m.handleKey(msg)
		return mdl, tea.Batch(cmd, keyCmd)

	case tea.MouseMsg:
		m.lastInteraction = time.Now()
		cmd := m.restartTickIfNeeded()
		mdl, mouseCmd := m.handleMouse(msg)
		return mdl, tea.Batch(cmd, mouseCmd)
	}
	return m, nil
}

func (m Model) handleTickMsg(_ tickMsg) (tea.Model, tea.Cmd) {
	m.animFrame++
	interval := m.nextTickInterval()
	if interval == 0 {
		m.tickRunning = false
		return m, nil
	}
	return m, scheduleTickCmd(interval)
}

func (m Model) handleAutoRefresh() (tea.Model, tea.Cmd) {
	// Background refresh: pull latest data without blocking the UI on "Fetching...".
	if !m.refreshing && m.onRefresh != nil {
		_ = m.onRefresh(RefreshRequest{TimeWindow: m.timeWindow})
	}
	return m, autoRefreshCmd(m.refreshInterval)
}

func (m Model) handleDaemonInstallResultMsg(msg daemonInstallResultMsg) (tea.Model, tea.Cmd) {
	m.daemon.installing = false
	if msg.err != nil {
		m.daemon.status = DaemonError
		m.daemon.message = msg.err.Error()
	} else {
		m.daemon.installDone = true
		m.daemon.status = DaemonStarting
	}
	return m, nil
}

func (m Model) handleSnapshotsMsg(msg SnapshotsMsg) (tea.Model, tea.Cmd) {
	msgWindow := msg.TimeWindow
	if msgWindow == "" {
		msgWindow = core.TimeWindow30d
	}
	if msgWindow != m.timeWindow {
		return m, nil
	}

	pendingRefresh := m.refreshing && m.pendingRefreshRequestID > 0 && msg.RequestID == m.pendingRefreshRequestID

	if msg.RequestID > 0 && msg.RequestID < m.lastSnapshotRequestID && !pendingRefresh {
		return m, nil
	}
	if m.refreshing && m.pendingRefreshRequestID > 0 && !pendingRefresh {
		if len(msg.Snapshots) > 0 && snapshotsReady(msg.Snapshots) {
			m.snapshots = msg.Snapshots
			m.lastDataUpdate = time.Now()
			m.invalidateRenderCaches()
			if msg.RequestID > m.lastSnapshotRequestID {
				m.lastSnapshotRequestID = msg.RequestID
			}
		}
		return m, m.restartTickIfNeeded()
	}
	if pendingRefresh {
		if snapshotsReady(msg.Snapshots) {
			m.snapshots = msg.Snapshots
			m.lastDataUpdate = time.Now()
			m.invalidateRenderCaches()
		}
		m.refreshing = false
		m.refreshAll = false
		m.pendingRefreshRequestID = 0
		if msg.RequestID > m.lastSnapshotRequestID {
			m.lastSnapshotRequestID = msg.RequestID
		}
		if len(m.snapshots) > 0 || snapshotsReady(m.snapshots) {
			m.hasData = true
			m.daemon.status = DaemonRunning
		}
		for id, snap := range m.snapshots {
			info := computeDisplayInfo(snap, dashboardWidget(snap.ProviderID), m.resolveHideCosts(snap))
			if info.reason != "" {
				snap.EnsureMaps()
				snap.Diagnostics["display_branch"] = info.reason
				m.snapshots[id] = snap
			}
		}
		m.ensureSnapshotProvidersKnown()
		m.rebuildSortedIDs()
		return m, m.restartTickIfNeeded()
	}
	if m.refreshing && m.hasData && !snapshotsReady(msg.Snapshots) {
		return m, nil
	}
	m.snapshots = msg.Snapshots
	m.refreshing = false
	m.refreshAll = false
	m.pendingRefreshRequestID = 0
	m.lastDataUpdate = time.Now()
	m.invalidateRenderCaches()
	if msg.RequestID > m.lastSnapshotRequestID {
		m.lastSnapshotRequestID = msg.RequestID
	}
	if len(msg.Snapshots) > 0 || snapshotsReady(msg.Snapshots) {
		m.hasData = true
		m.daemon.status = DaemonRunning
	}
	for id, snap := range m.snapshots {
		info := computeDisplayInfo(snap, dashboardWidget(snap.ProviderID), m.resolveHideCosts(snap))
		if info.reason != "" {
			snap.EnsureMaps()
			snap.Diagnostics["display_branch"] = info.reason
			m.snapshots[id] = snap
		}
	}
	m.ensureSnapshotProvidersKnown()
	m.rebuildSortedIDs()
	return m, m.restartTickIfNeeded()
}

func (m Model) handleValidateKeyResultMsg(msg validateKeyResultMsg) (tea.Model, tea.Cmd) {
	if msg.Valid {
		m.settings.apiKeyStatus = "valid ✓ — saving..."
		return m, m.saveCredentialCmd(msg.AccountID, m.settings.apiKeyInput)
	}
	m.settings.apiKeyStatus = "invalid ✗"
	if msg.Error != "" {
		errMsg := msg.Error
		if len(errMsg) > 40 {
			errMsg = errMsg[:37] + "..."
		}
		m.settings.apiKeyStatus = "invalid: " + errMsg
	}
	return m, nil
}

func (m Model) handleCredentialSavedMsg(msg credentialSavedMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		m.settings.apiKeyStatus = "save failed"
	} else {
		m.ensureProviderTracking()
		m.settings.apiKeyStatus = "saved ✓"
		apiKey := m.settings.apiKeyInput
		m.settings.apiKeyEditing = false
		m.settings.apiKeyInput = ""
		if m.onAddAccount != nil {
			providerID := m.accountProviders[msg.AccountID]
			acct := core.AccountConfig{
				ID:       msg.AccountID,
				Provider: providerID,
				Auth:     "api_key",
				Token:    apiKey,
			}
			m.onAddAccount(acct)
		}
		if m.providerOrderIndex(msg.AccountID) < 0 {
			m.providerOrder = append(m.providerOrder, msg.AccountID)
			m.providerEnabled[msg.AccountID] = true
		}
		m.refreshing = true
	}
	return m, nil
}

func (m Model) handleIntegrationInstallResultMsg(msg integrationInstallResultMsg) (tea.Model, tea.Cmd) {
	m.settings.integrationStatus = msg.Statuses
	if msg.Err != nil {
		errMsg := msg.Err.Error()
		if len(errMsg) > 80 {
			errMsg = errMsg[:77] + "..."
		}
		m.settings.status = "integration install failed: " + errMsg
	} else {
		m.settings.status = "integration installed"
	}
	return m, nil
}

func (m Model) handleSplashKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "enter":
		if (m.daemon.status == DaemonNotInstalled || m.daemon.status == DaemonOutdated) && !m.daemon.installing {
			m.daemon.installing = true
			m.daemon.message = "Setting up background helper..."
			return m, m.installDaemonCmd()
		}
	}
	return m, nil
}

func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.showHelp {
		if msg.Action == tea.MouseActionPress {
			m.showHelp = false
			return m, nil
		}
		return m, nil
	}
	if m.settings.show {
		return m.handleSettingsMouse(msg)
	}
	if m.filter.active || m.analyticsFilter.active {
		return m, nil
	}
	if msg.Action != tea.MouseActionPress {
		return m, nil
	}

	switch msg.Button {
	case tea.MouseButtonWheelUp:
		return m.handleMouseScroll(-m.mouseScrollStep(), msg.X), nil
	case tea.MouseButtonWheelDown:
		return m.handleMouseScroll(m.mouseScrollStep(), msg.X), nil
	case tea.MouseButtonLeft:
		return m.handleMouseLeftClick(msg)
	default:
		return m, nil
	}
}

func (m Model) handleMouseScroll(scroll, mouseX int) Model {
	if m.screen != screenDashboard {
		if m.screen == screenAnalytics {
			m.analyticsScrollY += scroll
			if m.analyticsScrollY < 0 {
				m.analyticsScrollY = 0
			}
		}
		return m
	}
	if m.mode == modeDetail {
		m.detailOffset += scroll
		if m.detailOffset < 0 {
			m.detailOffset = 0
		}
		return m
	}
	if m.mode == modeList {
		ids := m.filteredIDs()
		if len(ids) == 0 {
			return m
		}

		if m.activeDashboardView() != dashboardViewSplit {
			step := 1
			if m.activeDashboardView() == dashboardViewBento {
				tileW := 36
				if m.width < 40 {
					tileW = max(28, m.width-4)
				}
				step = max(1, (m.width-2)/(tileW+2))
			}
			if scroll < 0 {
				m.cursor = clamp(m.cursor-step, 0, len(ids)-1)
			} else if scroll > 0 {
				m.cursor = clamp(m.cursor+step, 0, len(ids)-1)
			}
			m.detailOffset = 0
			m.invalidateRenderCaches()
			return m
		}

		leftW := m.width / 3
		if leftW < minLeftWidth {
			leftW = minLeftWidth
		}
		if leftW > maxLeftWidth {
			leftW = maxLeftWidth
		}
		if leftW > m.width-34 {
			leftW = m.width - 34
		}
		if leftW < 10 {
			leftW = m.width / 2
		}

		if mouseX > 0 && mouseX <= leftW && len(ids) > 0 {
			if scroll < 0 && m.cursor > 0 {
				m.cursor--
				m.detailOffset = 0
			} else if scroll > 0 && m.cursor < len(ids)-1 {
				m.cursor++
				m.detailOffset = 0
			}
		} else {
			m.detailOffset += scroll
			if m.detailOffset < 0 {
				m.detailOffset = 0
			}
		}
	}
	return m
}

func (m Model) handleMouseLeftClick(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	headerLines := 2
	footerLines := 2

	// Check if click is on header (e.g. screen tabs)
	if msg.Y < headerLines {
		if msg.Y == 0 {
			screens := m.availableScreens()
			if len(screens) > 1 {
				bolt := PulseChar(
					accentBoldStyle.Render("⚡"),
					lipgloss.NewStyle().Foreground(colorDim).Bold(true).Render("⚡"),
					m.animFrame,
				)
				brandText := RenderGradientText("agentUsage", m.animFrame)
				startX := lipgloss.Width(bolt) + 1 + lipgloss.Width(brandText) + 1

				for i, screen := range screens {
					label := screenLabelByTab[screen]
					tabStr := fmt.Sprintf("%d:%s", i+1, label)
					tabStyle := screenTabInactiveStyle
					if screen == m.screen {
						tabStyle = screenTabActiveStyle
					}
					rendered := tabStyle.Render(tabStr)
					tw := lipgloss.Width(rendered)
					if msg.X >= startX && msg.X < startX+tw {
						if m.screen != screen {
							m.screen = screen
							m.mode = modeList
							m.detailOffset = 0
							m.tileOffset = 0
						}
						return m, nil
					}
					startX += tw
				}
			}
		}
		return m, nil
	}

	// Check if click is on footer
	if msg.Y >= m.height-footerLines {
		return m.handleFooterClick(msg)
	}

	if m.screen != screenDashboard {
		return m, nil
	}

	if m.mode == modeDetail {
		// Clicking at the top header area exits detail view
		if msg.Y <= headerLines+1 {
			m = m.exitDetailMode()
			return m, nil
		}
		return m, nil
	}

	ids := m.filteredIDs()
	if len(ids) == 0 {
		return m, nil
	}

	contentH := m.height - headerLines - footerLines
	if contentH < 1 {
		contentH = 1
	}
	clickContentY := msg.Y - headerLines
	if clickContentY < 0 {
		return m, nil
	}

	var clickedIdx int = -1
	switch m.activeDashboardView() {
	case dashboardViewMatrix:
		clickedIdx = m.matrixHitTest(clickContentY, contentH, ids)
	case dashboardViewBento:
		clickedIdx = m.bentoHitTest(msg.X, clickContentY, contentH, ids)
	case dashboardViewBars, dashboardViewDials, dashboardViewStrips:
		clickedIdx = m.boardHitTest(clickContentY, contentH, ids)
	default: // dashboardViewSplit
		leftW := m.width / 3
		if leftW < minLeftWidth {
			leftW = minLeftWidth
		}
		if leftW > maxLeftWidth {
			leftW = maxLeftWidth
		}
		if leftW > m.width-34 {
			leftW = m.width - 34
		}
		if leftW < 10 {
			leftW = m.width / 2
		}

		if msg.X <= leftW {
			// Clicked in the left navigator list -> select item without entering full detail view
			clickedIdx = m.splitListHitTest(leftW, contentH, clickContentY, ids)
		}
	}

	if clickedIdx >= 0 && clickedIdx < len(ids) {
		if m.activeDashboardView() != dashboardViewSplit && clickedIdx == m.cursor {
			// In non-split views, clicking the already selected card/row enters detail mode
			return m.enterDetailMode(), nil
		}
		m.cursor = clickedIdx
		m.detailOffset = 0
		m.detailTab = 0
		m.tileOffset = 0
		m.invalidateRenderCaches()
	}
	return m, nil
}

func (m Model) matrixHitTest(clickContentY, h int, ids []string) int {
	if len(ids) == 0 {
		return -1
	}
	type itemPos struct {
		globalIdx int
		line      int
	}
	var items []itemPos
	cursorLine := 0
	currentLine := 0

	groups := m.groupIDsByProvider(ids)
	for _, grp := range groups {
		headerLine := currentLine
		currentLine++ // ◈ PROVIDER (N agents) ALL OK
		currentLine++ // Table header: ACCOUNT STATUS QUOTA 1 ...
		for rowIdx := range grp.accountIDs {
			globalIdx := grp.indices[rowIdx]
			if globalIdx == m.cursor {
				cursorLine = currentLine
			}
			items = append(items, itemPos{globalIdx: globalIdx, line: currentLine})
			if rowIdx == 0 {
				items = append(items, itemPos{globalIdx: globalIdx, line: headerLine})
			}
			currentLine++
		}
		currentLine++ // Empty line between provider groups
	}

	totalLines := currentLine
	start := 0
	if totalLines > h && h > 0 {
		start = cursorLine - (h / 2)
		if start < 0 {
			start = 0
		}
		if start+h > totalLines {
			start = totalLines - h
			if start < 0 {
				start = 0
			}
		}
	}

	targetLine := start + clickContentY
	for _, it := range items {
		if it.line == targetLine {
			return it.globalIdx
		}
	}
	return -1
}

func (m Model) bentoHitTest(clickX, clickContentY, h int, ids []string) int {
	if len(ids) == 0 {
		return -1
	}
	tileW := 36
	if m.width < 40 {
		tileW = max(28, m.width-4)
	}
	cols := max(1, (m.width-2)/(tileW+2))

	groups := m.groupIDsByProvider(ids)

	type bentoBox struct {
		globalIdx  int
		minY, maxY int
		minX, maxX int
	}
	var boxes []bentoBox
	currentY := 0
	cursorLine := 0

	for _, grp := range groups {
		headerY := currentY
		currentY++ // Header line

		for i := 0; i < len(grp.accountIDs); i += cols {
			end := min(i+cols, len(grp.accountIDs))
			rowAccountIDs := grp.accountIDs[i:end]
			rowIndices := grp.indices[i:end]
			cardH := 8 // 5 content + 2 border + 1 margin

			for colIdx := range rowAccountIDs {
				globalIdx := rowIndices[colIdx]
				if globalIdx == m.cursor {
					cursorLine = currentY
				}
				xStart := colIdx * (tileW + 2)
				boxes = append(boxes, bentoBox{
					globalIdx: globalIdx,
					minY:      currentY,
					maxY:      currentY + cardH,
					minX:      xStart,
					maxX:      xStart + tileW + 2,
				})
				if colIdx == 0 && i == 0 {
					boxes = append(boxes, bentoBox{
						globalIdx: globalIdx,
						minY:      headerY,
						maxY:      headerY + 1,
						minX:      0,
						maxX:      m.width,
					})
				}
			}
			currentY += cardH
		}
		currentY++ // Empty line
	}

	totalLines := currentY
	start := 0
	if totalLines > h && h > 0 {
		start = cursorLine - (h / 3)
		if start < 0 {
			start = 0
		}
		if start+h > totalLines {
			start = totalLines - h
			if start < 0 {
				start = 0
			}
		}
	}

	targetY := start + clickContentY
	for _, b := range boxes {
		if targetY >= b.minY && targetY < b.maxY {
			if cols == 1 || (clickX >= b.minX && clickX < b.maxX) {
				return b.globalIdx
			}
		}
	}
	return -1
}

func (m Model) boardHitTest(clickContentY, h int, ids []string) int {
	if len(ids) == 0 {
		return -1
	}
	groups := m.groupIDsByProvider(ids)
	view := m.activeDashboardView()
	now := m.viewNow()
	cardW := clamp(m.width-4, 30, 90)

	type cardBox struct {
		globalIdx  int
		minY, maxY int
	}
	var boxes []cardBox
	currentY := 0
	cursorLine := 0

	for _, grp := range groups {
		headerY := currentY
		currentY++ // Header line

		for rowIdx, id := range grp.accountIDs {
			globalIdx := grp.indices[rowIdx]
			snap := m.snapshots[id]
			if globalIdx == m.cursor {
				cursorLine = currentY
			}

			cardH := 8
			switch view {
			case dashboardViewBars:
				rendered := m.renderBarCard(snap, false, cardW, now)
				cardH = strings.Count(rendered, "\n") + 1
			case dashboardViewDials:
				rendered := m.renderDialCard(snap, false, cardW, now)
				cardH = strings.Count(rendered, "\n") + 1
			case dashboardViewStrips:
				rendered := m.renderStripRow(snap, false, cardW, now)
				cardH = strings.Count(rendered, "\n") + 1
			}

			boxes = append(boxes, cardBox{
				globalIdx: globalIdx,
				minY:      currentY,
				maxY:      currentY + cardH,
			})
			if rowIdx == 0 {
				boxes = append(boxes, cardBox{
					globalIdx: globalIdx,
					minY:      headerY,
					maxY:      headerY + 1,
				})
			}
			currentY += cardH
		}
		currentY++ // Empty line
	}

	totalLines := currentY
	start := 0
	if totalLines > h && h > 0 {
		start = cursorLine - (h / 3)
		if start < 0 {
			start = 0
		}
		if start+h > totalLines {
			start = totalLines - h
			if start < 0 {
				start = 0
			}
		}
	}

	targetY := start + clickContentY
	for _, b := range boxes {
		if targetY >= b.minY && targetY < b.maxY {
			return b.globalIdx
		}
	}
	return -1
}

func (m Model) handleFooterClick(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Y != m.height-1 {
		return m, nil
	}
	plain := ansi.Strip(m.renderFooterStatusLine(m.width))
	if len(plain) == 0 {
		return m, nil
	}

	x := msg.X

	inToken := func(token string, leftPad, rightPad int) bool {
		idx := strings.Index(plain, token)
		if idx < 0 {
			return false
		}
		start := idx - leftPad
		if start < 0 {
			start = 0
		}
		end := idx + len(token) + rightPad
		return x >= start && x <= end
	}

	// 1. Check "back" (Esc back)
	if inToken("back", 4, 2) {
		if m.mode == modeDetail {
			m = m.exitDetailMode()
			return m, nil
		}
	}

	// 2. Check "menu" (p menu)
	if inToken("menu", 3, 2) {
		m.openSettingsModal()
		m.settings.tab = settingsTabProviders
		return m, nil
	}

	// 3. Check "layout" (v layout)
	if inToken("layout", 3, 2) {
		if m.screen == screenDashboard {
			next := m.nextDashboardView(1)
			m.setDashboardView(next)
			return m, m.persistDashboardViewCmd()
		}
	}

	// 4. Check "mode" (u mode)
	if inToken("mode", 3, 2) {
		if m.screen == screenDashboard {
			m.toggleUsageMode()
			return m, m.persistDashboardUsageModeCmd()
		}
	}

	// 5. Check "refresh all" (R refresh all) - must check before "refresh"
	if inToken("refresh all", 3, 2) {
		return m.triggerRefreshAll()
	}

	// 6. Check "refresh" (r refresh)
	if inToken("refresh", 3, 2) {
		return m.triggerRefreshFocused()
	}

	// 7. Check "help" (? help)
	if inToken("help", 3, 2) {
		m.showHelp = !m.showHelp
		return m, nil
	}

	return m, nil
}

func (m Model) splitListHitTest(w, h, clickContentY int, ids []string) int {
	if len(ids) == 0 {
		return -1
	}
	start, end := listVisibleWindow(m.snapshots, ids, m.cursor, h)
	itemHeight := 3

	currentY := 0
	if start > 0 {
		currentY++ // "▲ more" indicator line
	}

	for i := start; i < end; i++ {
		snap, ok := m.snapshots[ids[i]]
		if !ok {
			continue
		}
		pID := snap.ProviderID
		isFirstInGroup := (i == 0) || (i > 0 && m.snapshots[ids[i-1]].ProviderID != pID)
		if isFirstInGroup {
			if clickContentY == currentY {
				return i
			}
			currentY++
		}
		if clickContentY >= currentY && clickContentY < currentY+itemHeight {
			return i
		}
		currentY += itemHeight
	}

	return -1
}

func (m Model) handleSettingsMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Action != tea.MouseActionPress {
		return m, nil
	}

	switch msg.Button {
	case tea.MouseButtonWheelUp:
		return m.handleSettingsWheel(-1, msg.X, msg.Y)
	case tea.MouseButtonWheelDown:
		return m.handleSettingsWheel(1, msg.X, msg.Y)
	case tea.MouseButtonLeft:
		return m.handleSettingsLeftClick(msg)
	default:
		return m, nil
	}
}

func (m Model) handleSettingsWheel(dir int, mouseX, mouseY int) (tea.Model, tea.Cmd) {
	switch m.settings.tab {
	case settingsTabProviders:
		ids := m.settingsIDs()
		if len(ids) > 0 {
			m.settings.cursor = clamp(m.settings.cursor+dir, 0, len(ids)-1)
		}
	case settingsTabTheme:
		themes := AvailableThemes()
		if len(themes) > 0 {
			m.settings.themeCursor = clamp(m.settings.themeCursor+dir, 0, len(themes)-1)
		}
	case settingsTabAPIKeys:
		ids := m.apiKeysTabIDs()
		if len(ids) > 0 {
			m.settings.cursor = clamp(m.settings.cursor+dir, 0, len(ids)-1)
		}
	case settingsTabWidgetSections:
		if m.width > 0 && mouseX > 0 && mouseX < m.width/2 {
			entries := m.activeSectionEntryCount()
			if entries > 0 {
				m.settings.sectionRowCursor = clamp(m.settings.sectionRowCursor+dir, 0, entries-1)
			}
		} else {
			scroll := m.mouseScrollStep() * dir
			m.settings.previewOffset += scroll
			if m.settings.previewOffset < 0 {
				m.settings.previewOffset = 0
			}
		}
	case settingsTabIntegrations:
		statuses := m.settings.integrationStatus
		if len(statuses) > 0 {
			m.settings.cursor = clamp(m.settings.cursor+dir, 0, len(statuses)-1)
		}
	case settingsTabBoxes:
		boxes := m.settings.boxes.boxes
		if len(boxes) > 0 {
			m.settings.boxes.cursor = clamp(m.settings.boxes.cursor+dir, 0, len(boxes)-1)
		}
	case settingsTabTelemetry:
		scroll := m.mouseScrollStep() * dir
		m.settings.bodyOffset += scroll
		if m.settings.bodyOffset < 0 {
			m.settings.bodyOffset = 0
		}
	}
	return m, nil
}

func (m Model) handleSettingsLeftClick(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.width < 40 || m.height < 12 {
		return m, nil
	}

	contentW := m.width - 24
	if contentW < 68 {
		contentW = 68
	}
	if contentW > 92 {
		contentW = 92
	}
	panelInnerW := contentW - 4
	if panelInnerW < 40 {
		panelInnerW = 40
	}

	const modalBodyHeight = 20
	contentH := modalBodyHeight
	maxAllowed := m.height - 14
	if maxAllowed < 8 {
		maxAllowed = 8
	}
	if contentH > maxAllowed {
		contentH = maxAllowed
	}

	panelW := contentW + 2
	panelH := contentH + 9
	if m.settings.status != "" {
		panelH++
	}

	sideBySide := m.width >= contentW*2+12 && m.settings.tab == settingsTabWidgetSections
	totalModalW := panelW
	if sideBySide {
		totalModalW = contentW*2 + 6
	}

	modalStartX := (m.width - totalModalW) / 2
	modalStartY := (m.height - panelH) / 2

	// Check if clicked outside modal
	if msg.X < modalStartX || msg.X >= modalStartX+totalModalW || msg.Y < modalStartY || msg.Y >= modalStartY+panelH {
		m.settings.show = false
		m.settings.bodyOffset = 0
		m.settings.previewOffset = 0
		return m, nil
	}

	// Check if clicked on tabs line
	tabsY := modalStartY + 3
	if msg.Y == tabsY {
		tabsStartX := modalStartX + 3
		tabsWidth := panelInnerW
		n := len(settingsTabNames)
		if n > 0 && msg.X >= tabsStartX && msg.X < tabsStartX+tabsWidth {
			gap := 1
			cellW := (tabsWidth - gap*(n-1)) / n
			if cellW < 6 {
				cellW = 6
				gap = 0
				cellW = tabsWidth / n
			}
			relX := msg.X - tabsStartX
			stride := cellW + gap
			if stride > 0 {
				clickedTab := relX / stride
				if clickedTab >= 0 && clickedTab < n {
					m.settings.tab = settingsModalTab(clickedTab)
					m.settings.bodyOffset = 0
					m.settings.previewOffset = 0
					return m, nil
				}
			}
		}
		return m, nil
	}

	// Check if clicked inside body area
	bodyStartY := modalStartY + 5
	relY := msg.Y - bodyStartY
	if relY >= 0 && relY < contentH {
		relX := msg.X - (modalStartX + 2)
		return m.handleSettingsBodyLeftClick(relX, relY, panelInnerW, contentH)
	}

	return m, nil
}

func (m Model) handleSettingsBodyLeftClick(relX, relY, panelInnerW, contentH int) (tea.Model, tea.Cmd) {
	switch m.settings.tab {
	case settingsTabProviders:
		ids := m.settingsIDs()
		if len(ids) == 0 {
			return m, nil
		}
		if relY == 1 {
			m.openAddAccountModal()
			return m, nil
		}
		if relY < 4 {
			return m, nil
		}
		cursor := clamp(m.settings.cursor, 0, len(ids)-1)

		// Rebuild item lines mapping to determine clicked account index
		var itemIndices []int
		prevProvider := ""
		cursorLineIdx := 0
		for i, id := range ids {
			providerID := m.accountProviders[id]
			if snap, ok := m.snapshots[id]; ok && snap.ProviderID != "" {
				providerID = snap.ProviderID
			}
			if providerID == "" {
				providerID = "other"
			}
			if providerID != prevProvider {
				if len(itemIndices) > 0 {
					itemIndices = append(itemIndices, -1) // blank spacer
				}
				itemIndices = append(itemIndices, -1) // group header
				prevProvider = providerID
			}
			if i == cursor {
				cursorLineIdx = len(itemIndices)
			}
			itemIndices = append(itemIndices, i)
		}

		availableH := contentH - 4
		if availableH < 3 {
			availableH = 3
		}
		start := 0
		if len(itemIndices) > availableH {
			contentCap := max(1, availableH-2)
			start = max(0, cursorLineIdx-contentCap/2)
			if start+contentCap > len(itemIndices) {
				start = max(0, len(itemIndices)-contentCap)
			}
		}

		clickLine := relY - 4
		if start > 0 {
			clickLine--
		}
		targetLine := start + clickLine
		if targetLine >= 0 && targetLine < len(itemIndices) {
			idx := itemIndices[targetLine]
			if idx >= 0 && idx < len(ids) {
				m.settings.cursor = idx
				id := ids[idx]
				m.providerEnabled[id] = !m.isProviderEnabled(id)
				m.rebuildSortedIDs()
				m.settings.status = "saving settings..."
				return m, m.persistDashboardPrefsCmd()
			}
		}

	case settingsTabTheme:
		themes := AvailableThemes()
		if len(themes) == 0 {
			return m, nil
		}
		if relY >= 5 {
			cursor := clamp(m.settings.themeCursor, 0, len(themes)-1)
			start, end := listWindow(len(themes), cursor, max(1, contentH-5))
			targetIdx := start + (relY - 5)
			if targetIdx >= start && targetIdx < end && targetIdx < len(themes) {
				m.settings.themeCursor = targetIdx
				name := themes[targetIdx].Name
				if SetThemeByName(name) {
					m.invalidateRenderCaches()
					m.settings.status = "saving theme..."
					return m, m.persistThemeCmd(name)
				}
			}
		}

	case settingsTabAPIKeys:
		ids := m.apiKeysTabIDs()
		if len(ids) == 0 {
			return m, nil
		}
		if relY == 1 {
			m.openAddAccountModal()
			return m, nil
		}
		if relY >= 5 {
			cursor := clamp(m.settings.cursor, 0, len(ids)-1)
			start, end := listWindow(len(ids), cursor, max(1, contentH-5))
			targetIdx := start + (relY - 5)
			if targetIdx >= start && targetIdx < end && targetIdx < len(ids) {
				m.settings.cursor = targetIdx
				id := ids[targetIdx]
				providerID := providerForAccountID(id, m.accountProviders)
				if supportsBrowserSessionProvider(providerID) {
					return m.startBrowserSessionConnect(id, providerID)
				}
				m.settings.apiKeyEditing = true
				m.settings.apiKeyEditAccountID = id
				m.settings.apiKeyInput = ""
				m.settings.apiKeyStatus = ""
				return m, nil
			}
		}

	case settingsTabWidgetSections:
		if relY == 0 {
			m.settings.sectionSubTab = 1 - m.settings.sectionSubTab
			m.settings.sectionRowCursor = 0
			m.settings.previewOffset = 0
			return m, nil
		}
		entriesCount := len(m.widgetSectionEntries())
		if m.settings.sectionSubTab == 1 {
			entriesCount = len(m.detailWidgetSectionEntries())
		}
		if relY >= 4 && entriesCount > 0 {
			start, end := listWindow(entriesCount, m.settings.sectionRowCursor, max(1, contentH-5))
			targetIdx := start + (relY - 4)
			if targetIdx >= start && targetIdx < end && targetIdx < entriesCount {
				m.settings.sectionRowCursor = targetIdx
				cmd := m.toggleSelectedActiveSection()
				return m, cmd
			}
		}

	case settingsTabIntegrations:
		statuses := m.settings.integrationStatus
		if len(statuses) == 0 {
			return m, nil
		}
		if relY >= 3 {
			cursor := clamp(m.settings.cursor, 0, len(statuses)-1)
			start, end := listWindow(len(statuses), cursor, max(1, contentH-7))
			targetIdx := start + (relY-3)/2
			if targetIdx >= start && targetIdx < end && targetIdx < len(statuses) {
				m.settings.cursor = targetIdx
				next, cmd, _ := m.handleSettingsTabIntegrationsKey(tea.KeyMsg{Type: tea.KeyEnter})
				return next, cmd
			}
		}

	case settingsTabBoxes:
		boxes := m.settings.boxes.boxes
		if len(boxes) > 0 && relY >= 4 {
			cursor := clamp(m.settings.boxes.cursor, 0, len(boxes)-1)
			start, end := listWindow(len(boxes), cursor, max(1, contentH-6))
			targetIdx := start + (relY - 4)
			if targetIdx >= start && targetIdx < end && targetIdx < len(boxes) {
				m.settings.boxes.cursor = targetIdx
				next, cmd, _ := m.handleSettingsTabBoxesKey(tea.KeyMsg{Type: tea.KeyEnter})
				return next, cmd
			}
		}

	case settingsTabTelemetry:
		rows := m.telemetryRows()
		if len(rows) > 0 && relY >= 3 {
			cursor := m.telemetryRowCursor()
			start, end := listWindow(len(rows), cursor, max(1, contentH-4))
			targetIdx := start + (relY - 3)
			if targetIdx >= start && targetIdx < end && targetIdx < len(rows) {
				m.settings.cursor = targetIdx
				next, cmd, _ := m.activateTelemetryRow(rows)
				return next, cmd
			}
		}
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.settings.addAccount.active {
		return m.handleAddAccountKey(msg)
	}
	if msg.String() == "?" && !m.filter.active && !m.analyticsFilter.active && !m.settings.show {
		m.showHelp = !m.showHelp
		return m, nil
	}
	if m.showHelp {
		m.showHelp = false
		return m, nil
	}
	if m.settings.show {
		return m.handleSettingsModalKey(msg)
	}

	if !m.filter.active && !m.analyticsFilter.active {
		if m.screen == screenDashboard && m.mode == modeDetail {
			switch msg.String() {
			case "tab", "shift+tab", "left", "h", "right", "l":
				return m.handleDetailKey(msg)
			}
		}
		switch msg.String() {
		case "a", "A":
			m.openAddAccountModal()
			return m, nil
		case "p", "P", ",", "S":
			m.openSettingsModal()
			m.settings.tab = settingsTabProviders
			return m, nil

		case "v":
			if m.screen == screenDashboard {
				next := m.nextDashboardView(1)
				m.setDashboardView(next)
				return m, m.persistDashboardViewCmd()
			}
		case "V":
			if m.screen == screenDashboard {
				next := m.nextDashboardView(-1)
				m.setDashboardView(next)
				return m, m.persistDashboardViewCmd()
			}

		case "tab":
			if m.screen == screenDashboard && m.mode != modeDetail {
				next := m.nextDashboardView(1)
				m.setDashboardView(next)
				return m, m.persistDashboardViewCmd()
			}
			m.screen = m.nextScreen(1)
			m.mode = modeList
			m.detailOffset = 0
			m.tileOffset = 0
			return m, nil
		case "shift+tab":
			if m.screen == screenDashboard && m.mode != modeDetail {
				next := m.nextDashboardView(-1)
				m.setDashboardView(next)
				return m, m.persistDashboardViewCmd()
			}
			m.screen = m.nextScreen(-1)
			m.mode = modeList
			m.detailOffset = 0
			m.tileOffset = 0
			return m, nil
		case "t":
			m.invalidateRenderCaches()
			return m, m.persistThemeCmd(CycleTheme())
		case "T":
			m.invalidateRenderCaches()
			return m, m.persistThemeCmd(CycleThemeBackward())
		case "c":
			if m.screen == screenDashboard {
				if mdl, cmd, handled := m.toggleHideCostsOverride(); handled {
					return mdl, cmd
				}
			}
		case "w":
			return m.cycleTimeWindow()
		case "u", "U":
			if m.screen == screenDashboard {
				m.toggleUsageMode()
				return m, m.persistDashboardUsageModeCmd()
			}
		}
	}

	if m.screen == screenAnalytics {
		return m.handleAnalyticsKey(msg)
	}
	return m.handleDashboardTilesKey(msg)
}

// toggleHideCostsOverride cycles the per-account hide_costs override for the
// currently focused dashboard tile/row through auto (nil) → hide (true) →
// show (false) → auto. Returns handled=false when no tile is focused (so the
// keystroke can fall through to other handlers).
func (m Model) toggleHideCostsOverride() (Model, tea.Cmd, bool) {
	ids := m.filteredIDs()
	accountID := m.selectedTileID(ids)
	if accountID == "" {
		return m, nil, false
	}
	next := m.cycleHideCostsOverride(accountID)
	m.invalidateRenderCaches()
	return m, m.persistDashboardProviderHideCostsCmd(accountID, next), true
}

func (m Model) handleDashboardTilesKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.filter.active {
		return m.handleFilterKey(msg)
	}
	if m.mode == modeDetail {
		return m.handleDetailKey(msg)
	}
	return m.handleListKey(msg)
}

func (m Model) handleAnalyticsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.analyticsFilter.active {
		return m.handleAnalyticsFilterKey(msg)
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "s":
		m.analyticsSortBy = (m.analyticsSortBy + 1) % analyticsSortCount
		m.invalidateAnalyticsCache()
	case "/":
		m.analyticsFilter.active = true
		m.analyticsFilter.text = ""
	case "esc":
		if m.analyticsFilter.text != "" {
			m.analyticsFilter.text = ""
			m.invalidateAnalyticsCache()
		} else {
			m.analyticsScrollY = 0
		}
	case "r":
		return m.triggerRefreshFocused()
	case "R":
		return m.triggerRefreshAll()
	case "j", "down":
		m.analyticsScrollY++
	case "k", "up":
		if m.analyticsScrollY > 0 {
			m.analyticsScrollY--
		}
	case "pgdown":
		m.analyticsScrollY += 10
	case "pgup":
		if m.analyticsScrollY > 10 {
			m.analyticsScrollY -= 10
		} else {
			m.analyticsScrollY = 0
		}
	case "home", "g":
		m.analyticsScrollY = 0
	case "end", "G":
		m.analyticsScrollY = 9999
	}
	return m, nil
}

func (m Model) handleAnalyticsFilterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.analyticsFilter.active = false
	case "esc":
		m.analyticsFilter.active = false
		if m.analyticsFilter.text != "" {
			m.analyticsFilter.text = ""
			m.invalidateAnalyticsCache()
		}
	case "backspace":
		if len(m.analyticsFilter.text) > 0 {
			m.analyticsFilter.text = m.analyticsFilter.text[:len(m.analyticsFilter.text)-1]
			m.invalidateAnalyticsCache()
		}
	default:
		if len(msg.String()) == 1 {
			m.analyticsFilter.text += msg.String()
			m.invalidateAnalyticsCache()
		}
	}
	return m, nil
}

func (m Model) availableScreens() []screenTab {
	if !m.experimentalAnalytics {
		return []screenTab{screenDashboard}
	}
	return []screenTab{screenDashboard, screenAnalytics}
}

func (m Model) nextScreen(step int) screenTab {
	screens := m.availableScreens()
	if len(screens) == 0 {
		return screenDashboard
	}

	idx := 0
	for i, screen := range screens {
		if screen == m.screen {
			idx = i
			break
		}
	}

	next := (idx + step) % len(screens)
	if next < 0 {
		next += len(screens)
	}
	return screens[next]
}

func (m Model) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	ids := m.filteredIDs()
	pageStep := m.listPageStep()
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			m.detailOffset = 0
			m.detailTab = 0
			m.tileOffset = 0
		}
	case "down", "j":
		if m.cursor < len(ids)-1 {
			m.cursor++
			m.detailOffset = 0
			m.detailTab = 0
			m.tileOffset = 0
		}
	case "shift+down", "shift+j":
		if len(ids) > 0 {
			m.cursor = clamp(m.cursor+pageStep, 0, len(ids)-1)
			m.detailOffset = 0
			m.detailTab = 0
			m.tileOffset = 0
		}
	case "shift+up", "shift+k":
		if len(ids) > 0 {
			m.cursor = clamp(m.cursor-pageStep, 0, len(ids)-1)
			m.detailOffset = 0
			m.detailTab = 0
			m.tileOffset = 0
		}
	case "pgdown", "ctrl+d":
		if m.activeDashboardView() != dashboardViewSplit {
			m.cursor = clamp(m.cursor+pageStep, 0, len(ids)-1)
			m.detailOffset = 0
		} else {
			m.detailOffset += m.detailPageStep()
		}
	case "pgup", "ctrl+u":
		if m.activeDashboardView() != dashboardViewSplit {
			m.cursor = clamp(m.cursor-pageStep, 0, len(ids)-1)
			m.detailOffset = 0
		} else {
			m.detailOffset -= m.detailPageStep()
			if m.detailOffset < 0 {
				m.detailOffset = 0
			}
		}
	case "home", "g":
		if m.activeDashboardView() != dashboardViewSplit {
			m.cursor = 0
			m.detailOffset = 0
		} else {
			m.detailOffset = 0
		}
	case "end", "G":
		if m.activeDashboardView() != dashboardViewSplit {
			if len(ids) > 0 {
				m.cursor = len(ids) - 1
			}
			m.detailOffset = 0
		} else {
			m.detailOffset = 9999
		}
	case "ctrl+o":
		if id := m.selectedTileID(ids); id != "" {
			m.expandedModelMixTiles[id] = !m.expandedModelMixTiles[id]
		}
	case "left", "h":
		if m.activeDashboardView() == dashboardViewBento {
			if m.cursor > 0 {
				m.cursor--
				m.tileOffset = 0
				m.detailOffset = 0
			}
		}
	case "right", "l":
		if m.activeDashboardView() == dashboardViewBento {
			if m.cursor < len(ids)-1 {
				m.cursor++
				m.tileOffset = 0
				m.detailOffset = 0
			}
		} else {
			m = m.enterDetailMode()
		}
	case "enter":
		m = m.enterDetailMode()
	case "/":
		m.filter.active = true
		m.filter.text = ""
	case "esc":
		if m.filter.text != "" {
			m.filter.text = ""
			m.cursor = 0
			m.tileOffset = 0
		} else {
			m.tileOffset = 0
		}
	case "r":
		return m.triggerRefreshFocused()
	case "R":
		return m.triggerRefreshAll()
	}
	return m, nil
}

func (m Model) handleDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc", "backspace":
		m = m.exitDetailMode()
	case "shift+tab", "left", "h":
		m = m.navigateDetailSection(-1)
	case "tab", "right", "l":
		m = m.navigateDetailSection(1)
	case "up", "k":
		if m.detailOffset > 0 {
			m.detailOffset--
		}
	case "down", "j":
		m.detailOffset++
	case "g":
		m.detailOffset = 0
	case "G":
		m.detailOffset = 9999
	case "[":
		if m.detailTab > 0 {
			m.detailTab--
			m.detailOffset = 0
		}
	case "]":
		m.detailTab++
		m.detailOffset = 0
	case "pgdown", "ctrl+d":
		m.detailOffset += m.detailPageStep()
	case "pgup", "ctrl+u":
		m.detailOffset -= m.detailPageStep()
		if m.detailOffset < 0 {
			m.detailOffset = 0
		}
	case "r":
		return m.triggerRefreshFocused()
	case "R":
		return m.triggerRefreshAll()
	}
	return m, nil
}

func (m Model) navigateDetailSection(step int) Model {
	starts := m.detailSectionStarts()
	if len(starts) == 0 {
		return m
	}

	current := max(0, m.detailOffset)
	if step > 0 {
		for _, start := range starts {
			if start > current {
				m.detailOffset = start
				return m
			}
		}
		m.detailOffset = starts[len(starts)-1]
		return m
	}

	prev := 0
	for _, start := range starts {
		if start >= current {
			break
		}
		prev = start
	}
	m.detailOffset = prev
	return m
}

func (m Model) detailSectionStarts() []int {
	ids := m.filteredIDs()
	if len(ids) == 0 || m.cursor < 0 || m.cursor >= len(ids) {
		return nil
	}

	snap, ok := m.snapshots[ids[m.cursor]]
	if !ok {
		return nil
	}

	width := m.width - 2
	if width < 20 {
		width = 20
	}
	return CockpitSectionStarts(
		snap,
		m.viewNow(),
		width,
		m.warnThreshold,
		m.critThreshold,
		m.timeWindow,
		m.resolveHideCosts(snap),
		m.usageMode,
	)
}

func (m Model) detailPageStep() int {
	step := m.height / 2
	if step < 3 {
		step = 3
	}
	return step
}

func (m Model) handleFilterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.filter.active = false
		m.cursor = 0
		m.tileOffset = 0
	case "esc":
		m.filter.text = ""
		m.filter.active = false
		m.cursor = 0
		m.tileOffset = 0
	case "backspace":
		if len(m.filter.text) > 0 {
			m.filter.text = m.filter.text[:len(m.filter.text)-1]
		}
	default:
		if len(msg.String()) == 1 {
			m.filter.text += msg.String()
		}
	}
	return m, nil
}

func (m Model) handleTilesKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	ids := m.filteredIDs()
	cols := m.tileCols()
	scrollModeWidget := m.shouldUseWidgetScroll()
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.cursor >= cols {
			m.cursor -= cols
			m.tileOffset = 0
		}
	case "down", "j":
		if m.cursor+cols < len(ids) {
			m.cursor += cols
			m.tileOffset = 0
		}
	case "left", "h":
		if m.cursor > 0 {
			m.cursor--
			m.tileOffset = 0
		}
	case "right", "l":
		if m.cursor < len(ids)-1 {
			m.cursor++
			m.tileOffset = 0
		}
	case "pgdown", "ctrl+d":
		if scrollModeWidget {
			m.tileOffset += m.widgetScrollStep()
		} else {
			m.tileOffset += m.tileScrollStep()
		}
	case "pgup", "ctrl+u":
		if scrollModeWidget {
			m.tileOffset -= m.widgetScrollStep()
		} else {
			m.tileOffset -= m.tileScrollStep()
		}
		if m.tileOffset < 0 {
			m.tileOffset = 0
		}
	case "ctrl+o":
		if id := m.selectedTileID(ids); id != "" {
			m.expandedModelMixTiles[id] = !m.expandedModelMixTiles[id]
		}
	case "home":
		m.tileOffset = 0
	case "end":
		m.tileOffset = 9999
	case "enter":
		m = m.enterDetailMode()
	case "/":
		m.filter.active = true
		m.filter.text = ""
	case "esc":
		if m.filter.text != "" {
			m.filter.text = ""
			m.cursor = 0
			m.tileOffset = 0
		}
	case "r":
		return m.triggerRefreshFocused()
	case "R":
		return m.triggerRefreshAll()
	}
	return m, nil
}
