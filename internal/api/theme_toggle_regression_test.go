package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func TestBaseTemplateOmitsThemeToggleControl(t *testing.T) {
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

	if strings.Contains(rendered, `class="theme-toggle"`) {
		t.Fatalf("did not expect theme toggle control in base layout")
	}
	if strings.Contains(rendered, `data-theme-toggle`) {
		t.Fatalf("did not expect theme toggle mount point in base layout")
	}
}

func TestBaseTemplateUsesExternalCSPCompatibleScripts(t *testing.T) {
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

	if !strings.Contains(rendered, `<script src="/static/js/theme-bootstrap.js?v=`) {
		t.Fatalf("expected base template to include external theme bootstrap script")
	}
	if !strings.Contains(rendered, `<script defer src="/static/js/app.js?v=`) {
		t.Fatalf("expected base template to include external shared app script")
	}
	if strings.Contains(rendered, "alpine.min.js") {
		t.Fatalf("did not expect alpine script after CSP preparation")
	}

	scriptTagPattern := regexp.MustCompile(`(?is)<script\b[^>]*>`)
	scriptTags := scriptTagPattern.FindAllString(rendered, -1)
	for _, tag := range scriptTags {
		if !strings.Contains(strings.ToLower(tag), "src=") {
			t.Fatalf("expected base template to avoid inline scripts, found tag %q", tag)
		}
	}
}
