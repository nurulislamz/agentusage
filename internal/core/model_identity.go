package core

import (
	"strings"
)

type canonicalModelIdentity struct {
	LineageID  string
	ReleaseID  string
	Vendor     string
	Family     string
	Variant    string
	Confidence float64
	Reason     string
	Canonical  string // Canonical model name for consistent identification
}

func normalizeCanonicalModel(providerID, rawModelID string, cfg ModelNormalizationConfig) canonicalModelIdentity {
	raw := strings.TrimSpace(rawModelID)
	if raw == "" {
		return canonicalModelIdentity{
			LineageID:  "unknown/unknown",
			Vendor:     "unknown",
			Family:     "unknown",
			Confidence: 0.10,
			Reason:     "empty",
		}
	}

	if ov, ok := findModelOverride(providerID, raw, cfg.Overrides); ok {
		lineage := strings.TrimSpace(ov.CanonicalLineage)
		if lineage == "" {
			lineage = "unknown/" + normalizeModelToken(raw)
		}
		release := strings.TrimSpace(ov.CanonicalRelease)
		vendor, family := parseVendorFamilyFromCanonical(lineage)
		identity := canonicalModelIdentity{
			LineageID:  lineage,
			ReleaseID:  release,
			Vendor:     vendor,
			Family:     family,
			Variant:    parseVariantFromCanonical(lineage),
			Confidence: 1.0,
			Reason:     "override",
			Canonical:  ov.CanonicalModel, // Add canonical model name from override
		}
		return identity
	}

	vendorFromProvider := canonicalVendorFromProvider(providerID)
	model := strings.ToLower(strings.TrimSpace(raw))
	model = strings.TrimPrefix(model, "models/")
	model = strings.Trim(model, "/")

	explicitVendor := ""
	if slashIdx := strings.IndexByte(model, '/'); slashIdx >= 0 {
		prefix := model[:slashIdx]
		if isKnownVendor(prefix) {
			explicitVendor = prefix
			model = model[slashIdx+1:]
		}
	}

	releaseDate := extractReleaseDate(model)
	modelNoDate := stripReleaseDate(model)
	if modelNoDate == "" {
		modelNoDate = model
	}
	norm := normalizeModelToken(modelNoDate)
	if norm == "" {
		norm = "unknown"
	}
	tokens := splitModelTokens(norm)

	vendor := explicitVendor
	if vendor == "" {
		vendor = detectVendorFromModel(tokens, vendorFromProvider)
	}
	family := detectFamily(tokens)
	variant := detectVariant(tokens)

	identity := canonicalModelIdentity{
		Vendor:     vendor,
		Family:     family,
		Variant:    variant,
		Confidence: 0.60,
		Reason:     "heuristic",
	}

	switch family {
	case "claude":
		claude := canonicalizeClaude(tokens)
		identity.Vendor = FirstNonEmpty(identity.Vendor, "anthropic")
		identity.Family = "claude"
		identity.Variant = FirstNonEmpty(claude.variant, identity.Variant)
		identity.LineageID = identity.Vendor + "/" + claude.lineage
		identity.Confidence = claude.confidence
		identity.Reason = claude.reason
		identity.Canonical = identity.Vendor + "/" + claude.lineage
	case "gpt":
		gpt := canonicalizeGPT(tokens)
		identity.Vendor = FirstNonEmpty(identity.Vendor, "openai")
		identity.Family = "gpt"
		identity.Variant = FirstNonEmpty(gpt.variant, identity.Variant)
		identity.LineageID = identity.Vendor + "/" + gpt.lineage
		identity.Confidence = gpt.confidence
		identity.Reason = gpt.reason
		identity.Canonical = "openai/gpt-" + FirstNonEmpty(gpt.variant, "unknown")
	case "gemini":
		gem := canonicalizeGemini(tokens)
		identity.Vendor = FirstNonEmpty(identity.Vendor, "google")
		identity.Family = "gemini"
		identity.Variant = FirstNonEmpty(gem.variant, identity.Variant)
		identity.LineageID = identity.Vendor + "/" + gem.lineage
		identity.Confidence = gem.confidence
		identity.Reason = gem.reason
		identity.Canonical = "google/gemini-" + FirstNonEmpty(gem.variant, "unknown")
	default:
		v := identity.Vendor
		if v == "" {
			v = "unknown"
		}
		identity.Vendor = v
		if identity.Family == "" {
			identity.Family = "unknown"
		}
		identity.LineageID = v + "/" + norm
		if explicitVendor != "" {
			identity.Confidence = 0.90
			identity.Reason = "explicit_vendor"
		} else if v != "unknown" {
			identity.Confidence = 0.72
			identity.Reason = "provider_vendor"
		}
		identity.Canonical = v + "/" + norm
	}

	if releaseDate != "" {
		identity.ReleaseID = identity.LineageID + "@" + releaseDate
	}

	return identity
}

