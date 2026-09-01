package main

import (
	"context"
	"fmt"
	"time"

	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/daemon"
	"github.com/nurulislamz/agentusage/internal/providers"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("config error:", err)
		return
	}
	fmt.Printf("Loaded config: %d accounts, %d auto-detected\n", len(cfg.Accounts), len(cfg.AutoDetectedAccounts))

	accts := daemon.ResolveAccounts(&cfg)
	fmt.Printf("Resolved accounts: %d\n", len(accts))
	for _, a := range accts {
		fmt.Printf("  - %s (%s)\n", a.ID, a.Provider)
	}

	all := providers.AllProviders()
	pMap := make(map[string]core.UsageProvider)
	for _, p := range all {
		pMap[p.ID()] = p
	}

	for _, a := range accts {
		p, ok := pMap[a.Provider]
		if !ok {
			fmt.Printf("Provider not found: %s\n", a.Provider)
			continue
		}
		fmt.Printf("Fetching %s (%s)... ", a.ID, a.Provider)
		start := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		snap, err := p.Fetch(ctx, a)
		cancel()
		fmt.Printf("done in %v (err=%v, status=%s, msg=%s)\n", time.Since(start), err, snap.Status, snap.Message)
	}
}
