//go:build windows

package antigravity

func lockCredentialFile(path string) (func(), error) {
	return func() {}, nil
}
