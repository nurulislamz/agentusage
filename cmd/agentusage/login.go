package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/providers/claude_code"
)

// newLoginCommand returns `agentusage login <provider>`. Only Claude is
// supported: it gives agentusage its own renewable OAuth sign-in for the
// 5h/7d quota gauges, so they keep working when the Claude Code CLI's token
// lapses between sessions.
func newLoginCommand() *cobra.Command {
	return &cobra.Command{
		Use:       "login claude",
		Short:     "Sign agentusage in to Claude for live 5h/7d quota",
		Args:      cobra.ExactValidArgs(1),
		ValidArgs: []string{"claude"},
		Long: `Signs agentusage in to your Claude account with its own OAuth token
(scope user:profile, read-only usage access). The token is stored in
agentusage's credentials.json and renewed automatically, independently of
the Claude Code CLI's own login.

Open the printed link in any browser, approve, and paste the code shown
back here. Works on headless machines.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runClaudeLogin(cmd.Context(), cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}
}

func newLogoutCommand() *cobra.Command {
	return &cobra.Command{
		Use:       "logout claude",
		Short:     "Remove agentusage's Claude sign-in",
		Args:      cobra.ExactValidArgs(1),
		ValidArgs: []string{"claude"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := config.DeleteOAuthToken(claude_code.AgentOAuthAccountID); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Removed agentusage's Claude sign-in. (Claude Code's own login is unaffected.)")
			return nil
		},
	}
}

func runClaudeLogin(ctx context.Context, in io.Reader, out io.Writer) error {
	if ctx == nil {
		ctx = context.Background()
	}
	login, err := claude_code.NewClaudeLogin()
	if err != nil {
		return err
	}

	fmt.Fprintln(out, "Open this link in a browser and approve access:")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "  "+login.AuthorizeURL)
	fmt.Fprintln(out)
	fmt.Fprint(out, "Paste the code shown after approving: ")

	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && line == "" {
		return fmt.Errorf("reading code: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	tok, err := login.Complete(ctx, line)
	if err != nil {
		return err
	}

	fmt.Fprintln(out)
	fmt.Fprintf(out, "Signed in. Token saved to %s", config.CredentialsPath())
	if tok.ExpiresAt != "" {
		fmt.Fprintf(out, " (renews automatically; current token valid until %s)", tok.ExpiresAt)
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "The daemon picks it up on its next poll; no restart needed.")
	return nil
}
