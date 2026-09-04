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

func TestRootCommands_HelpOutputDoesNotContainRemovedCommands(t *testing.T) {
	root := newRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error running --help: %v", err)
	}

	helpOutput := out.String()

	// Available Commands section
	availIdx := strings.Index(helpOutput, "Available Commands:")
	if availIdx == -1 {
		t.Fatalf("expected 'Available Commands:' section in help output, got:\n%s", helpOutput)
	}
	availSection := helpOutput[availIdx:]
	if flagsIdx := strings.Index(availSection, "Flags:"); flagsIdx != -1 {
		availSection = availSection[:flagsIdx]
	}

	for _, removed := range removedCommands {
		// Check that the command name is not listed as an available command
		for _, line := range strings.Split(availSection, "\n") {
			fields := strings.Fields(line)
			if len(fields) > 0 && fields[0] == removed {
				t.Errorf("removed command %q found in Available Commands section: %q", removed, line)
			}
		}
	}

	for _, retained := range retainedCommands {
		found := false
		for _, line := range strings.Split(availSection, "\n") {
			fields := strings.Fields(line)
			if len(fields) > 0 && fields[0] == retained {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("retained command %q not found in Available Commands section", retained)
		}
	}
}
