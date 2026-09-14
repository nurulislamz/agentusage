package antigravity

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
)

func TestProbeLiveBoxes(t *testing.T) {
	if testing.Short() {
		t.Skip("live probe")
	}
	p := New()
	for _, box := range []string{"chaos", "physics", "mohammed", "nurulz"} {
		acct := core.AccountConfig{ID: "antigravity-" + box}
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		snap, err := p.Fetch(ctx, acct)
		cancel()
		diag, _ := json.Marshal(snap.Diagnostics)
		raw, _ := json.Marshal(snap.Raw)
		t.Logf("box=%s status=%s msg=%q err=%v diag=%s raw=%s", box, snap.Status, snap.Message, err, diag, raw)
	}
}