type canonicalBuild struct {
	lineage    string
	variant    string
	confidence float64
	reason     string
}

func canonicalizeClaude(tokens []string) canonicalBuild {
	variant := firstMatch(tokens, "opus", "sonnet", "haiku")
	version := extractVersionNearVariant(tokens, variant)
	if variant != "" && version != "" {
		return canonicalBuild{
			lineage:    "claude-" + variant + "-" + version,
			variant:    variant,
			confidence: 0.95,
			reason:     "family_parse",
		}
	}
	if variant != "" {
		return canonicalBuild{
			lineage:    "claude-" + variant,
			variant:    variant,
			confidence: 0.82,
			reason:     "family_parse_variant_only",
		}
	}
	version = firstVersionToken(tokens)
	if version != "" {
		return canonicalBuild{
			lineage:    "claude-" + version,
			confidence: 0.78,
			reason:     "family_parse_version_only",
		}
	}
	return canonicalBuild{
		lineage:    "claude",
		confidence: 0.72,
		reason:     "family_only",
	}
}

func canonicalizeGPT(tokens []string) canonicalBuild {
	version := firstVersionToken(tokens)
	variant := firstMatch(tokens, "codex", "mini", "nano", "turbo", "chat", "pro")
	var lineage string
	if version != "" && variant != "" {
		lineage = "gpt-" + version + "-" + variant
	} else if version != "" {
		lineage = "gpt-" + version
	} else if variant != "" {
		lineage = "gpt-" + variant
	} else {
		lineage = "gpt"
	}
	confidence := 0.80
	if version != "" {
		confidence = 0.90
	}
	if variant != "" && version != "" {
		confidence = 0.93
	}
	return canonicalBuild{
		lineage:    lineage,
		variant:    variant,
		confidence: confidence,
		reason:     "family_parse",
	}
}

func canonicalizeGemini(tokens []string) canonicalBuild {
	version := firstVersionToken(tokens)
	variant := firstMatch(tokens, "pro", "flash", "ultra", "nano", "lite")
	var lineage string
	if version != "" && variant != "" {
		lineage = "gemini-" + version + "-" + variant
	} else if version != "" {
		lineage = "gemini-" + version
	} else if variant != "" {
		lineage = "gemini-" + variant
	} else {
		lineage = "gemini"
	}
	confidence := 0.80
	if version != "" {
		confidence = 0.88
	}
	return canonicalBuild{
		lineage:    lineage,
		variant:    variant,
		confidence: confidence,
		reason:     "family_parse",
	}
}

func findModelOverride(providerID, rawModelID string, overrides []ModelNormalizationOverride) (ModelNormalizationOverride, bool) {
	targetProvider := strings.ToLower(strings.TrimSpace(providerID))
	targetModel := strings.ToLower(strings.TrimSpace(rawModelID))
	for _, ov := range overrides {
		modelMatch := strings.ToLower(strings.TrimSpace(ov.RawModelID)) == targetModel
		if !modelMatch {
			continue
		}
		ovProvider := strings.ToLower(strings.TrimSpace(ov.Provider))
		if ovProvider == "" || ovProvider == targetProvider {
			return ov, true
		}
	}
	return ModelNormalizationOverride{}, false
}

func canonicalVendorFromProvider(providerID string) string {
	switch strings.ToLower(strings.TrimSpace(providerID)) {
	case "anthropic", "claude_code":
		return "anthropic"
	case "openai", "codex":
		return "openai"
	case "gemini_api", "gemini_cli":
		return "google"
	case "mistral":
		return "mistral"
	case "xai":
		return "xai"
	case "deepseek":
		return "deepseek"
	case "groq":
		return "groq"
	case "openrouter":
		return "openrouter"
	case "cursor":
		return "cursor"
	case "copilot":
		return "copilot"
	default:
		return ""
	}
}

func isKnownVendor(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "anthropic", "openai", "google", "mistral", "xai", "deepseek", "groq", "meta", "openrouter":
		return true
	default:
		return false
	}
}

func detectVendorFromModel(tokens []string, fallback string) string {
	if containsToken(tokens, "claude") {
		return "anthropic"
	}
	if containsToken(tokens, "gpt") || containsToken(tokens, "codex") {
		return "openai"
	}
	if containsToken(tokens, "gemini") {
		return "google"
	}
	if containsToken(tokens, "grok") {
		return "xai"
	}
	if containsToken(tokens, "mistral") || containsToken(tokens, "mixtral") || containsToken(tokens, "codestral") {
		return "mistral"
	}
	if containsToken(tokens, "deepseek") {
		return "deepseek"
	}
	if containsToken(tokens, "llama") {
		return "meta"
	}
	if fallback != "" {
		return fallback
	}
	return "unknown"
}

