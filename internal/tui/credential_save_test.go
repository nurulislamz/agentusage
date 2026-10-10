package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/core"
)

func TestHandleValidateKeyResultMsg_SavesValidatedKeyNotLiveBuffer(t *testing.T) {
	svc := &addAccountTestServices{validateResult: true}
	m := newTestModelForAddAccount()
	m.SetServices(svc)
	m.settings.apiKeyEditing = true
	m.settings.apiKeyEditAccountID = "openai"
	m.settings.apiKeyInput = "sk-live-buffer-for-other-account"

	next, cmd := m.handleValidateKeyResultMsg(validateKeyResultMsg{
		AccountID: "openai",
		APIKey:    "sk-validated-original",
		Valid:     true,
	})
	if cmd == nil {
		t.Fatal("expected save command after valid key")
	}
	model := next.(Model)
	if model.settings.apiKeyStatus != "valid ✓ — saving..." {
		t.Fatalf("status = %q", model.settings.apiKeyStatus)
	}

	msgs := unwrapCmd(cmd)
	if len(msgs) != 1 {
		t.Fatalf("save cmd produced %d msgs, want 1", len(msgs))
	}
	saved, ok := msgs[0].(credentialSavedMsg)
	if !ok {
		t.Fatalf("got %T, want credentialSavedMsg", msgs[0])
	}
	if saved.Err != nil {
		t.Fatalf("save err: %v", saved.Err)
	}
	if svc.savedCredAcct != "openai" {
		t.Fatalf("saved account = %q, want openai", svc.savedCredAcct)
	}
	if svc.savedCredKey != "sk-validated-original" {
		t.Fatalf("saved key = %q, want the validated key, not the live buffer", svc.savedCredKey)
	}
}

func TestHandleAPIKeyEditKey_FreezesInputWhileValidating(t *testing.T) {
	m := newTestModelForAddAccount()
	m.settings.apiKeyEditing = true
	m.settings.apiKeyEditAccountID = "openai"
	m.settings.apiKeyInput = "sk-original"
	m.settings.apiKeyStatus = "validating..."

	next, cmd := m.handleAPIKeyEditKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	if cmd != nil {
		t.Fatal("expected no command while validating")
	}
	model := next.(Model)
	if model.settings.apiKeyInput != "sk-original" {
		t.Fatalf("input mutated during validating: %q", model.settings.apiKeyInput)
	}

	next, _ = model.handleAPIKeyEditKey(tea.KeyMsg{Type: tea.KeyBackspace})
	model = next.(Model)
	if model.settings.apiKeyInput != "sk-original" {
		t.Fatalf("backspace mutated input during validating: %q", model.settings.apiKeyInput)
	}
}

func TestDashboardConfigProviders_PreservesHideCosts(t *testing.T) {
	hide := true
	m := NewModel(
		80, 95, false,
		config.DashboardConfig{
			Providers: []config.DashboardProviderConfig{
				{AccountID: "openai", Enabled: true, HideCosts: &hide},
				{AccountID: "anthropic", Enabled: false},
			},
		},
		[]core.AccountConfig{
			{ID: "openai", Provider: "openai"},
			{ID: "anthropic", Provider: "anthropic"},
		},
		core.TimeWindow(""),
	)

	got := m.dashboardConfigProviders()
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}

	byID := map[string]config.DashboardProviderConfig{}
	for _, p := range got {
		byID[p.AccountID] = p
	}
	openai := byID["openai"]
	if openai.HideCosts == nil || !*openai.HideCosts {
		t.Fatalf("openai HideCosts = %v, want true", openai.HideCosts)
	}
	if !openai.Enabled {
		t.Fatal("openai should stay enabled")
	}
	anthropic := byID["anthropic"]
	if anthropic.HideCosts != nil {
		t.Fatalf("anthropic HideCosts = %v, want nil", anthropic.HideCosts)
	}
	if anthropic.Enabled {
		t.Fatal("anthropic should stay disabled")
	}
}
