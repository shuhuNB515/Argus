package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Autumn-27/artex/report"
)

type browserEvidenceRecord struct {
	ID             string          `json:"id"`
	TaskID         string          `json:"task_id"`
	RunID          string          `json:"run_id"`
	NodeID         string          `json:"node_id"`
	Kind           string          `json:"kind"`
	URL            string          `json:"url"`
	Timestamp      time.Time       `json:"timestamp"`
	PageState      json.RawMessage `json:"page_state,omitempty"`
	ScreenshotPath string          `json:"screenshot_path,omitempty"`
	Provenance     string          `json:"provenance"`
}

type browserEvidenceRequest struct {
	RunID          string          `json:"run_id"`
	NodeID         string          `json:"node_id"`
	Kind           string          `json:"kind"`
	URL            string          `json:"url"`
	Timestamp      time.Time       `json:"timestamp"`
	PageState      json.RawMessage `json:"page_state"`
	ScreenshotPath string          `json:"screenshot_path"`
	Provenance     string          `json:"provenance"`
}

func (s *Server) browserEvidenceDir(taskID string) string {
	return filepath.Join(s.m.dir, "tasks", taskID, "evidence", "browser")
}

func (s *Server) browserEvidence(w http.ResponseWriter, r *http.Request) {
	t := s.m.ResolveTask(r.PathValue("id"))
	if t == nil {
		writeErr(w, http.StatusNotFound, "task not found")
		return
	}
	if r.Method == http.MethodGet {
		s.browserEvidenceList(w, t.ID)
		return
	}
	var req browserEvidenceRequest
	if err := decode(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.URL) == "" || strings.TrimSpace(req.Kind) == "" {
		writeErr(w, http.StatusBadRequest, "url and kind are required")
		return
	}
	taskID, err := parseTaskID(t.ID)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	policy, err := s.srcPolicyForTask(taskID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := policy.AuthorizeURL(req.URL); err != nil {
		writeErr(w, http.StatusForbidden, "browser evidence URL is outside the task target")
		return
	}
	now := req.Timestamp
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if now.Location() != time.UTC {
		now = now.UTC()
	}
	id := now.Format("20060102T150405.000000000Z")
	dir := s.browserEvidenceDir(t.ID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	record := browserEvidenceRecord{
		ID: id, TaskID: t.ID, RunID: req.RunID, NodeID: req.NodeID, Kind: req.Kind,
		URL: req.URL, Timestamp: now, PageState: req.PageState,
		Provenance: nonEmpty(req.Provenance, "browser/authorized-readonly"),
	}
	if req.ScreenshotPath != "" {
		path, err := copyBrowserScreenshot(req.ScreenshotPath, dir, id)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		record.ScreenshotPath = path
	}
	data, _ := json.MarshalIndent(record, "", "  ")
	if err := os.WriteFile(filepath.Join(dir, id+".json"), append(data, '\n'), 0o644); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, record)
}

func (s *Server) browserEvidenceList(w http.ResponseWriter, taskID string) {
	refs := loadBrowserEvidence(s.browserEvidenceDir(taskID))
	writeJSON(w, http.StatusOK, refs)
}

func loadBrowserEvidence(dir string) []report.EvidenceRef {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []report.EvidenceRef{}
	}
	refs := make([]report.EvidenceRef, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}
		var item browserEvidenceRecord
		if json.Unmarshal(data, &item) != nil || item.ID == "" {
			continue
		}
		refs = append(refs, report.EvidenceRef{ID: item.ID, Kind: item.Kind, URL: item.URL, RunID: item.RunID, NodeID: item.NodeID, Timestamp: item.Timestamp, Provenance: item.Provenance})
	}
	sort.SliceStable(refs, func(i, j int) bool { return refs[i].Timestamp.Before(refs[j].Timestamp) })
	return refs
}

func copyBrowserScreenshot(source, dir, id string) (string, error) {
	path := filepath.Clean(source)
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("screenshot is not readable")
	}
	if !info.Mode().IsRegular() || info.Size() > 12<<20 {
		return "", fmt.Errorf("screenshot must be a regular file no larger than 12 MB")
	}
	in, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("screenshot is not readable")
	}
	defer in.Close()
	name := id + filepath.Ext(path)
	dst := filepath.Join(dir, name)
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o644)
	if err != nil {
		return "", err
	}
	if _, err = io.Copy(out, io.LimitReader(in, 12<<20+1)); err != nil {
		_ = out.Close()
		_ = os.Remove(dst)
		return "", err
	}
	if err = out.Close(); err != nil {
		_ = os.Remove(dst)
		return "", err
	}
	return filepath.ToSlash(filepath.Join("evidence", "browser", name)), nil
}

func nonEmpty(value, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}
