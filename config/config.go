// Package config loads runtime configuration from a JSON file, with environment
// variables taking precedence. Currently it carries the PostgreSQL connection.
package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// Database is the PostgreSQL connection config. Either set DSN directly, or set
// the component fields and a DSN is assembled from them.
type Database struct {
	DSN      string `json:"dsn"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	SSLMode  string `json:"sslmode"`
}

// Config is the on-disk config file shape.
type Config struct {
	Database Database `json:"database"`
	SkillDir string   `json:"skill_dir"`
	// Language is the bootstrap/default user-facing output language for Argus
	// ("zh"/"en"). Empty → DefaultLanguage. At
	// runtime a value stored in the settings table (edited in the UI) overrides it.
	Language string `json:"language"`
}

// SupportedLanguages keeps legacy configuration values readable for existing
// data, while the shipped UI and default language expose only Chinese and English.
var SupportedLanguages = []string{"zh", "en"}

// DefaultLanguage is the out-of-box user-facing language when nothing is
// configured. Argus defaults to Simplified Chinese; English is also available.
const DefaultLanguage = "zh"

// NormalizeLanguage lower-cases and validates a language code, returning "" when
// the code is empty or unsupported (so callers can fall through to the next source).
func NormalizeLanguage(code string) string {
	code = strings.ToLower(strings.TrimSpace(code))
	for _, c := range SupportedLanguages {
		if c == code {
			return c
		}
	}
	return ""
}

// CurrentLanguage, when wired by the server, returns the active user-facing
// output language picked at runtime (settings table) layered over the bootstrap
// default. It is the single source backend display strings (enum labels, etc.)
// read through ActiveLanguage. nil (unit tests / standalone) → ActiveLanguage
// falls back to the Argus default language.
var CurrentLanguage func() string

// ActiveLanguage resolves the language for backend display strings. Precedence:
// the runtime hook (settings table, via the server) over an unwired fallback of
// the configured default through that hook; the fallback is used for standalone
// startup and tests that do not wire the runtime settings hook.
func ActiveLanguage() string {
	if CurrentLanguage != nil {
		if v := NormalizeLanguage(CurrentLanguage()); v != "" {
			return v
		}
	}
	return DefaultLanguage
}

// Language resolves the default user-facing output language with precedence:
//
//	env ARGUS_LANG  >  config file (language)  >  DefaultLanguage
//
// This is the bootstrap default only. The runtime value a user picks in the UI is
// stored in the settings table and takes precedence over this — the server layers
// the two and wires the result into the agent and message catalogs.
func Language() string {
	if v := NormalizeLanguage(os.Getenv("ARGUS_LANG")); v != "" {
		return v
	}
	if v := NormalizeLanguage(Load().Language); v != "" {
		return v
	}
	return DefaultLanguage
}

// BaseDir is the directory that anchors runtime artifacts. Desktop app bundles
// use the per-user application-support directory so installing an app in a
// read-only location such as /Applications does not prevent startup.
//
// When launched via `go run`, the binary is throwaway: it sits either in a temp
// build dir (cache miss → fresh link) OR straight inside the Go build cache
// (cache hit → run from $GOCACHE/.../...-d). We detect both and fall back to the
// current working directory so dev artifacts (data/, transcripts) and config
// resolve against the project dir, not the throwaway binary's location.
func BaseDir() string {
	if dir := strings.TrimSpace(os.Getenv("ARGUS_HOME")); dir != "" {
		return dir
	}
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	dir := filepath.Dir(exe)
	if runtime.GOOS == "windows" {
		if local := os.Getenv("LOCALAPPDATA"); local != "" {
			return filepath.Join(local, "Argus")
		}
		if userDir, err := os.UserConfigDir(); err == nil {
			return filepath.Join(userDir, "Argus")
		}
	}
	if isGoRunDir(dir) {
		return "." // throwaway `go run` binary → use CWD
	}
	if bundleRoot, ok := macOSBundleRoot(exe); ok {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			dir := filepath.Join(home, "Library", "Application Support", "Argus")
			_ = os.MkdirAll(dir, 0o755)
			return dir
		}
		return filepath.Join(bundleRoot, "Contents", "Resources")
	}
	if runtime.GOOS == "linux" {
		if data := os.Getenv("XDG_DATA_HOME"); filepath.IsAbs(data) {
			return filepath.Join(data, "Argus")
		}
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, ".local", "share", "Argus")
		}
	}
	return dir
}

// macOSBundleRoot returns the .app directory when the executable is launched
// from Contents/MacOS inside a macOS application bundle.
func macOSBundleRoot(exe string) (string, bool) {
	macOSDir := filepath.Dir(exe)
	contentsDir := filepath.Dir(macOSDir)
	appDir := filepath.Dir(contentsDir)
	if filepath.Base(contentsDir) != "Contents" || filepath.Base(macOSDir) != "MacOS" {
		return "", false
	}
	if !strings.HasSuffix(filepath.Base(appDir), ".app") {
		return "", false
	}
	return appDir, true
}

// isGoRunDir recognizes Go's temporary build directory and build cache.
// A release extracted into the system temporary directory is still a release.
func isGoRunDir(dir string) bool {
	for _, seg := range strings.Split(filepath.ToSlash(dir), "/") {
		if seg == "go-build" || strings.HasPrefix(seg, "go-build") && len(seg) > len("go-build") && seg[len("go-build")] >= '0' && seg[len("go-build")] <= '9' {
			return true
		}
	}
	return false
}

// Path honors ARGUS_CONFIG first. Releases use the writable user data directory;
// source development through go run can use config.json in the project directory.
func Path() string {
	if v := strings.TrimSpace(os.Getenv("ARGUS_CONFIG")); v != "" {
		return v
	}
	if runtime.GOOS == "windows" || os.Getenv("ARGUS_HOME") != "" {
		return filepath.Join(BaseDir(), "config.json")
	}
	if exe, err := os.Executable(); err == nil && !isGoRunDir(filepath.Dir(exe)) {
		return filepath.Join(BaseDir(), "config.json")
	}
	var candidates []string
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(cwd, "config.json"))
	}
	candidates = append(candidates, filepath.Join(BaseDir(), "config.json"))
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return candidates[0]
}

// Load reads and parses the config file. A missing/unreadable file yields a zero
// Config (so callers fall back to defaults) rather than an error.
func Load() Config {
	var c Config
	b, err := os.ReadFile(Path())
	if err != nil {
		return c
	}
	_ = json.Unmarshal(b, &c)
	return c
}

// SkillDir returns the skill root directory with precedence:
//
//	env ARGUS_SKILL_DIR > config file > bundled skills > BaseDir()/skills
//
// The directory is created if it does not exist.
func SkillDir() string {
	var d string
	if v := strings.TrimSpace(os.Getenv("ARGUS_SKILL_DIR")); v != "" {
		d = v
	} else if v := strings.TrimSpace(Load().SkillDir); v != "" {
		d = v
	} else if exe, err := os.Executable(); err == nil {
		if bundleRoot, ok := macOSBundleRoot(exe); ok {
			d = filepath.Join(bundleRoot, "Contents", "Resources", "skills")
		} else if info, err := os.Stat(filepath.Join(filepath.Dir(exe), "skills")); err == nil && info.IsDir() {
			d = filepath.Join(filepath.Dir(exe), "skills")
		}
	}
	if d == "" {
		if runtime.GOOS == "windows" {
			if exe, err := os.Executable(); err == nil {
				d = filepath.Join(filepath.Dir(exe), "skills")
			}
		}
	}
	if d == "" {
		d = filepath.Join(BaseDir(), "skills")
	}
	_ = os.MkdirAll(d, 0o755)
	return d
}

// PostgresDSN resolves the connection string with precedence:
//
//	env ARGUS_PG_DSN  >  config file (database.dsn, or assembled from fields)
//
// There is NO built-in fallback: when neither source supplies a database config,
// it returns an error naming the config path it inspected, so startup fails loudly
// instead of silently connecting to a wrong default. source describes where the
// DSN came from (for startup logging).
func PostgresDSN() (dsn, source string, err error) {
	if v := firstNonEmptyEnv("ARGUS_PG_DSN", "ARGUS_PG_DSN"); v != "" {
		return v, trCfg("环境变量 ARGUS_PG_DSN"), nil
	}
	db := Load().Database
	if d := strings.TrimSpace(db.DSN); d != "" {
		return d, fmt.Sprintf(trCfg("配置文件 %s (database.dsn)"), Path()), nil
	}
	if db.Host != "" || db.DBName != "" || db.User != "" {
		return db.buildDSN(), fmt.Sprintf(trCfg("配置文件 %s (database 字段)"), Path()), nil
	}
	return "", "", fmt.Errorf(trCfg("找不到数据库设置: 环境变量 ARGUS_PG_DSN 未设置, 配置文件 %s 江户 database (dsn 或 host/user/dbname) 没有设置. 请创建配置文件或设置环境变量并重试"), Path())
}

// HasPostgresConfiguration reports whether the user explicitly configured an
// external PostgreSQL instance. An empty result means the desktop runtime
// should use its embedded, persistent PostgreSQL instance instead.
func HasPostgresConfiguration() bool {
	if firstNonEmptyEnv("ARGUS_PG_DSN", "ARGUS_PG_DSN") != "" {
		return true
	}
	db := Load().Database
	return strings.TrimSpace(db.DSN) != "" || db.Host != "" || db.DBName != "" || db.User != ""
}

func firstNonEmptyEnv(names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(os.Getenv(name)); value != "" {
			return value
		}
	}
	return ""
}

func (d Database) buildDSN() string {
	host := d.Host
	if host == "" {
		host = "127.0.0.1"
	}
	port := d.Port
	if port == 0 {
		port = 5432
	}
	ssl := d.SSLMode
	if ssl == "" {
		ssl = "disable"
	}
	u := url.URL{
		Scheme: "postgres",
		Host:   host + ":" + strconv.Itoa(port),
		Path:   "/" + d.DBName,
	}
	if d.User != "" {
		if d.Password != "" {
			u.User = url.UserPassword(d.User, d.Password)
		} else {
			u.User = url.User(d.User)
		}
	}
	u.RawQuery = url.Values{"sslmode": {ssl}}.Encode()
	return u.String()
}

// String is a redacted view of the resolved DSN (password masked) for logging.
func Redact(dsn string) string {
	u, err := url.Parse(dsn)
	if err != nil {
		return dsn
	}
	if u.User != nil {
		if _, hasPw := u.User.Password(); hasPw {
			u.User = url.UserPassword(u.User.Username(), "****")
		}
	}
	return fmt.Sprintf("%s", u.String())
}
