package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/ovumcy/internal/models"
	"gorm.io/gorm"
)

func TestSettingsSymptomsHTMXCreateArchiveRestoreRerendersSection(t *testing.T) {
	ctx := newSettingsSymptomsHTMXTestContext(t, "settings-symptoms-htmx@example.com")

	createForm := url.Values{
		"csrf_token": {ctx.csrfToken},
		"name":       {"Joint stiffness"},
		"icon":       {"J"},
	}
	renderedCreate := performSettingsSymptomsHTMXRequest(t, ctx, http.MethodPost, "/api/symptoms", createForm)
	assertSettingsSymptomsHTMXContains(t, renderedCreate, `id="settings-symptoms-section"`, "symptom section rerender")
	assertSettingsSymptomsHTMXContains(t, renderedCreate, `data-symptom-name="Joint stiffness"`, "new custom symptom row")

	stored := models.SymptomType{}
	if err := ctx.database.Where("user_id = ? AND name = ?", ctx.user.ID, "Joint stiffness").First(&stored).Error; err != nil {
		t.Fatalf("load created custom symptom: %v", err)
	}
	if stored.Color != "#E8799F" {
		t.Fatalf("expected default symptom color, got %q", stored.Color)
	}

	archiveForm := url.Values{"csrf_token": {ctx.csrfToken}}
	renderedArchive := performSettingsSymptomsHTMXRequest(t, ctx, http.MethodPost, "/api/symptoms/"+strconv.FormatUint(uint64(stored.ID), 10)+"/archive", archiveForm)
	assertSettingsSymptomsHTMXContains(t, renderedArchive, `data-symptom-state="archived"`, "archived custom symptom row")

	restoreForm := url.Values{"csrf_token": {ctx.csrfToken}}
	renderedRestore := performSettingsSymptomsHTMXRequest(t, ctx, http.MethodPost, "/api/symptoms/"+strconv.FormatUint(uint64(stored.ID), 10)+"/restore", restoreForm)
	assertSettingsSymptomsHTMXContains(t, renderedRestore, `data-symptom-state="active"`, "active custom symptom row after restore")
}

func TestSettingsSymptomsHTMXUpdateDuplicateShowsRowLocalError(t *testing.T) {
	app, database := newOnboardingTestAppWithCSRF(t)
	user := createOnboardingTestUser(t, database, "settings-symptoms-htmx-duplicate@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookieWithCSRF(t, app, user.Email, "StrongPass1")
	csrfCookie, csrfToken := loadSettingsCSRFContext(t, app, authCookie)

	active := models.SymptomType{
		UserID: user.ID,
		Name:   "Joint stiffness",
		Icon:   "✨",
		Color:  "#334455",
	}
	if err := database.Create(&active).Error; err != nil {
		t.Fatalf("create active symptom: %v", err)
	}

	archivedAt := time.Now().UTC()
	archived := models.SymptomType{
		UserID:     user.ID,
		Name:       "Joint support",
		Icon:       "🔥",
		Color:      "#14B8A6",
		ArchivedAt: &archivedAt,
	}
	if err := database.Create(&archived).Error; err != nil {
		t.Fatalf("create archived symptom: %v", err)
	}

	updateForm := url.Values{
		"csrf_token": {csrfToken},
		"name":       {"Joint stiffness"},
		"icon":       {"🔥"},
	}
	updateRequest := httptest.NewRequest(http.MethodPost, "/api/symptoms/"+strconv.FormatUint(uint64(archived.ID), 10), strings.NewReader(updateForm.Encode()))
	updateRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	updateRequest.Header.Set("HX-Request", "true")
	updateRequest.Header.Set("Cookie", joinCookieHeader(authCookie, cookiePair(csrfCookie)))

	updateResponse, err := app.Test(updateRequest, -1)
	if err != nil {
		t.Fatalf("update duplicate symptom htmx request failed: %v", err)
	}
	defer updateResponse.Body.Close()

	if updateResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected htmx update status 200, got %d", updateResponse.StatusCode)
	}
	updateBody, err := io.ReadAll(updateResponse.Body)
	if err != nil {
		t.Fatalf("read htmx update body: %v", err)
	}
	renderedUpdate := string(updateBody)
	if !strings.Contains(renderedUpdate, `data-symptom-id="`+strconv.FormatUint(uint64(archived.ID), 10)+`"`) {
		t.Fatalf("expected archived symptom row to be present after duplicate update, got %q", renderedUpdate)
	}
	if !strings.Contains(renderedUpdate, `data-symptom-row-error`) {
		t.Fatalf("expected row-local error block in htmx response, got %q", renderedUpdate)
	}
	if !strings.Contains(renderedUpdate, `That symptom name already exists in your list.`) {
		t.Fatalf("expected localized duplicate-name error, got %q", renderedUpdate)
	}
	if !strings.Contains(renderedUpdate, `id="settings-symptom-name-`+strconv.FormatUint(uint64(archived.ID), 10)+`"`) ||
		!strings.Contains(renderedUpdate, `value="Joint stiffness"`) {
		t.Fatalf("expected duplicate draft name to remain in the archived row, got %q", renderedUpdate)
	}
}

