package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ovumcy/ovumcy-web/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func TestRegisterRejectsWeakNumericPassword(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	email := "weak-register@example.com"

	form := url.Values{
		"email":            {email},
		"password":         {"12345678"},
		"confirm_password": {"12345678"},
	}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("register request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.StatusCode)
	}

	errorValue := readAPIError(t, response.Body)
	if errorValue != "weak password" {
		t.Fatalf("expected weak password error, got %q", errorValue)
	}

	var usersCount int64
	if err := database.Model(&models.User{}).Where("email = ?", email).Count(&usersCount).Error; err != nil {
		t.Fatalf("count users: %v", err)
	}
	if usersCount != 0 {
		t.Fatalf("expected user not to be created, found %d records", usersCount)
	}
}

func TestRegisterRejectsPasswordMismatch(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	email := "mismatch-register@example.com"

	form := url.Values{
		"email":            {email},
		"password":         {"StrongPass1"},
		"confirm_password": {"StrongPass2"},
	}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("register mismatch request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.StatusCode)
	}

	errorValue := readAPIError(t, response.Body)
	if errorValue != "password mismatch" {
		t.Fatalf("expected password mismatch error, got %q", errorValue)
	}

	var usersCount int64
	if err := database.Model(&models.User{}).Where("email = ?", email).Count(&usersCount).Error; err != nil {
		t.Fatalf("count users: %v", err)
	}
	if usersCount != 0 {
		t.Fatalf("expected user not to be created, found %d records", usersCount)
	}
}

func TestRegisterRejectsCaseInsensitiveDuplicateEmail(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	existingEmail := "QA-Test2@Ovumcy.Local"

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("StrongPass1"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	existingUser := models.User{
		Email:               existingEmail,
		PasswordHash:        string(passwordHash),
		Role:                models.RoleOwner,
		OnboardingCompleted: true,
		CycleLength:         models.DefaultCycleLength,
		PeriodLength:        models.DefaultPeriodLength,
		AutoPeriodFill:      true,
		CreatedAt:           time.Now().UTC(),
	}
	if err := database.Create(&existingUser).Error; err != nil {
		t.Fatalf("create existing user: %v", err)
	}

	form := url.Values{
		"email":            {"qa-test2@ovumcy.local"},
		"password":         {"StrongPass1"},
		"confirm_password": {"StrongPass1"},
	}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("register duplicate request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.StatusCode)
	}

	errorValue := readAPIError(t, response.Body)
	if errorValue != "invalid input" {
		t.Fatalf("expected generic invalid input error, got %q", errorValue)
	}

	var usersCount int64
	if err := database.Model(&models.User{}).Where("lower(trim(email)) = ?", "qa-test2@ovumcy.local").Count(&usersCount).Error; err != nil {
		t.Fatalf("count normalized users: %v", err)
	}
	if usersCount != 1 {
		t.Fatalf("expected exactly one normalized email record, found %d", usersCount)
	}
}

func TestRegisterRejectsExactDuplicateEmail(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	existingEmail := "qatest2@ovumcy.local"

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("StrongPass1"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	existingUser := models.User{
		Email:               existingEmail,
		PasswordHash:        string(passwordHash),
		Role:                models.RoleOwner,
		OnboardingCompleted: true,
		CycleLength:         models.DefaultCycleLength,
		PeriodLength:        models.DefaultPeriodLength,
		AutoPeriodFill:      true,
		CreatedAt:           time.Now().UTC(),
	}
	if err := database.Create(&existingUser).Error; err != nil {
		t.Fatalf("create existing user: %v", err)
	}

	form := url.Values{
		"email":            {existingEmail},
		"password":         {"StrongPass1"},
		"confirm_password": {"StrongPass1"},
	}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("register exact duplicate request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.StatusCode)
	}

	errorValue := readAPIError(t, response.Body)
	if errorValue != "invalid input" {
		t.Fatalf("expected generic invalid input error, got %q", errorValue)
	}

	var usersCount int64
	if err := database.Model(&models.User{}).Where("email = ?", existingEmail).Count(&usersCount).Error; err != nil {
		t.Fatalf("count exact users: %v", err)
	}
	if usersCount != 1 {
		t.Fatalf("expected exactly one exact email record, found %d", usersCount)
	}
}

func TestRegisterRejectsExactDuplicateEmailHTMLFlow(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	existingEmail := "qatest2@ovumcy.local"

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("StrongPass1"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	existingUser := models.User{
		Email:               existingEmail,
		PasswordHash:        string(passwordHash),
		Role:                models.RoleOwner,
		OnboardingCompleted: true,
		CycleLength:         models.DefaultCycleLength,
		PeriodLength:        models.DefaultPeriodLength,
		AutoPeriodFill:      true,
		CreatedAt:           time.Now().UTC(),
	}
	if err := database.Create(&existingUser).Error; err != nil {
		t.Fatalf("create existing user: %v", err)
	}

	form := url.Values{
		"email":            {existingEmail},
		"password":         {"StrongPass1"},
		"confirm_password": {"StrongPass1"},
	}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("register exact duplicate html request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected status 303, got %d", response.StatusCode)
	}
	location := response.Header.Get("Location")
	if strings.TrimSpace(location) != "/register" {
		t.Fatalf("expected redirect to /register, got %q", location)
	}

	flashValue := responseCookieValue(response.Cookies(), flashCookieName)
	if flashValue == "" {
		t.Fatalf("expected flash cookie in register duplicate-email response")
	}

	followRequest := httptest.NewRequest(http.MethodGet, "/register", nil)
	followRequest.Header.Set("Accept-Language", "en")
	followRequest.Header.Set("Cookie", flashCookieName+"="+flashValue)

	followResponse, err := app.Test(followRequest, -1)
	if err != nil {
		t.Fatalf("follow register request failed: %v", err)
	}
	defer followResponse.Body.Close()

	followBody, err := io.ReadAll(followResponse.Body)
	if err != nil {
		t.Fatalf("read follow register body: %v", err)
	}

	rendered := strings.ToLower(string(followBody))
	if strings.Contains(rendered, "already exists") || strings.Contains(rendered, "already registered") {
		t.Fatalf("expected duplicate register page to avoid account existence phrases")
	}

	var usersCount int64
	if err := database.Model(&models.User{}).Where("lower(trim(email)) = ?", existingEmail).Count(&usersCount).Error; err != nil {
		t.Fatalf("count exact users: %v", err)
	}
	if usersCount != 1 {
		t.Fatalf("expected exactly one normalized email record, found %d", usersCount)
	}
}
