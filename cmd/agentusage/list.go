package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/nurulislamz/agentusage/internal/boxes"
	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/daemon"
)

type AccountListItem struct {
	ID       string `json:"id"`
	Provider string `json:"provider"`
	Auth     string `json:"auth"`
	Status   string `json:"status"`
	BoxName  string `json:"box_name,omitempty"`
	Source   string `json:"source,omitempty"`
}

type listOptions struct {
	format  string
	json    bool
	idsOnly bool
}

func newListCommand() *cobra.Command {
	var opts listOptions

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List available provider accounts and box IDs",
		Long: `List all detected and configured accounts and boxes across all providers.

Returns the ID for each account, which can be passed to 'agentusage get <id>' (or 'agu get <id>')
to query usage and rate limits.

By default, prints a table in a terminal and JSON when piped. Use --json or -q for scripting.`,
		Example: strings.Join([]string{
			"  agentusage list",
			"  agentusage list --json",
			"  agu list -q",
		}, "\n"),
		RunE: func(_ *cobra.Command, _ []string) error {
			return runList(os.Stdout, opts)
		},
	}

	fl := cmd.Flags()
	fl.StringVar(&opts.format, "format", "", "output format: table, json, or ids (default: auto)")
	fl.BoolVar(&opts.json, "json", false, "output as JSON (alias for --format=json)")
	fl.BoolVarP(&opts.idsOnly, "quiet", "q", false, "print only account IDs, one per line")

	return cmd
}

func runList(out io.Writer, opts listOptions) error {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	accounts := daemon.ResolveAccounts(&cfg)
	items := buildListItems(accounts)

	format := strings.ToLower(strings.TrimSpace(opts.format))
	if opts.json {
		format = "json"
	} else if opts.idsOnly {
		format = "ids"
	}

	if format == "" {
		if isTerminalWriter(out) {
			format = "table"
		} else {
			format = "json"
		}
	}

	switch format {
	case "json":
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(items)
	case "ids", "quiet":
		for _, it := range items {
			fmt.Fprintln(out, it.ID)
		}
		return nil
	case "table":
		w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tPROVIDER\tAUTH\tSTATUS")
		for _, it := range items {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", it.ID, it.Provider, it.Auth, it.Status)
		}
		return w.Flush()
	default:
		return fmt.Errorf("unknown format %q (supported: table, json, ids)", format)
	}
}

func buildListItems(accounts []core.AccountConfig) []AccountListItem {
	items := make([]AccountListItem, 0, len(accounts))

	for _, acct := range accounts {
		auth := acct.Auth
		if auth == "" {
			auth = "local"
		}
		status := resolveAccountStatus(acct)
		boxName := acct.Hint("box_name", "")
		source := acct.Hint("credential_source", "")

		items = append(items, AccountListItem{
			ID:       acct.ID,
			Provider: acct.Provider,
			Auth:     auth,
			Status:   status,
			BoxName:  boxName,
			Source:   source,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Provider != items[j].Provider {
			return items[i].Provider < items[j].Provider
		}
		return items[i].ID < items[j].ID
	})

	return items
}

func resolveAccountStatus(acct core.AccountConfig) string {
	if acct.Provider == "antigravity" {
		box := acct.Hint("box_name", "")
		if box != "" {
			b := boxes.InspectBox(boxes.DefaultContainersDir(), box)
			if b.Status != "" {
				return string(b.Status)
			}
		}
		dir := acct.Path("config_dir", "")
		if dir != "" {
			tokenFile := acct.Path("oauth_token_file", "")
			if tokenFile == "" {
				tokenFile = dir + "/antigravity-oauth-token"
			}
			if fileExists(tokenFile) {
				return "Ready"
			}
		}
	}

	if acct.Token != "" || acct.APIKeyEnv != "" {
		return "Ready"
	}
	if acct.Hint("credential_source", "") != "" || acct.Hint("auth_file", "") != "" || acct.Binary != "" {
		return "Ready"
	}
	return "Configured"
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func isTerminalWriter(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}