func detectFamily(tokens []string) string {
	switch {
	case containsToken(tokens, "claude"):
		return "claude"
	case containsToken(tokens, "gemini"):
		return "gemini"
	case containsToken(tokens, "gpt"), containsToken(tokens, "codex"):
		return "gpt"
	case containsToken(tokens, "grok"):
		return "grok"
	default:
		return ""
	}
}

func detectVariant(tokens []string) string {
	return firstMatch(tokens,
		"opus", "sonnet", "haiku",
		"mini", "nano", "turbo", "pro", "flash", "ultra",
		"codex",
	)
}

func isWordByte(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}

func matchISODate(s string, i int) bool {
	if i+10 > len(s) {
		return false
	}
	if s[i] != '2' || s[i+1] != '0' || s[i+2] < '0' || s[i+2] > '9' || s[i+3] < '0' || s[i+3] > '9' {
		return false
	}
	sep1 := s[i+4]
	if sep1 != '-' && sep1 != '_' {
		return false
	}
	m0, m1 := s[i+5], s[i+6]
	if !((m0 == '0' && m1 >= '1' && m1 <= '9') || (m0 == '1' && m1 >= '0' && m1 <= '2')) {
		return false
	}
	sep2 := s[i+7]
	if sep2 != '-' && sep2 != '_' {
		return false
	}
	d0, d1 := s[i+8], s[i+9]
	if !((d0 == '0' && d1 >= '1' && d1 <= '9') ||
		((d0 == '1' || d0 == '2') && d1 >= '0' && d1 <= '9') ||
		(d0 == '3' && (d1 == '0' || d1 == '1'))) {
		return false
	}
	return true
}

func matchCompactDate(s string, i int) bool {
	if i+8 > len(s) {
		return false
	}
	if i > 0 && isWordByte(s[i-1]) {
		return false
	}
	if i+8 < len(s) && isWordByte(s[i+8]) {
		return false
	}
	if s[i] != '2' || s[i+1] != '0' || s[i+2] < '0' || s[i+2] > '9' || s[i+3] < '0' || s[i+3] > '9' {
		return false
	}
	m0, m1 := s[i+4], s[i+5]
	if !((m0 == '0' && m1 >= '1' && m1 <= '9') || (m0 == '1' && m1 >= '0' && m1 <= '2')) {
		return false
	}
	d0, d1 := s[i+6], s[i+7]
	if !((d0 == '0' && d1 >= '1' && d1 <= '9') ||
		((d0 == '1' || d0 == '2') && d1 >= '0' && d1 <= '9') ||
		(d0 == '3' && (d1 == '0' || d1 == '1'))) {
		return false
	}
	return true
}

func extractReleaseDate(raw string) string {
	for i := 0; i+10 <= len(raw); i++ {
		if matchISODate(raw, i) {
			var buf [8]byte
			buf[0] = raw[i]
			buf[1] = raw[i+1]
			buf[2] = raw[i+2]
			buf[3] = raw[i+3]
			buf[4] = raw[i+5]
			buf[5] = raw[i+6]
			buf[6] = raw[i+8]
			buf[7] = raw[i+9]
			return string(buf[:])
		}
	}
	for i := 0; i+8 <= len(raw); i++ {
		if matchCompactDate(raw, i) {
			return raw[i : i+8]
		}
	}
	return ""
}

func stripReleaseDate(raw string) string {
	if len(raw) < 8 {
		return strings.Trim(raw, "-_ ")
	}
	hasISO := false
	for i := 0; i+10 <= len(raw); i++ {
		if matchISODate(raw, i) {
			hasISO = true
			break
		}
	}
	hasCompact := false
	if !hasISO {
		for i := 0; i+8 <= len(raw); i++ {
			if matchCompactDate(raw, i) {
				hasCompact = true
				break
			}
		}
		if !hasCompact {
			return strings.Trim(raw, "-_ ")
		}
	}

	var b strings.Builder
	b.Grow(len(raw))
	i := 0
	for i < len(raw) {
		if i+10 <= len(raw) && matchISODate(raw, i) {
			i += 10
			continue
		}
		if i+8 <= len(raw) && matchCompactDate(raw, i) {
			i += 8
			continue
		}
		b.WriteByte(raw[i])
		i++
	}
	return strings.Trim(b.String(), "-_ ")
}

