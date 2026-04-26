package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestLanguageSwitchSetsCookieAndRendersLocalizedLogin(t *testing.T) {
	tests := []struct {
		name               string
		switchLanguage     string
		expectedCookie     string
		expectedHTMLLang   string
		expectedTitle      string
		expectedHelperText string
	}{
		{
			name:               "english",
			switchLanguage:     "en",
			expectedCookie:     "en",
			expectedHTMLLang:   "en",
			expectedTitle:      "Stay signed in for 30 days",
			expectedHelperText: "only until you close the browser",
		},
		{
			name:               "russian",
			switchLanguage:     "ru",
			expectedCookie:     "ru",
			expectedHTMLLang:   "ru",
			expectedTitle:      "Оставаться в системе 30 дней",
			expectedHelperText: "только до закрытия браузера",
		},
		{
			name:               "spanish",
			switchLanguage:     "es",
			expectedCookie:     "es",
			expectedHTMLLang:   "es",
			expectedTitle:      "Mantener la sesión iniciada durante 30 días",
			expectedHelperText: "solo hasta que cierres el navegador",
		},
		{
			name:               "german",
			switchLanguage:     "de",
			expectedCookie:     "de",
			expectedHTMLLang:   "de",
			expectedTitle:      "30 Tage angemeldet bleiben",
			expectedHelperText: "bis Sie den Browser schließen",
		},
		{
			name:               "french",
			switchLanguage:     "fr",
			expectedCookie:     "fr",
			expectedHTMLLang:   "fr",
			expectedTitle:      "Rester connecté(e) pendant 30 jours",
			expectedHelperText: "fermeture du navigateur",
		},
		{
			name:               "persian",
			switchLanguage:     "fa",
			expectedCookie:     "fa",
			expectedHTMLLang:   "fa",
			expectedTitle:      "مرا تا ۳۰ روز در وضعیت ورود نگه دار",
			expectedHelperText: "زمان بستن مرورگر",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			app, _ := newOnboardingTestApp(t)

			switchForm := url.Values{
				"lang": {testCase.switchLanguage},
				"next": {"/login"},
			}
			switchRequest := httptest.NewRequest(http.MethodPost, "/lang", strings.NewReader(switchForm.Encode()))
			switchRequest.Header.Set("Accept-Language", "en")
			switchRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			switchResponse, err := app.Test(switchRequest, -1)
			if err != nil {
				t.Fatalf("switch language request failed: %v", err)
			}
			defer switchResponse.Body.Close()

			if switchResponse.StatusCode != http.StatusSeeOther {
				t.Fatalf("expected status 303, got %d", switchResponse.StatusCode)
			}
			if location := switchResponse.Header.Get("Location"); location != "/login" {
				t.Fatalf("expected redirect to /login, got %q", location)
			}

			languageCookie := responseCookieValue(switchResponse.Cookies(), "ovumcy_lang")
			if languageCookie != testCase.expectedCookie {
				t.Fatalf("expected ovumcy_lang cookie value %q, got %q", testCase.expectedCookie, languageCookie)
			}

			loginRequest := httptest.NewRequest(http.MethodGet, "/login", nil)
			loginRequest.Header.Set("Cookie", "ovumcy_lang="+languageCookie)
			loginResponse, err := app.Test(loginRequest, -1)
			if err != nil {
				t.Fatalf("localized login request failed: %v", err)
			}
			defer loginResponse.Body.Close()

			loginBody, err := io.ReadAll(loginResponse.Body)
			if err != nil {
				t.Fatalf("read localized login body: %v", err)
			}
			rendered := string(loginBody)
			if !strings.Contains(rendered, `<html lang="`+testCase.expectedHTMLLang+`"`) {
				t.Fatalf("expected login page html lang to be %q", testCase.expectedHTMLLang)
			}
			if testCase.expectedHTMLLang == "fa" && !strings.Contains(rendered, `dir="rtl"`) {
				t.Fatalf("expected persian login page to render with rtl direction")
			}
			if !strings.Contains(rendered, testCase.expectedTitle) {
				t.Fatalf("expected localized remember-me control on login form, got %q", rendered)
			}
			if !strings.Contains(rendered, testCase.expectedHelperText) {
				t.Fatalf("expected localized remember-me helper text on login form, got %q", rendered)
			}
		})
	}
}

func TestLoginPageRendersVisibleLanguageSwitchForm(t *testing.T) {
	app, _ := newOnboardingTestApp(t)

	request := httptest.NewRequest(http.MethodGet, "/login", nil)
	request.Header.Set("Accept-Language", "ru")

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("login page request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected login page status 200, got %d", response.StatusCode)
	}

	rendered := mustReadBodyString(t, response.Body)
	assertBodyContainsAll(t, rendered,
		bodyStringMatch{fragment: `action="/lang"`, message: "expected visible public language switch form"},
		bodyStringMatch{fragment: `data-language-switch-form`, message: "expected public language switch hook"},
		bodyStringMatch{fragment: `data-language-switch-option="ru"`, message: "expected russian language option"},
		bodyStringMatch{fragment: `name="next" value="/login"`, message: "expected language switch to preserve the current auth path"},
		bodyStringMatch{fragment: `class="lang-link lang-link-active"`, message: "expected current language option to be visibly active"},
	)
}
