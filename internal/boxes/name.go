package boxes

import (
	"fmt"
	"path/filepath"
	"strings"
)

// validateBoxName rejects empty names, flag-like names, path components, and
// "." / ".." so Join(root, name) cannot escape the containers directory.
func validateBoxName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("box name cannot be empty")
	}
	if name == "." || name == ".." {
		return fmt.Errorf("invalid box name: %q", name)
	}
	if strings.HasPrefix(name, "-") || strings.ContainsAny(name, " /\\:") {
		return fmt.Errorf("invalid box name: %q", name)
	}
	// filepath.Base("a/b") == "b"; also catches Windows separators after Clean.
	if filepath.Base(name) != name {
		return fmt.Errorf("invalid box name: %q", name)
	}
	return nil
}

// resolveBoxProfileDir returns root/name only when name is a single safe
// path segment that stays strictly under root after cleaning.
func resolveBoxProfileDir(root, name string) (string, error) {
	if err := validateBoxName(name); err != nil {
		return "", err
	}
	root = filepath.Clean(strings.TrimSpace(root))
	if root == "" || root == "." {
		return "", fmt.Errorf("cannot determine containers directory")
	}
	profileDir := filepath.Join(root, name)
	rel, err := filepath.Rel(root, profileDir)
	if err != nil || rel == "." || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return "", fmt.Errorf("invalid box name: %q", name)
	}
	if filepath.Base(rel) != rel {
		return "", fmt.Errorf("invalid box name: %q", name)
	}
	return profileDir, nil
}