func TestSettingsSymptomsHTMXCreateTooLongClearsDraftName(t *testing.T) {
	app, database := newOnboardingTestAppWithCSRF(t)
	user := createOnboardingTestUser(t, database, "settings-symptoms-htmx-too-long@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookieWithCSRF(t, app, user.Email, "StrongPass1")
	csrfCookie, csrfToken := loadSettingsCSRFContext(t, app, authCookie)

	createForm := url.Values{
		"csrf_token": {csrfToken},
		"name":       {"12345678901234567890123456789012345678901"},
		"icon":       {"✨"},
	}
	createRequest := httptest.NewRequest(http.MethodPost, "/api/symptoms", strings.NewReader(createForm.Encode()))
	createRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	createRequest.Header.Set("HX-Request", "true")
	createRequest.Header.Set("Cookie", joinCookieHeader(authCookie, cookiePair(csrfCookie)))

	createResponse, err := app.Test(createRequest, -1)
	if err != nil {
		t.Fatalf("create too-long symptom htmx request failed: %v", err)
	}
	defer createResponse.Body.Close()

	if createResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected htmx create status 200, got %d", createResponse.StatusCode)
	}
	createBody, err := io.ReadAll(createResponse.Body)
	if err != nil {
		t.Fatalf("read htmx create body: %v", err)
	}
	renderedCreate := string(createBody)
	if !strings.Contains(renderedCreate, `Use 40 characters or fewer. For longer details, use notes.`) {
		t.Fatalf("expected localized too-long error, got %q", renderedCreate)
	}
	if !strings.Contains(renderedCreate, `id="settings-new-symptom-name"`) {
		t.Fatalf("expected symptom create field in htmx response, got %q", renderedCreate)
	}
	if strings.Contains(renderedCreate, `value="12345678901234567890123456789012345678901"`) {
		t.Fatalf("expected too-long create draft to be cleared, got %q", renderedCreate)
	}
}

func TestSettingsSymptomsHTMXUpdateTooLongRestoresSavedRowValue(t *testing.T) {
	app, database := newOnboardingTestAppWithCSRF(t)
	user := createOnboardingTestUser(t, database, "settings-symptoms-htmx-update-too-long@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookieWithCSRF(t, app, user.Email, "StrongPass1")
	csrfCookie, csrfToken := loadSettingsCSRFContext(t, app, authCookie)

	symptom := models.SymptomType{
		UserID: user.ID,
		Name:   "Joint ease",
		Icon:   "💧",
		Color:  "#38BDF8",
	}
	if err := database.Create(&symptom).Error; err != nil {
		t.Fatalf("create custom symptom: %v", err)
	}

	updateForm := url.Values{
		"csrf_token": {csrfToken},
		"name":       {"12345678901234567890123456789012345678901"},
		"icon":       {"🔥"},
	}
	updateRequest := httptest.NewRequest(http.MethodPost, "/api/symptoms/"+strconv.FormatUint(uint64(symptom.ID), 10), strings.NewReader(updateForm.Encode()))
	updateRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	updateRequest.Header.Set("HX-Request", "true")
	updateRequest.Header.Set("Cookie", joinCookieHeader(authCookie, cookiePair(csrfCookie)))

	updateResponse, err := app.Test(updateRequest, -1)
	if err != nil {
		t.Fatalf("update too-long symptom htmx request failed: %v", err)
	}
	defer updateResponse.Body.Close()

	if updateResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected htmx update status 200, got %d", updateResponse.StatusCode)
	}
	updateBody, err := io.ReadAll(updateResponse.Body)
	if err != nil {
		t.Fatalf("read htmx update body: %v", err)
	}
	renderedUpdate := string(updateBody)
	if !strings.Contains(renderedUpdate, `Use 40 characters or fewer. For longer details, use notes.`) {
		t.Fatalf("expected localized too-long error, got %q", renderedUpdate)
	}
	if strings.Contains(renderedUpdate, `value="12345678901234567890123456789012345678901"`) {
		t.Fatalf("expected too-long edit draft to be discarded, got %q", renderedUpdate)
	}
	if !strings.Contains(renderedUpdate, `id="settings-symptom-name-`+strconv.FormatUint(uint64(symptom.ID), 10)+`"`) ||
		!strings.Contains(renderedUpdate, `value="Joint ease"`) {
		t.Fatalf("expected saved symptom name to remain in row, got %q", renderedUpdate)
	}
}

