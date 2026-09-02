package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/core"
)

type addAccountTestServices struct {
	fakeServices
	savedAccount    core.AccountConfig
	savedCredAcct   string
	savedCredKey    string
	validateResult  bool
	validateErrMsg  string
	validateCalls   int
	browsersList    []string
}

func (s *addAccountTestServices) SaveAccount(acct core.AccountConfig) error {
	s.savedAccount = acct
	return nil
}

func (s *addAccountTestServices) SaveCredential(accountID, apiKey string) error {
	s.savedCredAcct = accountID
	s.savedCredKey = apiKey
	return nil
}

func (s *addAccountTestServices) ValidateAPIKey(accountID, providerID, apiKey string) (bool, string) {
	s.validateCalls++
	return s.validateResult, s.validateErrMsg
}

func (s *addAccountTestServices) AvailableBrowsers() ([]string, error) {
	return s.browsersList, nil
}

func unwrapCmd(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		var msgs []tea.Msg
		for _, c := range batch {
			if c != nil {
				msgs = append(msgs, unwrapCmd(c)...)
			}
		}
		return msgs
	}
	return []tea.Msg{msg}
}

func newTestModelForAddAccount() Model {
	m := NewModel(
		80, 95, false,
		config.DashboardConfig{},
		[]core.AccountConfig{
			{ID: "openai", Provider: "openai"},
			{ID: "anthropic", Provider: "anthropic"},
		},
		core.TimeWindow(""),
	)
	m.width = 100
	m.height = 30
	snap1 := core.NewUsageSnapshot("openai", "openai")
	snap1.Status = core.StatusOK
	snap2 := core.NewUsageSnapshot("anthropic", "anthropic")
	snap2.Status = core.StatusOK
	m.snapshots = map[string]core.UsageSnapshot{
		"openai":    snap1,
		"anthropic": snap2,
	}
	m.hasData = true
	return m
}

func TestAddAccount_OpenAndCloseModal(t *testing.T) {
	m := newTestModelForAddAccount()

	// Press 'a' to open Add Account modal
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = res.(Model)

	if !m.settings.addAccount.active {
		t.Fatalf("expected addAccount modal to be active")
	}
	if len(m.settings.addAccount.providerList) == 0 {
		t.Fatalf("expected providerList to be populated")
	}

	// Verify overlay renders without panic
	view := m.View()
	if !strings.Contains(view, "Add Provider Account") {
		t.Fatalf("expected view to contain 'Add Provider Account', got: %s", view)
	}

	// Press Esc to close
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = res.(Model)

	if m.settings.addAccount.active {
		t.Fatalf("expected addAccount modal to be inactive after Esc")
	}
}

func TestAddAccount_DefaultAccountIDGeneration(t *testing.T) {
	m := newTestModelForAddAccount()
	m.openAddAccountModal()

	providers := m.settings.addAccount.providerList
	if len(providers) == 0 {
		t.Fatalf("empty provider list")
	}

	// For openai, since 'openai' already exists in m.accountProviders,
	// default account ID should be 'openai-2'
	for i, p := range providers {
		if p.ID == "openai" {
			m.settings.addAccount.providerCursor = i
			m.settings.addAccount.accountIDEdited = false
			m.syncAddAccountProviderState()
			if m.settings.addAccount.accountID != "openai-2" {
				t.Fatalf("expected default account ID 'openai-2', got '%s'", m.settings.addAccount.accountID)
			}
			break
		}
	}
}

func TestAddAccount_FieldNavigationAndInput(t *testing.T) {
	m := newTestModelForAddAccount()
	m.openAddAccountModal()

	// Initial field: Provider
	if m.settings.addAccount.field != addAccountFieldProvider {
		t.Fatalf("expected initial field provider, got %d", m.settings.addAccount.field)
	}

	// Press Enter to move to Account ID field
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = res.(Model)
	if m.settings.addAccount.field != addAccountFieldAccountID {
		t.Fatalf("expected field Account ID, got %d", m.settings.addAccount.field)
	}

	// Type custom account ID
	m.settings.addAccount.accountID = ""
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m', 'y', '-', 'k', 'e', 'y'}})
	m = res.(Model)
	if m.settings.addAccount.accountID != "my-key" {
		t.Fatalf("expected accountID 'my-key', got '%s'", m.settings.addAccount.accountID)
	}

	// Backspace
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = res.(Model)
	if m.settings.addAccount.accountID != "my-ke" {
		t.Fatalf("expected accountID 'my-ke', got '%s'", m.settings.addAccount.accountID)
	}

	// Tab navigation
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = res.(Model)
	// Should advance to AuthType or Credential
	if m.settings.addAccount.field == addAccountFieldAccountID {
		t.Fatalf("expected field to advance on Tab")
	}
}

