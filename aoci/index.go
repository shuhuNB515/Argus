package aoci

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	ManifestName   = "aoci.txt"
	MetaVolumeName = "aoci.meta.txt"
	CodeVolumeName = "aoci.code.txt"
	stateName      = ".aoci.state.json"
	activeName     = ".aoci.active.json"
	StateMissing   = "missing"
	StateAligned   = "aligned"
	StateStale     = "stale"
	maxFileBytes   = 256 << 10
	maxIndexBytes  = 16 << 20
)

type Status struct {
	State       string `json:"state"`
	Root        string `json:"root"`
	Files       int    `json:"files"`
	Bytes       int64  `json:"bytes"`
	GeneratedAt string `json:"generated_at,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
}

type stateFile struct {
	GeneratedAt string `json:"generated_at"`
	Fingerprint string `json:"fingerprint"`
	Files       int    `json:"files"`
	Bytes       int64  `json:"bytes"`
}

type activeFile struct {
	Project string `json:"project"`
}

type sourceFile struct {
	rel  string
	abs  string
	size int64
	hash string
	data []byte
}

var symbolPattern = regexp.MustCompile(`(?m)^\s*(?:func|type|class|interface|def|function|export\s+(?:const|function|class)|CREATE\s+TABLE)\s+([A-Za-z_][A-Za-z0-9_]*)`)

var sourceExtensions = map[string]bool{
	".c": true, ".cc": true, ".cpp": true, ".cs": true, ".css": true, ".go": true, ".h": true,
	".html": true, ".java": true, ".js": true, ".jsx": true, ".json": true, ".kt": true, ".md": true,
	".php": true, ".py": true, ".rb": true, ".rs": true, ".scss": true, ".sh": true, ".sql": true,
	".swift": true, ".toml": true, ".ts": true, ".tsx": true, ".vue": true, ".xml": true, ".yaml": true, ".yml": true,
}

var ignoredDirs = map[string]bool{
	".git": true, ".idea": true, ".next": true, ".turbo": true, ".venv": true, "build": true,
	"coverage": true, "dist": true, "node_modules": true, "out": true, "target": true, "vendor": true,
}

func scan(root string, includeData bool) ([]sourceFile, string, int64, error) {
	root = filepath.Clean(root)
	files := make([]sourceFile, 0, 128)
	var total int64
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != root && (ignoredDirs[entry.Name()] || strings.HasPrefix(entry.Name(), ".aoci")) {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		name := strings.ToLower(entry.Name())
		if name == strings.ToLower(ManifestName) || name == strings.ToLower(MetaVolumeName) || name == strings.ToLower(CodeVolumeName) || name == stateName || name == activeName || strings.HasPrefix(name, ".env") || strings.Contains(name, "credential") || strings.Contains(name, "secret") || strings.HasSuffix(name, ".pem") || strings.HasSuffix(name, ".key") {
			return nil
		}
		if !sourceExtensions[strings.ToLower(filepath.Ext(name))] {
			return nil
		}
		info, err := entry.Info()
		if err != nil || info.Size() > maxFileBytes {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil || bytesContainZero(data) {
			return nil
		}
		sum := sha256.Sum256(data)
		rel, _ := filepath.Rel(root, path)
		item := sourceFile{rel: filepath.ToSlash(rel), abs: path, size: info.Size(), hash: hex.EncodeToString(sum[:])}
		if includeData {
			item.data = data
		}
		files = append(files, item)
		total += info.Size()
		return nil
	})
	if err != nil {
		return nil, "", 0, err
	}
	sort.Slice(files, func(i, j int) bool { return files[i].rel < files[j].rel })
	h := sha256.New()
	for _, file := range files {
		fmt.Fprintf(h, "%s\x00%d\x00%s\n", file.rel, file.size, file.hash)
	}
	return files, hex.EncodeToString(h.Sum(nil)), total, nil
}

func bytesContainZero(data []byte) bool {
	for _, b := range data {
		if b == 0 {
			return true
		}
	}
	return false
}

func Build(root string) (Status, error) {
	root = filepath.Clean(root)
	files, fingerprint, total, err := scan(root, true)
	if err != nil {
		return Status{}, err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	project := filepath.Base(root)
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
		project = "Argus"
	}
	manifest := fmt.Sprintf("#AOCI-ROOT-MANIFEST: 1\n#Format-Version: cognition-volumes/v1\n#Locale: zh-CN\n#Project: %s\n#Generated-At: %s\n#Fingerprint: %s\n#Volume: id=meta kind=meta path=%s format=meta-v1 depends=- state=enabled\n#Volume: id=code kind=code path=%s format=object-fras-v2 depends=meta state=enabled\n", project, now, fingerprint, MetaVolumeName, CodeVolumeName)

	var meta strings.Builder
	meta.WriteString("#AOCI-META-VOLUME: 1\n#Object-Protocol: repository-cognition-object/v2\n#FRAS-Discipline: 2\n#Object-Kinds: code=file database=table\n")
	var code strings.Builder
	code.WriteString("#AOCI-CODE-VOLUME: 1\n#Whole-Index: repository source snapshot\n")
	remaining := maxIndexBytes
	for _, file := range files {
		symbols := symbolPattern.FindAllSubmatch(file.data, 24)
		names := make([]string, 0, len(symbols))
		for _, match := range symbols {
			if len(match) > 1 {
				names = append(names, string(match[1]))
			}
		}
		fmt.Fprintf(&meta, "%s[%s]: F:source file | R:%s | A:%s | S:%d bytes\n", file.rel, file.hash[:8], languageOf(file.rel), strings.Join(names, ","), file.size)
		if remaining <= 0 {
			continue
		}
		chunk := file.data
		// Keep the generated whole-index portable and diff-friendly even when a
		// source file uses Windows line endings.
		chunk = bytes.ReplaceAll(chunk, []byte("\r\n"), []byte("\n"))
		if len(chunk) > remaining {
			chunk = chunk[:remaining]
		}
		fmt.Fprintf(&code, "\n@@ %s sha256=%s bytes=%d\n", file.rel, file.hash, file.size)
		code.Write(chunk)
		if len(chunk) == 0 || chunk[len(chunk)-1] != '\n' {
			code.WriteByte('\n')
		}
		remaining -= len(chunk)
	}
	state := stateFile{GeneratedAt: now, Fingerprint: fingerprint, Files: len(files), Bytes: total}
	stateJSON, _ := json.MarshalIndent(state, "", "  ")
	for name, data := range map[string][]byte{
		ManifestName: []byte(manifest), MetaVolumeName: []byte(meta.String()), CodeVolumeName: []byte(code.String()), stateName: stateJSON,
	} {
		if err := atomicWrite(filepath.Join(root, name), data); err != nil {
			return Status{}, err
		}
	}
	return Status{State: StateAligned, Root: root, Files: len(files), Bytes: total, GeneratedAt: now, Fingerprint: fingerprint}, nil
}

func StatusOf(root string) Status {
	root = filepath.Clean(root)
	status := Status{State: StateMissing, Root: root}
	data, err := os.ReadFile(filepath.Join(root, stateName))
	if err != nil {
		return status
	}
	var saved stateFile
	if json.Unmarshal(data, &saved) != nil {
		return status
	}
	_, fingerprint, total, err := scan(root, false)
	if err != nil {
		return status
	}
	status.Files, status.Bytes, status.GeneratedAt, status.Fingerprint = saved.Files, total, saved.GeneratedAt, fingerprint
	if fingerprint == saved.Fingerprint {
		status.State = StateAligned
	} else {
		status.State = StateStale
	}
	return status
}

func RuntimeContext(root string) string {
	var out strings.Builder
	for _, name := range []string{ManifestName, MetaVolumeName} {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			continue
		}
		out.Write(data)
		if len(data) > 0 && data[len(data)-1] != '\n' {
			out.WriteByte('\n')
		}
	}
	return out.String()
}

func SetActiveProject(workspaceRoot, projectRoot string) error {
	workspaceRoot = filepath.Clean(workspaceRoot)
	projectRoot = filepath.Clean(projectRoot)
	rel, err := filepath.Rel(workspaceRoot, projectRoot)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("AOCI project is outside the workspace")
	}
	data, _ := json.MarshalIndent(activeFile{Project: filepath.ToSlash(rel)}, "", "  ")
	return atomicWrite(filepath.Join(workspaceRoot, activeName), data)
}

// SetExternalActiveProject is used by the local runtime when its data directory
// lives inside a separate source checkout. The workspace API continues to use
// SetActiveProject, which remains confined to the browsable workspace root.
func SetExternalActiveProject(workspaceRoot, projectRoot string) error {
	workspaceRoot = filepath.Clean(workspaceRoot)
	projectRoot = filepath.Clean(projectRoot)
	if !filepath.IsAbs(projectRoot) {
		return fmt.Errorf("AOCI external project must be absolute")
	}
	info, err := os.Stat(projectRoot)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("AOCI external project is not a directory")
	}
	data, _ := json.MarshalIndent(activeFile{Project: filepath.ToSlash(projectRoot)}, "", "  ")
	return atomicWrite(filepath.Join(workspaceRoot, activeName), data)
}

func RuntimeContextFor(start string) string {
	root := runtimeRoot(start)
	if StatusOf(root).State != StateAligned {
		return ""
	}
	return RuntimeContext(root)
}

// RuntimeStatusFor resolves the active AOCI project just like RuntimeContextFor
// and reports whether the index is still aligned with the source tree. Agents
// use this as an explicit runtime contract instead of trusting generated files.
func RuntimeStatusFor(start string) Status {
	root := runtimeRoot(start)
	return StatusOf(root)
}

func runtimeRoot(start string) string {
	current := filepath.Clean(start)
	for i := 0; i < 16; i++ {
		data, err := os.ReadFile(filepath.Join(current, activeName))
		if err == nil {
			var active activeFile
			if json.Unmarshal(data, &active) == nil && active.Project != "" {
				if filepath.IsAbs(filepath.FromSlash(active.Project)) {
					return filepath.Clean(filepath.FromSlash(active.Project))
				}
				project := filepath.Clean(filepath.FromSlash(active.Project))
				project = filepath.Clean(filepath.Join(current, project))
				if rel, relErr := filepath.Rel(current, project); relErr == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
					return project
				}
			}
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return filepath.Clean(start)
}

func atomicWrite(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".aoci-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if _, err = tmp.Write(data); err == nil {
		err = tmp.Sync()
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(name, path)
}

func languageOf(path string) string {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	if ext == "" {
		return "text"
	}
	return ext
}
