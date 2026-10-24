// Package srcagentpolicy adapts ARTEX task execution to SRC Agent's
// mode-specific authorization contract.
package srcagentpolicy

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type Mode string

const (
	ModeSRC      Mode = "MODE_SRC"
	ModeCTF      Mode = "MODE_CTF"
	ModeLocalLab Mode = "MODE_LOCAL_LAB"
	ModeAnalysis Mode = "MODE_ANALYSIS"
)

// ModeContext is attached to an ARTEX run. TestRange is meaningful only in
// SRC mode; other modes never inherit SRC authorization.
type ModeContext struct {
	Mode      Mode
	Target    string
	TestRange []string
	LocalRoot string
	Ports     []int
}

// TargetContext is the persisted, canonical authorization contract for one SRC
// run. The executor receives this exact value; it never derives authorization
// from a task title or from an LLM message at call time.
type TargetContext = ModeContext

var (
	targetLine = regexp.MustCompile(`(?im)^\s*target\s*[:=]\s*(\S+)`)
	rangeLine  = regexp.MustCompile(`(?im)^\s*(?:test\s*range|range)\s*[:=]\s*(.+)$`)
	portsLine  = regexp.MustCompile(`(?im)^\s*ports?\s*[:=]\s*(.+)$`)
)

// ParseTargetContext accepts the human-readable scope form used by the SRC
// conversation UI, for example:
//
//	Target: https://example.com/
//	Test Range: example.com, *.api.example.com
//	Ports: 80,443
func ParseTargetContext(text string) (TargetContext, error) {
	targetMatch := targetLine.FindStringSubmatch(text)
	target := ""
	if len(targetMatch) == 2 {
		target = strings.TrimSpace(targetMatch[1])
	} else {
		// Chat messages commonly start with a URL followed by a natural-language
		// request. Treat a single URL literal as the target, while keeping the
		// canonical range and port defaults below. Structured Target/Test Range
		// input remains supported and takes precedence when present.
		literals := urlLiteral.FindAllString(text, -1)
		if len(literals) == 0 {
			return TargetContext{}, fmt.Errorf("SRC Target is required")
		}
		if len(literals) > 1 {
			return TargetContext{}, fmt.Errorf("SRC Target must be explicit when multiple URLs are provided")
		}
		target = strings.TrimRight(strings.TrimSpace(literals[0]), ".,;)]}\"'")
	}
	u, err := url.Parse(target)
	if err != nil || u.Hostname() == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return TargetContext{}, fmt.Errorf("invalid SRC Target")
	}
	ranges := []string{strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))}
	if match := rangeLine.FindStringSubmatch(text); len(match) == 2 {
		ranges = nil
		for _, item := range strings.FieldsFunc(match[1], func(r rune) bool { return r == ',' || r == ';' || r == '\n' }) {
			item = strings.TrimSpace(item)
			if item != "" {
				ranges = append(ranges, strings.ToLower(strings.TrimSuffix(item, ".")))
			}
		}
	}
	if len(ranges) == 0 {
		return TargetContext{}, fmt.Errorf("SRC Test Range is required")
	}
	ports := []int{80, 443}
	if match := portsLine.FindStringSubmatch(text); len(match) == 2 {
		ports = nil
		seen := map[int]bool{}
		for _, item := range strings.FieldsFunc(match[1], func(r rune) bool { return r == ',' || r == ';' || r == ' ' || r == '\n' }) {
			p, parseErr := strconv.Atoi(strings.TrimSpace(item))
			if parseErr != nil || p < 1 || p > 65535 || seen[p] {
				continue
			}
			seen[p] = true
			ports = append(ports, p)
		}
		if len(ports) == 0 {
			return TargetContext{}, fmt.Errorf("SRC Ports must contain a valid port")
		}
	}
	return TargetContext{Mode: ModeSRC, Target: target, TestRange: ranges, Ports: ports}, nil
}

func (c ModeContext) AllowsURL(raw string) bool {
	return c.AuthorizeURL(raw) == nil
}

