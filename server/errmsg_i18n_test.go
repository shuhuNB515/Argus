package server

import (
	"strings"
	"testing"
	"unicode"

	"github.com/shuhuNB515/Argus/config"
)

func hasHanText(s string) bool {
	return strings.IndexFunc(s, func(r rune) bool { return unicode.Is(unicode.Han, r) }) >= 0
}

func TestTrMsgSwitchesBetweenChineseAndEnglish(t *testing.T) {
	t.Cleanup(func() { config.CurrentLanguage = nil })
	const source = errAssetIDRequired
	if !hasHanText(source) {
		t.Fatalf("catalog source should be Chinese: %q", source)
	}

	config.CurrentLanguage = func() string { return "zh" }
	if got := trMsg(source); got != source {
		t.Fatalf("Chinese message changed: %q", got)
	}

	config.CurrentLanguage = func() string { return "en" }
	if got := trMsg(source); got == source || got == "" || hasHanText(got) {
		t.Fatalf("English message was not translated: %q", got)
	}

	config.CurrentLanguage = func() string { return "fr" }
	if got := trMsg(source); got != source {
		t.Fatalf("unsupported language must fall back to Chinese: %q", got)
	}
}
