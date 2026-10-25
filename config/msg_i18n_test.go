package config

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unicode"
)

func containsHan(s string) bool {
	return strings.IndexFunc(s, func(r rune) bool { return unicode.Is(unicode.Han, r) }) >= 0
}

func TestSupportedLanguagesAreChineseAndEnglish(t *testing.T) {
	want := []string{"zh", "en"}
	if strings.Join(SupportedLanguages, ",") != strings.Join(want, ",") {
		t.Fatalf("SupportedLanguages=%v, want %v", SupportedLanguages, want)
	}
	if DefaultLanguage != "zh" {
		t.Fatalf("DefaultLanguage=%q, want zh", DefaultLanguage)
	}
	for _, unsupported := range []string{"ko", "es", "ja", "fr"} {
		if got := NormalizeLanguage(unsupported); got != "" {
			t.Errorf("NormalizeLanguage(%q)=%q, want empty", unsupported, got)
		}
	}
}

func TestStartupMessageSwitchesBetweenChineseAndEnglish(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ARGUS_PG_DSN", "")
	t.Setenv("ARGUS_CONFIG", filepath.Join(dir, "missing.json"))

	t.Setenv("ARGUS_LANG", "zh")
	zh := mustDSNErr(t)
	if !containsHan(zh) {
		t.Fatalf("Chinese startup error has no Chinese text: %q", zh)
	}

	t.Setenv("ARGUS_LANG", "en")
	en := mustDSNErr(t)
	if containsHan(en) || en == zh {
		t.Fatalf("English startup error was not translated: %q", en)
	}
	for _, id := range []string{"ARGUS_PG_DSN", "database", "dsn", "host/user/dbname"} {
		if !strings.Contains(en, id) {
			t.Errorf("English startup error must preserve %q: %q", id, en)
		}
	}
}

func mustDSNErr(t *testing.T) string {
	t.Helper()
	_, _, err := PostgresDSN()
	if err == nil {
		t.Fatal("missing config should return an error")
	}
	return err.Error()
}

func TestCatalogFormatVerbsMatch(t *testing.T) {
	verbs := regexp.MustCompile(`%[-+# 0-9.]*[a-zA-Z%]`)
	for name, cat := range map[string]map[string]map[string]string{"cfgCatalog": cfgCatalog, "logCatalog": logCatalog} {
		for source, translations := range cat {
			want := strings.Join(verbs.FindAllString(source, -1), " ")
			for lang, value := range translations {
				if lang != "zh" && lang != "en" {
					t.Errorf("%s[%q] contains unsupported language %q", name, source, lang)
				}
				if got := strings.Join(verbs.FindAllString(value, -1), " "); got != want {
					t.Errorf("%s[%q][%s] verbs=%q, want %q", name, source, lang, got, want)
				}
			}
		}
	}
}
