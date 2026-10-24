package srcagentpolicy

import "testing"

func TestModeContextFailsClosedOutsideSRCRange(t *testing.T) {
	c := ModeContext{Mode: ModeSRC, Target: "https://example.com", TestRange: []string{"example.com"}}
	if !c.AllowsURL("https://example.com/login") {
		t.Fatal("expected target URL to be authorized")
	}
	if c.AllowsURL("https://outside.example/login") {
		t.Fatal("expected URL outside test range to be denied")
	}
}

func TestCTFRequiresExplicitTargetAndAnalysisHasNoNetworkTarget(t *testing.T) {
	if (ModeContext{Mode: ModeCTF}).AllowsURL("https://challenge.local") {
		t.Fatal("unbound CTF context must deny network access")
	}
	if (ModeContext{Mode: ModeAnalysis, Target: "example.com"}).AllowsURL("https://example.com") {
		t.Fatal("analysis mode must deny network access")
	}
}

func TestSRCRequiresHTTPPortsAndExplicitSubdomains(t *testing.T) {
	c := ModeContext{Mode: ModeSRC, Target: "shu560.cn", TestRange: []string{"shu560.cn"}}
	for _, raw := range []string{"https://shu560.cn/", "http://shu560.cn:80/", "https://shu560.cn:443/"} {
		if !c.AllowsURL(raw) {
			t.Errorf("expected allowed: %s", raw)
		}
	}
	for _, raw := range []string{"https://www.shu560.cn/", "https://shu560.cn:8443/", "file:///etc/passwd"} {
		if c.AllowsURL(raw) {
			t.Errorf("expected denied: %s", raw)
		}
	}
	wild := ModeContext{Mode: ModeSRC, Target: "shu560.cn", TestRange: []string{"*.shu560.cn"}}
	if !wild.AllowsURL("https://www.shu560.cn/") {
		t.Fatal("explicit wildcard should allow subdomain")
	}
}

func TestTargetContextParsesExplicitRangeAndFailsClosedForBareHosts(t *testing.T) {
	ctx, err := ParseTargetContext("Target: https://shu560.cn/\nTest Range: shu560.cn\nPorts: 80,443")
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Target != "https://shu560.cn/" || len(ctx.TestRange) != 1 || len(ctx.Ports) != 2 {
		t.Fatalf("unexpected context: %+v", ctx)
	}
	if err := ctx.AuthorizeURL("https://shu560.cn/robots.txt"); err != nil {
		t.Fatalf("target URL denied: %v", err)
	}
	if err := ctx.AuthorizeHost("www.shu560.cn", 443); err == nil {
		t.Fatal("implicit subdomain was allowed")
	}
	if err := ctx.AuthorizeHost("shu560.cn", 8443); err == nil {
		t.Fatal("non-default port was allowed")
	}
}

func TestTargetContextParsesPlainURLFromChatMessage(t *testing.T) {
	ctx, err := ParseTargetContext("https://example.com/login 对这个网站进行安全测试")
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Target != "https://example.com/login" {
		t.Fatalf("unexpected target: %q", ctx.Target)
	}
	if len(ctx.TestRange) != 1 || ctx.TestRange[0] != "example.com" {
		t.Fatalf("unexpected test range: %+v", ctx.TestRange)
	}
	if len(ctx.Ports) != 2 || ctx.Ports[0] != 80 || ctx.Ports[1] != 443 {
		t.Fatalf("unexpected default ports: %+v", ctx.Ports)
	}

	ctx, err = ParseTargetContext("https://example.com/对这个网站进行安全测试")
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Target != "https://example.com/对这个网站进行安全测试" || ctx.TestRange[0] != "example.com" {
		t.Fatalf("unexpected URL-plus-Chinese-message context: %+v", ctx)
	}
}

func TestTargetContextRejectsMultipleUnstructuredURLs(t *testing.T) {
	if _, err := ParseTargetContext("https://example.com and https://other.example.com"); err == nil {
		t.Fatal("multiple unstructured URLs should require an explicit Target")
	}
}

func TestTargetContextExplicitWildcardAllowsOnlyMatchingSubdomains(t *testing.T) {
	ctx := TargetContext{Mode: ModeSRC, Target: "https://example.com/", TestRange: []string{"*.example.com"}, Ports: []int{80, 443}}
	if err := ctx.AuthorizeHost("api.example.com", 443); err != nil {
		t.Fatalf("explicit wildcard denied: %v", err)
	}
	if err := ctx.AuthorizeHost("example.com", 443); err != nil {
		t.Fatalf("wildcard should retain apex authorization: %v", err)
	}
	if err := ctx.AuthorizeHost("api.other.com", 443); err == nil {
		t.Fatal("unrelated host was allowed")
	}
}