func TestAddAccount_DirectAPIKeyValidationAndSave(t *testing.T) {
	m := newTestModelForAddAccount()
	svc := &addAccountTestServices{
		validateResult: true,
	}
	m.SetServices(svc)

	addedCallback := false
	m.SetOnAddAccount(func(acct core.AccountConfig) {
		addedCallback = true
	})

	m.openAddAccountModal()
	// Set specific provider
	for i, p := range m.settings.addAccount.providerList {
		if p.ID == "openai" {
			m.settings.addAccount.providerCursor = i
			m.settings.addAccount.accountID = "openai-work"
			m.settings.addAccount.accountIDEdited = true
			m.syncAddAccountProviderState()
			m.settings.addAccount.accountID = "openai-work"
			break
		}
	}

	// Set direct API key
	m.settings.addAccount.field = addAccountFieldCredential
	m.settings.addAccount.authMode = addAccountAuthDirectKey
	m.settings.addAccount.apiKey = "sk-test-12345"

	// Submit form
	res, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = res.(Model)

	if !m.settings.addAccount.validating {
		t.Fatalf("expected validating state to be true")
	}
	if cmd == nil {
		t.Fatalf("expected validation cmd")
	}

	// Execute cmd
	msgs := unwrapCmd(cmd)
	var valMsg validateNewAccountKeyResultMsg
	found := false
	for _, msg := range msgs {
		if v, ok := msg.(validateNewAccountKeyResultMsg); ok {
			valMsg = v
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected validateNewAccountKeyResultMsg in cmd msgs, got %v", msgs)
	}
	if !valMsg.Valid {
		t.Fatalf("expected valid API key")
	}

	// Feed result message back into Update
	res, nextCmd := m.Update(valMsg)
	m = res.(Model)

	if m.settings.addAccount.active {
		t.Fatalf("expected addAccount modal to be closed after successful save")
	}
	if !addedCallback {
		t.Fatalf("expected onAddAccount callback to have been called")
	}
	if m.accountProviders["openai-work"] != "openai" {
		t.Fatalf("expected accountProviders['openai-work'] == 'openai'")
	}
	if !m.providerEnabled["openai-work"] {
		t.Fatalf("expected providerEnabled['openai-work'] == true")
	}

	// Verify nextCmd executes credential & account save
	if nextCmd != nil {
		_ = unwrapCmd(nextCmd)
	}
	if svc.savedCredAcct != "openai-work" || svc.savedCredKey != "sk-test-12345" {
		t.Fatalf("expected credential saved for openai-work")
	}
}

func TestAddAccount_DirectAPIKeyValidationFailure(t *testing.T) {
	m := newTestModelForAddAccount()
	svc := &addAccountTestServices{
		validateResult: false,
		validateErrMsg: "401 Unauthorized",
	}
	m.SetServices(svc)

	m.openAddAccountModal()
	m.settings.addAccount.accountID = "openai-fail"
	m.settings.addAccount.field = addAccountFieldCredential
	m.settings.addAccount.authMode = addAccountAuthDirectKey
	m.settings.addAccount.apiKey = "sk-bad-key"

	// Submit form
	res, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = res.(Model)

	if cmd == nil {
		t.Fatalf("expected validation cmd")
	}
	msgs := unwrapCmd(cmd)
	var valMsg validateNewAccountKeyResultMsg
	for _, msg := range msgs {
		if v, ok := msg.(validateNewAccountKeyResultMsg); ok {
			valMsg = v
			break
		}
	}

	// Update with invalid result
	res, _ = m.Update(valMsg)
	m = res.(Model)

	if !m.settings.addAccount.active {
		t.Fatalf("expected modal to remain open on validation failure")
	}
	if !strings.Contains(m.settings.addAccount.status, "invalid") {
		t.Fatalf("expected status to show invalid error, got '%s'", m.settings.addAccount.status)
	}
}

func TestAddAccount_EnvVarSave(t *testing.T) {
	m := newTestModelForAddAccount()
	svc := &addAccountTestServices{}
	m.SetServices(svc)

	m.openAddAccountModal()
	m.settings.addAccount.accountID = "anthropic-env"
	m.settings.addAccount.field = addAccountFieldCredential
	m.settings.addAccount.authMode = addAccountAuthEnvVar
	m.settings.addAccount.apiKeyEnv = "ANTHROPIC_API_KEY"

	// Submit form
	res, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = res.(Model)

	if cmd == nil {
		t.Fatalf("expected save cmd")
	}
	msgs := unwrapCmd(cmd)
	var savedMsg accountSavedMsg
	found := false
	for _, msg := range msgs {
		if s, ok := msg.(accountSavedMsg); ok {
			savedMsg = s
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected accountSavedMsg in cmd msgs, got %v", msgs)
	}

	// Feed back into Update
	res, _ = m.Update(savedMsg)
	m = res.(Model)

	if m.settings.addAccount.active {
		t.Fatalf("expected modal to close on success")
	}
	if svc.savedAccount.ID != "anthropic-env" {
		t.Fatalf("expected savedAccount.ID == 'anthropic-env', got '%s'", svc.savedAccount.ID)
	}
	if svc.savedAccount.APIKeyEnv != "ANTHROPIC_API_KEY" {
		t.Fatalf("expected APIKeyEnv == 'ANTHROPIC_API_KEY', got '%s'", svc.savedAccount.APIKeyEnv)
	}
}

func TestAddAccount_EmptyAccountIDValidation(t *testing.T) {
	m := newTestModelForAddAccount()
	m.openAddAccountModal()
	m.settings.addAccount.accountID = "   "
	m.settings.addAccount.field = addAccountFieldSubmit

	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = res.(Model)

	if !strings.Contains(m.settings.addAccount.status, "cannot be empty") {
		t.Fatalf("expected status error for empty ID, got '%s'", m.settings.addAccount.status)
	}
	if m.settings.addAccount.field != addAccountFieldAccountID {
		t.Fatalf("expected focus to jump to Account ID field")
	}
}

func TestAddAccount_FromSettingsModalTabs(t *testing.T) {
	m := newTestModelForAddAccount()

	// Open settings modal on Providers tab
	m.openSettingsModal()
	m.settings.tab = settingsTabProviders

	// Press 'a'
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = res.(Model)
	if !m.settings.addAccount.active {
		t.Fatalf("expected addAccount modal active from Providers tab")
	}

	// Close addAccount modal
	m.closeAddAccountModal()

	// Switch to Keys tab
	m.openSettingsModal()
	m.settings.tab = settingsTabAPIKeys
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = res.(Model)
	if !m.settings.addAccount.active {
		t.Fatalf("expected addAccount modal active from Keys tab")
	}
}