func normalizeModelToken(raw string) string {
	if raw == "" {
		return "unknown"
	}
	var b strings.Builder
	b.Grow(len(raw))
	lastDash := false
	for i := 0; i < len(raw); i++ {
		r := raw[i]
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteByte(r)
			lastDash = false
		case r >= 'A' && r <= 'Z':
			b.WriteByte(r + ('a' - 'A'))
			lastDash = false
		case r >= '0' && r <= '9':
			b.WriteByte(r)
			lastDash = false
		case r == '.':
			b.WriteByte(r)
			lastDash = false
		default:
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "unknown"
	}
	return out
}

func splitModelTokens(model string) []string {
	norm := normalizeModelToken(model)
	if norm == "" || norm == "unknown" {
		return []string{"unknown"}
	}
	count := 1
	for i := 0; i < len(norm); i++ {
		if norm[i] == '-' {
			count++
		}
	}
	out := make([]string, 0, count)
	start := 0
	for i := 0; i < len(norm); i++ {
		if norm[i] == '-' {
			if i > start {
				out = append(out, norm[start:i])
			}
			start = i + 1
		}
	}
	if start < len(norm) {
		out = append(out, norm[start:])
	}
	return out
}

func isVersionToken(tok string) bool {
	if len(tok) == 0 {
		return false
	}
	dotSeen := false
	hasDigitsBefore := false
	hasDigitsAfter := false
	for i := 0; i < len(tok); i++ {
		c := tok[i]
		if c >= '0' && c <= '9' {
			if !dotSeen {
				hasDigitsBefore = true
			} else {
				hasDigitsAfter = true
			}
		} else if c == '.' {
			if dotSeen || !hasDigitsBefore {
				return false
			}
			dotSeen = true
		} else {
			return false
		}
	}
	if dotSeen {
		return hasDigitsBefore && hasDigitsAfter
	}
	return hasDigitsBefore
}

func firstVersionToken(tokens []string) string {
	for i, tok := range tokens {
		if !isVersionToken(tok) {
			continue
		}
		// join major/minor split across adjacent tokens (e.g. 4,6 -> 4.6)
		if !strings.Contains(tok, ".") && i+1 < len(tokens) && isAllDigits(tokens[i+1]) {
			return tok + "." + tokens[i+1]
		}
		return tok
	}
	return ""
}

func extractVersionNearVariant(tokens []string, variant string) string {
	if variant == "" {
		return firstVersionToken(tokens)
	}
	idx := -1
	for i, t := range tokens {
		if t == variant {
			idx = i
			break
		}
	}
	if idx < 0 {
		return firstVersionToken(tokens)
	}
	// right side first
	for i := idx + 1; i < len(tokens); i++ {
		if isVersionToken(tokens[i]) {
			if !strings.Contains(tokens[i], ".") && i+1 < len(tokens) && isAllDigits(tokens[i+1]) {
				return tokens[i] + "." + tokens[i+1]
			}
			return tokens[i]
		}
	}
	// then left side
	for i := idx - 1; i >= 0; i-- {
		if isVersionToken(tokens[i]) {
			if !strings.Contains(tokens[i], ".") && i+1 < len(tokens) && isAllDigits(tokens[i+1]) && i+1 == idx-0 {
				return tokens[i] + "." + tokens[i+1]
			}
			return tokens[i]
		}
	}
	return ""
}

func parseVendorFamilyFromCanonical(lineage string) (vendor, family string) {
	lineage = strings.TrimSpace(lineage)
	if lineage == "" {
		return "unknown", "unknown"
	}
	slashIdx := strings.IndexByte(lineage, '/')
	if slashIdx >= 0 {
		vendor = lineage[:slashIdx]
		rest := lineage[slashIdx+1:]
		dashIdx := strings.IndexByte(rest, '-')
		if dashIdx >= 0 {
			family = rest[:dashIdx]
		} else {
			family = rest
		}
		return vendor, family
	}
	dashIdx := strings.IndexByte(lineage, '-')
	if dashIdx >= 0 {
		family = lineage[:dashIdx]
	} else {
		family = lineage
	}
	return "unknown", family
}

func parseVariantFromCanonical(lineage string) string {
	model := lineage
	if slashIdx := strings.IndexByte(lineage, '/'); slashIdx >= 0 {
		model = lineage[slashIdx+1:]
	}
	tokens := splitModelTokens(model)
	return detectVariant(tokens)
}

func containsToken(tokens []string, target string) bool {
	for _, tok := range tokens {
		if tok == target {
			return true
		}
	}
	return false
}

func firstMatch(tokens []string, candidates ...string) string {
	for _, candidate := range candidates {
		if containsToken(tokens, candidate) {
			return candidate
		}
	}
	return ""
}

// FirstNonEmpty returns the first non-blank string from values (trimmed).
func FirstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