func (c ModeContext) AuthorizeURL(raw string) error {
	if raw == "" || c.Mode == ModeAnalysis {
		return fmt.Errorf("network access denied by SRC policy")
	}
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" {
		return fmt.Errorf("invalid network URL")
	}
	if c.Mode == ModeSRC && u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("protocol %q is outside the authorized SRC policy", u.Scheme)
	}
	port := 0
	if u.Port() != "" {
		port, err = strconv.Atoi(u.Port())
		if err != nil {
			return fmt.Errorf("invalid network port")
		}
	} else if u.Scheme == "http" {
		port = 80
	} else if u.Scheme == "https" {
		port = 443
	}
	host := strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))
	switch c.Mode {
	case ModeSRC:
		if c.Target == "" || len(c.TestRange) == 0 {
			return fmt.Errorf("SRC TargetContext is not configured")
		}
		for _, allowed := range c.TestRange {
			if hostMatches(host, allowed) {
				if c.portAllowed(port) {
					return nil
				}
				return fmt.Errorf("port %d is outside the authorized SRC policy", port)
			}
		}
		return fmt.Errorf("host %q is outside the authorized SRC Test Range", host)
	case ModeCTF:
		if c.Target != "" && hostMatches(host, c.Target) {
			return nil
		}
		return fmt.Errorf("host %q is outside the authorized CTF target", host)
	case ModeLocalLab:
		if c.LocalRoot != "" && strings.HasPrefix(raw, c.LocalRoot) {
			return nil
		}
		return fmt.Errorf("URL is outside the local lab root")
	default:
		return fmt.Errorf("network access is not enabled for this mode")
	}
}

func (c ModeContext) AuthorizeHost(rawHost string, port int) error {
	host := strings.TrimSpace(strings.ToLower(rawHost))
	if h, p, err := net.SplitHostPort(host); err == nil {
		host = h
		if port == 0 {
			port, _ = strconv.Atoi(p)
		}
	}
	host = strings.TrimSuffix(host, ".")
	if host == "" {
		return fmt.Errorf("empty network host")
	}
	if c.Mode != ModeSRC {
		if c.Mode == ModeCTF && c.Target != "" && hostMatches(host, c.Target) {
			return nil
		}
		return fmt.Errorf("host %q is not authorized", host)
	}
	for _, allowed := range c.TestRange {
		if hostMatches(host, allowed) {
			if c.portAllowed(port) {
				return nil
			}
			return fmt.Errorf("port %d is outside the authorized SRC policy", port)
		}
	}
	return fmt.Errorf("host %q is outside the authorized SRC Test Range", host)
}

func (c ModeContext) portAllowed(port int) bool {
	if c.Mode != ModeSRC {
		return c.Mode != ModeSRC
	}
	if port == 0 {
		return true // a bare host does not select a port; the executor defaults to 80/443
	}
	ports := c.Ports
	if len(ports) == 0 {
		ports = []int{80, 443}
	}
	for _, allowed := range ports {
		if port == allowed {
			return true
		}
	}
	return false
}

// CheckInput is the single pre-execution check for network-capable tool input.
// It checks URL literals everywhere, and host/domain fields for tools such as
// traffic_search that intentionally accept a bare host.
func (c ModeContext) CheckInput(value any) error {
	var walk func(any, string) error
	walk = func(v any, key string) error {
		switch x := v.(type) {
		case string:
			for _, literal := range urlLiteral.FindAllString(x, -1) {
				u := strings.TrimRight(literal, ".,;)]}\"'")
				if err := c.AuthorizeURL(u); err != nil {
					return fmt.Errorf("network target rejected: %s", sanitizeErrorTarget(u))
				}
			}
			if isHostKey(key) && !strings.Contains(x, "://") {
				if err := c.AuthorizeHost(x, 0); err != nil {
					return fmt.Errorf("network target rejected: %s", sanitizeErrorTarget(x))
				}
			}
		}
		switch x := v.(type) {
		case []any:
			for _, item := range x {
				if err := walk(item, key); err != nil {
					return err
				}
			}
		case map[string]any:
			for k, item := range x {
				if err := walk(item, strings.ToLower(k)); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return walk(value, "")
}

func isHostKey(key string) bool {
	switch strings.ToLower(key) {
	case "host", "hostname", "domain", "target", "endpoint", "base_url", "baseurl":
		return true
	default:
		return false
	}
}

func sanitizeErrorTarget(raw string) string {
	if u, err := url.Parse(raw); err == nil && u.Hostname() != "" {
		return u.Scheme + "://" + strings.ToLower(u.Hostname()) + ":" + u.Port()
	}
	return strings.TrimSpace(raw)
}

var urlLiteral = regexp.MustCompile(`https?://[^\s"'<>]+`)

func hostMatches(host, pattern string) bool {
	p := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(pattern), "/"))
	if parsed, err := url.Parse(p); err == nil && parsed.Hostname() != "" {
		p = parsed.Hostname()
	}
	wildcard := strings.HasPrefix(p, "*.")
	p = strings.TrimPrefix(p, "*.")
	if wildcard {
		return host == p || strings.HasSuffix(host, "."+p)
	}
	return host == p
}
