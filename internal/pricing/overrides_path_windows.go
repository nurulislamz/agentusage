//go:build windows

package pricing

import (
	"path/filepath"

	"github.com/nurulislamz/agentusage/internal/config"
)

// platformCustomOverridesPath places custom-pricing.json beside settings.json
// under %APPDATA%\agentusage on Windows (config.ConfigDir()).
func platformCustomOverridesPath() (string, error) {
	return filepath.Join(config.ConfigDir(), CustomOverridesFilename), nil
}
