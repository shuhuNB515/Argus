package server

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Autumn-27/artex/aoci"
)

func (s *Server) aociRoot(w http.ResponseWriter, rel string) (string, bool) {
	abs, ok := s.wsResolve(rel)
	if !ok {
		writeErr(w, http.StatusBadRequest, errWsIllegalPath)
		return "", false
	}
	info, err := os.Stat(abs)
	if err != nil {
		writeErr(w, http.StatusNotFound, errWsPathNotFound)
		return "", false
	}
	if !info.IsDir() {
		writeErr(w, http.StatusBadRequest, errWsNotDir)
		return "", false
	}
	return abs, true
}

func (s *Server) aociStatus(w http.ResponseWriter, r *http.Request) {
	if path := filepath.Clean(strings.TrimSpace(r.URL.Query().Get("path"))); path == "." || path == "" {
		status := aoci.RuntimeStatusFor(s.m.dir)
		writeJSON(w, http.StatusOK, struct {
			aoci.Status
			RuntimeRoot string `json:"runtime_root,omitempty"`
		}{Status: status, RuntimeRoot: status.Root})
		return
	}
	abs, ok := s.aociRoot(w, r.URL.Query().Get("path"))
	if !ok {
		return
	}
	status := aoci.StatusOf(abs)
	if rel, err := filepath.Rel(filepath.Clean(s.m.dir), filepath.Clean(abs)); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		status.Root = filepath.ToSlash(rel)
	} else {
		status.Root = abs
	}
	writeJSON(w, http.StatusOK, status)
}

func (s *Server) aociBuild(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path      string `json:"path"`
		ReloadMCP bool   `json:"reload_mcp"`
	}
	if err := decode(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	var abs string
	var ok bool
	if strings.TrimSpace(body.Path) == "" || filepath.Clean(strings.TrimSpace(body.Path)) == "." {
		status := aoci.RuntimeStatusFor(s.m.dir)
		abs, ok = status.Root, status.Root != ""
		if !ok {
			writeErr(w, http.StatusNotFound, errWsPathNotFound)
			return
		}
	} else {
		abs, ok = s.aociRoot(w, body.Path)
		if !ok {
			return
		}
	}
	status, err := aoci.Build(abs)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	var activeErr error
	if rel, err := filepath.Rel(filepath.Clean(s.m.dir), filepath.Clean(abs)); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		activeErr = aoci.SetActiveProject(s.m.dir, abs)
	} else {
		activeErr = aoci.SetExternalActiveProject(s.m.dir, abs)
	}
	if activeErr != nil {
		writeErr(w, http.StatusInternalServerError, activeErr.Error())
		return
	}
	status.Root = s.wsRel(abs)
	reloaded, reloadErrors := 0, []string{}
	if body.ReloadMCP {
		reloaded, reloadErrors = s.reloadEnabledMCPs(r.Context())
	}
	writeJSON(w, http.StatusOK, struct {
		aoci.Status
		MCPReloaded int      `json:"mcp_reloaded"`
		MCPErrors   []string `json:"mcp_errors"`
	}{Status: status, MCPReloaded: reloaded, MCPErrors: reloadErrors})
}

func (s *Server) reloadEnabledMCPs(parent context.Context) (int, []string) {
	if s.m == nil || s.m.pg == nil {
		return 0, []string{}
	}
	servers, err := s.m.pg.ListMCP()
	if err != nil {
		return 0, []string{err.Error()}
	}
	reloaded := 0
	errs := make([]string, 0)
	for _, server := range servers {
		if !server.Enabled {
			continue
		}
		ctx, cancel := context.WithTimeout(parent, 45*time.Second)
		err := s.discoverAndCacheMCP(ctx, server)
		cancel()
		if err != nil {
			errs = append(errs, server.Name+": "+err.Error())
			continue
		}
		reloaded++
	}
	return reloaded, errs
}

func (s *Server) aociContextFor(rel string) string {
	abs, ok := s.wsResolve(rel)
	if !ok {
		return ""
	}
	if info, err := os.Stat(abs); err != nil || !info.IsDir() {
		return ""
	}
	return aoci.RuntimeContext(filepath.Clean(abs))
}
