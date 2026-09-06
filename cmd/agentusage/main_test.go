package main

import (
	"bytes"
	"slices"
	"strings"
	"testing"
)

var removedCommands = []string{
	"blocks",
	"completion",
	"cursor",
	"export",
	"hub",
	"hub-view",
	"integrations",
	"monthly",
	"daily",
	"pricing",
	"session",
	"statusline",
	"telemetry",
	"tmux",
	"version",
	"weekly",
}

var retainedCommands = []string{
	"daemon",
	"detect",
	"doctor",
	"get",
	"help",
	"list",
	"serve",
}

var visibleEverydayCommands = []string{
	"doctor",
	"get",
	"list",
	"serve",
}

var visibleAdminCommands = []string{
	"daemon",
}

var compatibilityCommands = []string{
	"detect",
}

func TestRootCommands_RemovedCommandsNotPresent(t *testing.T) {
	root := newRootCommand()
	commands := root.Commands()

	var commandNames []string
	for _, c := range commands {
		commandNames = append(commandNames, c.Name())
	}

	for _, removed := range removedCommands {
		if slices.Contains(commandNames, removed) {
			t.Errorf("expected command %q to be removed, but it was present in Commands()", removed)
		}
	}
}

func TestRootCommands_RetainedCommandsPresent(t *testing.T) {
	root := newRootCommand()
	commands := root.Commands()

	var commandNames []string
	for _, c := range commands {
		commandNames = append(commandNames, c.Name())
	}

	for _, retained := range retainedCommands {
		if !slices.Contains(commandNames, retained) {
			t.Errorf("expected command %q to be present, but it was missing from Commands()", retained)
		}
	}

	if len(commandNames) != len(retainedCommands) {
		t.Errorf("expected exactly %d commands (%v), got %d (%v)",
			len(retainedCommands), retainedCommands, len(commandNames), commandNames)
	}
}

func TestRootCommands_ExecuteRemovedCommandsReturnError(t *testing.T) {
	for _, removed := range removedCommands {
		t.Run(removed, func(t *testing.T) {
			root := newRootCommand()
			var out bytes.Buffer
			root.SetOut(&out)
			root.SetErr(&out)
			root.SetArgs([]string{removed})

			err := root.Execute()
			if err == nil {
				t.Fatalf("expected error executing removed command %q, but got nil", removed)
			}
			if !strings.Contains(err.Error(), "unknown command") {
				t.Errorf("expected error message to contain 'unknown command', got: %v", err)
			}
		})
	}
}

func TestRootCommands_HelpOutputGroupsAndCompatibility(t *testing.T) {
	root := newRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error running --help: %v", err)
	}

	helpOutput := out.String()

	// Verify command group headings exist in help output
	if !strings.Contains(helpOutput, "Everyday Commands:") {
		t.Errorf("expected 'Everyday Commands:' group in help output, got:\n%s", helpOutput)
	}
	if !strings.Contains(helpOutput, "Advanced Service Administration:") {
		t.Errorf("expected 'Advanced Service Administration:' group in help output, got:\n%s", helpOutput)
	}

	// Verify removed commands are not anywhere in help output
	for _, removed := range removedCommands {
		if strings.Contains(helpOutput, "\n  "+removed+" ") {
			t.Errorf("removed command %q found in help output", removed)
		}
	}

	// Verify visible everyday commands are listed
	for _, everyday := range visibleEverydayCommands {
		if !strings.Contains(helpOutput, everyday) {
			t.Errorf("visible everyday command %q not found in help output", everyday)
		}
	}

	// Verify visible admin commands are listed
	for _, admin := range visibleAdminCommands {
		if !strings.Contains(helpOutput, admin) {
			t.Errorf("visible admin command %q not found in help output", admin)
		}
	}

	// Verify compatibility command (detect) is hidden from help output
	for _, comp := range compatibilityCommands {
		for _, line := range strings.Split(helpOutput, "\n") {
			fields := strings.Fields(line)
			if len(fields) > 0 && fields[0] == comp {
				t.Errorf("compatibility command %q should be hidden, but found in help output: %q", comp, line)
			}
		}
	}
}

func TestRootCommands_CompatibilityRoutesExecute(t *testing.T) {
	root := newRootCommand()
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"detect"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error executing compatibility route 'detect': %v", err)
	}

	// Assert deprecation diagnostic is sent to stderr
	errOutput := stderr.String()
	if !strings.Contains(errOutput, "deprecated") || !strings.Contains(errOutput, "doctor --detect") {
		t.Errorf("expected deprecation notice in stderr, got:\n%s", errOutput)
	}

	// Assert report output was produced on stdout
	outText := stdout.String()
	if !strings.Contains(outText, "Tools detected:") || !strings.Contains(outText, "Accounts detected:") {
		t.Errorf("expected detection report on stdout, got:\n%s", outText)
	}
}
