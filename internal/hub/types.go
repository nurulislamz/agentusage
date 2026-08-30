package hub

import (
	"time"

	"github.com/nurulislamz/openusage/internal/core"
)

type machineEntry struct {
	envelope   core.RemoteEnvelope
	receivedAt time.Time
}

type pushResponse struct {
	OK bool `json:"ok"`
}

type healthResponse struct {
	Status   string   `json:"status"`
	Machines []string `json:"machines"`
}
