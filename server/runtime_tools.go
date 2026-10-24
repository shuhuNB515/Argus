package server

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

// detectLocalExecutable resolves the executable from the service process PATH.
// The desktop service inherits the user's local environment, so this reports the
// binary that command execution will actually use.
func detectLocalExecutable(name string) (string, bool) {
	path, err := exec.LookPath(name)
	if err == nil {
		return path, true
	}
	// Apps opened from Finder often receive a minimal PATH that omits Homebrew.
	// Check the standard macOS package-manager locations explicitly as well.
	for _, dir := range []string{"/opt/homebrew/bin", "/usr/local/bin", "/usr/bin"} {
		candidate := filepath.Join(dir, name)
		info, statErr := os.Stat(candidate)
		if statErr == nil && !info.IsDir() && info.Mode()&0o111 != 0 {
			return candidate, true
		}
	}
	return "", false
}

func (s *Server) pgDetectNmap(w http.ResponseWriter, r *http.Request) {
	path, ok := detectLocalExecutable("nmap")
	if !ok {
		writeJSON(w, http.StatusOK, map[string]any{
			"available": false,
			"source":    "local",
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"available": true,
		"path":      path,
		"source":    "local",
	})
}