func TestSettingsSymptomsHTMXUpdateWithoutColorPreservesStoredValue(t *testing.T) {
	app, database := newOnboardingTestAppWithCSRF(t)
	user := createOnboardingTestUser(t, database, "settings-symptoms-htmx-preserve-color@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookieWithCSRF(t, app, user.Email, "StrongPass1")
	csrfCookie, csrfToken := loadSettingsCSRFContext(t, app, authCookie)

	symptom := models.SymptomType{
		UserID: user.ID,
		Name:   "Joint ease",
		Icon:   "💧",
		Color:  "#38BDF8",
	}
	if err := database.Create(&symptom).Error; err != nil {
		t.Fatalf("create custom symptom: %v", err)
	}

	updateForm := url.Values{
		"csrf_token": {csrfToken},
		"name":       {"Joint relief"},
		"icon":       {"🔥"},
	}
	updateRequest := httptest.NewRequest(http.MethodPost, "/api/symptoms/"+strconv.FormatUint(uint64(symptom.ID), 10), strings.NewReader(updateForm.Encode()))
	updateRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	updateRequest.Header.Set("HX-Request", "true")
	updateRequest.Header.Set("Cookie", joinCookieHeader(authCookie, cookiePair(csrfCookie)))

	updateResponse, err := app.Test(updateRequest, -1)
	if err != nil {
		t.Fatalf("update symptom htmx request failed: %v", err)
	}
	defer updateResponse.Body.Close()

	if updateResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected htmx update status 200, got %d", updateResponse.StatusCode)
	}

	stored := models.SymptomType{}
	if err := database.First(&stored, symptom.ID).Error; err != nil {
		t.Fatalf("reload updated custom symptom: %v", err)
	}
	if stored.Name != "Joint relief" {
		t.Fatalf("expected updated name, got %q", stored.Name)
	}
	if stored.Icon != "🔥" {
		t.Fatalf("expected updated icon, got %q", stored.Icon)
	}
	if stored.Color != "#38BDF8" {
		t.Fatalf("expected existing color to be preserved, got %q", stored.Color)
	}
}

type settingsSymptomsHTMXTestContext struct {
	app        *fiber.App
	database   *gorm.DB
	user       models.User
	authCookie string
	csrfCookie *http.Cookie
	csrfToken  string
}

func newSettingsSymptomsHTMXTestContext(t *testing.T, email string) settingsSymptomsHTMXTestContext {
	t.Helper()

	app, database := newOnboardingTestAppWithCSRF(t)
	user := createOnboardingTestUser(t, database, email, "StrongPass1", true)
	authCookie := loginAndExtractAuthCookieWithCSRF(t, app, user.Email, "StrongPass1")
	csrfCookie, csrfToken := loadSettingsCSRFContext(t, app, authCookie)

	return settingsSymptomsHTMXTestContext{
		app:        app,
		database:   database,
		user:       user,
		authCookie: authCookie,
		csrfCookie: csrfCookie,
		csrfToken:  csrfToken,
	}
}

func performSettingsSymptomsHTMXRequest(t *testing.T, ctx settingsSymptomsHTMXTestContext, method string, path string, form url.Values) string {
	t.Helper()

	request := httptest.NewRequest(method, path, strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("HX-Request", "true")
	request.Header.Set("Cookie", joinCookieHeader(ctx.authCookie, cookiePair(ctx.csrfCookie)))

	response, err := ctx.app.Test(request, -1)
	if err != nil {
		t.Fatalf("settings symptoms htmx request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected htmx status 200, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read htmx response body: %v", err)
	}

	return string(body)
}

func assertSettingsSymptomsHTMXContains(t *testing.T, rendered string, substring string, description string) {
	t.Helper()

	if !strings.Contains(rendered, substring) {
		t.Fatalf("expected %s, got %q", description, rendered)
	}
}

func loadSettingsCSRFContext(t *testing.T, app *fiber.App, authCookie string) (*http.Cookie, string) {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, "/settings", nil)
	request.Header.Set("Accept-Language", "en")
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("settings request for csrf context failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected settings status 200, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read settings body for csrf context: %v", err)
	}
	csrfToken := extractCSRFTokenFromHTML(t, string(body))
	csrfCookie := responseCookie(response.Cookies(), "ovumcy_csrf")
	if csrfCookie == nil || strings.TrimSpace(csrfCookie.Value) == "" {
		t.Fatalf("expected csrf cookie in settings response")
	}

	return csrfCookie, csrfToken
}
