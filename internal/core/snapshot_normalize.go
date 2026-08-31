package core

import "strings"

func NormalizeUsageSnapshotWithConfig(s UsageSnapshot, modelCfg ModelNormalizationConfig) UsageSnapshot {
	s.EnsureMaps()

	for k, v := range s.Raw {
		if k == "" || v == "" {
			continue
		}
		if isDiagnosticKey(k) {
			if _, ok := s.Diagnostics[k]; !ok {
				s.Diagnostics[k] = v
			}
			continue
		}
		if _, ok := s.Attributes[k]; !ok {
			s.Attributes[k] = v
		}
	}

	modelCfg = NormalizeModelNormalizationConfig(modelCfg)
	if modelCfg.Enabled {
		if len(s.ModelUsage) == 0 {
			s.ModelUsage = BuildModelUsageFromSnapshotMetrics(s)
		}
		s.ModelUsage = normalizeModelUsageRecords(s, modelCfg)
	}
	normalizeAnalyticsMetrics(&s)
	normalizeAnalyticsDailySeries(&s)

	return s
}

func isDiagnosticKey(key string) bool {
	start := 0
	for start < len(key) && (key[start] == ' ' || key[start] == '\t' || key[start] == '\n' || key[start] == '\r') {
		start++
	}
	end := len(key)
	for end > start && (key[end-1] == ' ' || key[end-1] == '\t' || key[end-1] == '\n' || key[end-1] == '\r') {
		end--
	}
	if start >= end {
		return false
	}
	s := key[start:end]
	return containsFoldASCII(s, "error") ||
		containsFoldASCII(s, "warning") ||
		hasSuffixFoldASCII(s, "_err") ||
		hasSuffixFoldASCII(s, "_warn")
}

func containsFoldASCII(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	if len(substr) == 0 {
		return true
	}
	limit := len(s) - len(substr)
	for i := 0; i <= limit; i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			c1 := s[i+j]
			c2 := substr[j]
			if c1 >= 'A' && c1 <= 'Z' {
				c1 += 'a' - 'A'
			}
			if c2 >= 'A' && c2 <= 'Z' {
				c2 += 'a' - 'A'
			}
			if c1 != c2 {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func hasSuffixFoldASCII(s, suffix string) bool {
	if len(suffix) > len(s) {
		return false
	}
	start := len(s) - len(suffix)
	for i := 0; i < len(suffix); i++ {
		c1 := s[start+i]
		c2 := suffix[i]
		if c1 >= 'A' && c1 <= 'Z' {
			c1 += 'a' - 'A'
		}
		if c2 >= 'A' && c2 <= 'Z' {
			c2 += 'a' - 'A'
		}
		if c1 != c2 {
			return false
		}
	}
	return true
}

func normalizeModelUsageRecords(s UsageSnapshot, cfg ModelNormalizationConfig) []ModelUsageRecord {
	if len(s.ModelUsage) == 0 {
		return nil
	}

	out := make([]ModelUsageRecord, 0, len(s.ModelUsage))
	for _, rec := range s.ModelUsage {
		rec.RawModelID = strings.TrimSpace(rec.RawModelID)
		if rec.RawModelID == "" {
			continue
		}
		if rec.Window == "" {
			rec.Window = "unknown"
		}
		if s.ProviderID != "" {
			rec.SetDimension("provider_id", s.ProviderID)
		}
		if s.AccountID != "" {
			rec.SetDimension("account_id", s.AccountID)
		}

		if rec.TotalTokens == nil {
			total := float64(0)
			hasAny := false
			if rec.InputTokens != nil {
				total += *rec.InputTokens
				hasAny = true
			}
			if rec.OutputTokens != nil {
				total += *rec.OutputTokens
				hasAny = true
			}
			if hasAny {
				rec.TotalTokens = Float64Ptr(total)
			}
		}

		identity := normalizeCanonicalModel(s.ProviderID, rec.RawModelID, cfg)
		rec.CanonicalLineageID = identity.LineageID
		rec.CanonicalReleaseID = identity.ReleaseID
		rec.CanonicalVendor = identity.Vendor
		rec.CanonicalFamily = identity.Family
		rec.CanonicalVariant = identity.Variant
		rec.Canonical = identity.Canonical
		rec.Confidence = identity.Confidence
		rec.Reason = identity.Reason
		groupID := rec.CanonicalLineageID
		if cfg.GroupBy == ModelNormalizationGroupRelease && rec.CanonicalReleaseID != "" {
			groupID = rec.CanonicalReleaseID
		}
		if groupID != "" && rec.Confidence >= cfg.MinConfidence {
			rec.SetDimension("canonical_group_id", groupID)
		}
		out = append(out, rec)
	}

	return out
}
