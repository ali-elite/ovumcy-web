package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func TestBaseTemplateAvoidsInlineScriptsUnderStrictCSP(t *testing.T) {
	app, _ := newOnboardingTestApp(t)

	request := httptest.NewRequest(http.MethodGet, "/login", nil)
	request.Header.Set("Accept-Language", "en")

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	rendered := string(body)

	scriptTagPattern := regexp.MustCompile(`(?is)<script\b[^>]*>`)
	scriptTags := scriptTagPattern.FindAllString(rendered, -1)
	if len(scriptTags) == 0 {
		t.Fatalf("expected at least one external script tag in base template")
	}
	for _, tag := range scriptTags {
		if !regexp.MustCompile(`(?i)\bsrc=`).MatchString(tag) {
			t.Fatalf("expected base template to avoid inline scripts, found tag %q", tag)
		}
	}
}
