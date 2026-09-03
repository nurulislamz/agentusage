package tui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nurulislamz/agentusage/internal/boxes"
	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/core"
)

type mockBoxesService struct {
	fakeServices
	createdBox string
	deletedBox string
	loggedIn   string
	boxList    []boxes.AntigravityBox
}

func (s *mockBoxesService) CreateAntigravityBox(name string) error {
	s.createdBox = name
	s.boxList = append(s.boxList, boxes.AntigravityBox{
		Name:      name,
		AccountID: "antigravity-" + name,
		Status:    boxes.StatusInitialized,
	})
	return nil
}

func (s *mockBoxesService) DeleteAntigravityBox(name string) error {
	s.deletedBox = name
	var filtered []boxes.AntigravityBox
	for _, b := range s.boxList {
		if b.Name != name {
			filtered = append(filtered, b)
		}
	}
	s.boxList = filtered
	return nil
}

func (s *mockBoxesService) ListAntigravityBoxes() ([]boxes.AntigravityBox, error) {
	return s.boxList, nil
}

func (s *mockBoxesService) LoginAntigravityBox(ctx context.Context, name string, onURL func(string)) error {
	s.loggedIn = name
	if onURL != nil {
		onURL("https://accounts.google.com/o/oauth2/auth?mock=1")
	}
	return nil
}

func TestSettingsModal_BoxesTab_LifecycleAndActions(t *testing.T) {
	svc := &mockBoxesService{
		boxList: []boxes.AntigravityBox{
			{Name: "box-alpha", AccountID: "antigravity-box-alpha", Status: boxes.StatusReady},
			{Name: "box-beta", AccountID: "antigravity-box-beta", Status: boxes.StatusInitialized},
		},
	}

	m := NewModel(0.2, 0.05, false, config.DashboardConfig{}, nil, core.TimeWindow30d)
	m.SetServices(svc)
	m.width = 120
	m.height = 40

	// 1. Open settings modal
	m.openSettingsModal()
	if !m.settings.show {
		t.Fatalf("expected settings modal to be open")
	}

	// 2. Switch to tab 7 (settingsTabBoxes)
	m.settings.tab = settingsTabBoxes

	// Trigger load command
	loadCmd := m.loadAntigravityBoxesCmd()
	msg := loadCmd()
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	if len(m.settings.boxes.boxes) != 2 {
		t.Fatalf("expected 2 boxes, got %d", len(m.settings.boxes.boxes))
	}

	// 3. Render body verification
	body := m.renderSettingsBoxesBody(100, 20)
	if body == "" {
		t.Fatalf("expected non-empty rendered body")
	}

	// 4. Test Key Navigation (j/k)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updatedModel.(Model)
	if m.settings.boxes.cursor != 1 {
		t.Errorf("cursor after down = %d, want 1", m.settings.boxes.cursor)
	}

	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updatedModel.(Model)
	if m.settings.boxes.cursor != 0 {
		t.Errorf("cursor after up = %d, want 0", m.settings.boxes.cursor)
	}

	// 5. Test Box Creation Input ('a', typing name, 'Enter')
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updatedModel.(Model)
	if !m.settings.boxes.creating {
		t.Fatalf("expected creating mode to be active")
	}

	for _, ch := range "work-box" {
		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		m = updatedModel.(Model)
	}
	if m.settings.boxes.createInput != "work-box" {
		t.Errorf("createInput = %q, want %q", m.settings.boxes.createInput, "work-box")
	}

	// Submit creation with Enter
	updatedModel, createCmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(Model)
	if createCmd == nil {
		t.Fatalf("expected createCmd on Enter")
	}

	createMsg := createCmd()
	updatedModel, postCreateCmd := m.Update(createMsg)
	m = updatedModel.(Model)

	if svc.createdBox != "work-box" {
		t.Errorf("svc.createdBox = %q, want %q", svc.createdBox, "work-box")
	}

	// Execute post create reload cmd
	if postCreateCmd != nil {
		reloadMsg := postCreateCmd()
		updatedModel, _ = m.Update(reloadMsg)
		m = updatedModel.(Model)
	}

	// 6. Test Login Action ('l' / Enter)
	m.settings.boxes.cursor = 0
	updatedModel, loginCmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m = updatedModel.(Model)
	if loginCmd == nil {
		t.Fatalf("expected loginCmd on 'l'")
	}

	loginMsg := loginCmd()
	updatedModel, _ = m.Update(loginMsg)
	m = updatedModel.(Model)

	if svc.loggedIn != "box-alpha" {
		t.Errorf("svc.loggedIn = %q, want %q", svc.loggedIn, "box-alpha")
	}

	// 7. Test Delete Action ('d')
	updatedModel, deleteCmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = updatedModel.(Model)
	if deleteCmd == nil {
		t.Fatalf("expected deleteCmd on 'd'")
	}

	deleteMsg := deleteCmd()
	updatedModel, _ = m.Update(deleteMsg)
	m = updatedModel.(Model)

	if svc.deletedBox != "box-alpha" {
		t.Errorf("svc.deletedBox = %q, want %q", svc.deletedBox, "box-alpha")
	}
}

func TestSettingsModal_BoxesTab_CreateCancelAndEmptyValidation(t *testing.T) {
	svc := &mockBoxesService{}
	m := NewModel(0.2, 0.05, false, config.DashboardConfig{}, nil, core.TimeWindow30d)
	m.SetServices(svc)
	m.openSettingsModal()
	m.settings.tab = settingsTabBoxes

	// 1. Enter create mode with 'a'
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updatedModel.(Model)
	if !m.settings.boxes.creating {
		t.Fatalf("expected creating to be true")
	}

	// 2. Press Esc to cancel
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(Model)
	if m.settings.boxes.creating {
		t.Errorf("expected creating to be false after Esc")
	}

	// 3. Press 'a', then Enter with empty input
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updatedModel.(Model)
	updatedModel, emptyCmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(Model)
	if emptyCmd != nil {
		t.Errorf("expected no command when submitting empty box name")
	}
	if m.settings.boxes.status == "" {
		t.Errorf("expected error status when submitting empty box name")
	}
}

