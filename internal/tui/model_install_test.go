package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/core"
)

func TestDaemonInstallResultSuccess(t *testing.T) {
	m := NewModel(0.2, 0.1, false, config.DashboardConfig{}, nil, core.TimeWindow30d)
	m.daemon.status = DaemonNotInstalled
	m.daemon.installing = true

	updated, _ := m.Update(daemonInstallResultMsg{err: nil})
	got, ok := updated.(Model)
	if !ok {
		t.Fatalf("updated model type = %T, want tui.Model", updated)
	}

	if got.daemon.installing {
		t.Fatal("expected daemonInstalling=false after successful install")
	}
	if got.daemon.status != DaemonStarting {
		t.Fatalf("daemonStatus = %q, want %q", got.daemon.status, DaemonStarting)
	}
	if !got.daemon.installDone {
		t.Fatal("expected daemonInstallDone=true after successful install")
	}
}

func TestDaemonInstallResultFailure(t *testing.T) {
	m := NewModel(0.2, 0.1, false, config.DashboardConfig{}, nil, core.TimeWindow30d)
	m.daemon.status = DaemonNotInstalled
	m.daemon.installing = true

	installErr := errors.New("failed to install daemon")
	updated, _ := m.Update(daemonInstallResultMsg{err: installErr})
	got, ok := updated.(Model)
	if !ok {
		t.Fatalf("updated model type = %T, want tui.Model", updated)
	}

	if got.daemon.installing {
		t.Fatal("expected daemonInstalling=false after failed install")
	}
	if got.daemon.status != DaemonError {
		t.Fatalf("daemonStatus = %q, want %q", got.daemon.status, DaemonError)
	}
	if got.daemon.message != "failed to install daemon" {
		t.Fatalf("daemonMessage = %q, want %q", got.daemon.message, "failed to install daemon")
	}
}

func TestSplashKeyAndMouse_EnterTriggersSetupAndRetry(t *testing.T) {
	t.Run("Enter key when DaemonNotInstalled triggers install", func(t *testing.T) {
		m := NewModel(0.2, 0.1, false, config.DashboardConfig{}, nil, core.TimeWindow30d)
		m.hasData = false
		m.daemon.status = DaemonNotInstalled
		m.onInstallDaemon = func() error { return nil }

		updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		got := updated.(Model)
		if !got.daemon.installing {
			t.Fatal("expected daemon.installing to be true after Enter on DaemonNotInstalled")
		}
		if cmd == nil {
			t.Fatal("expected non-nil cmd from installDaemonCmd")
		}
	})

	t.Run("Mouse click when DaemonNotInstalled triggers install", func(t *testing.T) {
		m := NewModel(0.2, 0.1, false, config.DashboardConfig{}, nil, core.TimeWindow30d)
		m.hasData = false
		m.daemon.status = DaemonNotInstalled
		m.onInstallDaemon = func() error { return nil }

		updated, cmd := m.Update(tea.MouseMsg{Button: tea.MouseButtonLeft, Action: tea.MouseActionPress, X: 10, Y: 10})
		got := updated.(Model)
		if !got.daemon.installing {
			t.Fatal("expected daemon.installing to be true after Left Click on DaemonNotInstalled")
		}
		if cmd == nil {
			t.Fatal("expected non-nil cmd from Left Click install")
		}
	})

	t.Run("Enter key when DaemonError triggers retry install", func(t *testing.T) {
		m := NewModel(0.2, 0.1, false, config.DashboardConfig{}, nil, core.TimeWindow30d)
		m.hasData = false
		m.daemon.status = DaemonError
		m.daemon.message = "could not connect to daemon"
		m.onInstallDaemon = func() error { return nil }

		updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		got := updated.(Model)
		if !got.daemon.installing {
			t.Fatal("expected daemon.installing to be true after Enter on DaemonError")
		}
		if cmd == nil {
			t.Fatal("expected non-nil cmd from Enter retry")
		}
	})

	t.Run("Mouse click when DaemonError triggers retry install", func(t *testing.T) {
		m := NewModel(0.2, 0.1, false, config.DashboardConfig{}, nil, core.TimeWindow30d)
		m.hasData = false
		m.daemon.status = DaemonError
		m.daemon.message = "could not connect to daemon"
		m.onInstallDaemon = func() error { return nil }

		updated, cmd := m.Update(tea.MouseMsg{Button: tea.MouseButtonLeft, Action: tea.MouseActionPress, X: 10, Y: 10})
		got := updated.(Model)
		if !got.daemon.installing {
			t.Fatal("expected daemon.installing to be true after Left Click on DaemonError")
		}
		if cmd == nil {
			t.Fatal("expected non-nil cmd from Left Click retry")
		}
	})

	t.Run("Splash progress on DaemonError shows retry prompt", func(t *testing.T) {
		m := Model{
			daemon:        daemonState{status: DaemonError, message: "service failed to start"},
			providerOrder: []string{"openai"},
			animFrame:     0,
		}
		lines := m.splashProgressLines()
		combined := strings.Join(lines, "\n")
		if !strings.Contains(combined, "Press Enter to retry") {
			t.Fatalf("expected 'Press Enter to retry' in splash lines on DaemonError, got:\n%s", combined)
		}
	})
}

