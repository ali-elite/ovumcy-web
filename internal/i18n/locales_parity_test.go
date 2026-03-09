package i18n

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

func TestLocaleKeysParity(t *testing.T) {
	locales := mustLoadAllLocaleMessages(t)
	reference, ok := locales[LangEN]
	if !ok {
		t.Fatalf("reference locale %q is missing", LangEN)
	}

	languages := make([]string, 0, len(locales))
	for language := range locales {
		languages = append(languages, language)
	}
	sort.Strings(languages)

	for _, language := range languages {
		if language == LangEN {
			continue
		}

		missing := missingKeys(reference, locales[language])
		extra := missingKeys(locales[language], reference)
		if len(missing) == 0 && len(extra) == 0 {
			continue
		}
		if len(missing) > 0 {
			t.Errorf("keys missing in %s locale: %s", language, strings.Join(missing, ", "))
		}
		if len(extra) > 0 {
			t.Errorf("unexpected keys in %s locale: %s", language, strings.Join(extra, ", "))
		}
	}
}

func mustLoadAllLocaleMessages(t *testing.T) map[string]map[string]string {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("resolve test file path: runtime.Caller failed")
	}
	localesDir := filepath.Join(filepath.Dir(thisFile), "locales")
	entries, err := os.ReadDir(localesDir)
	if err != nil {
		t.Fatalf("read locales dir: %v", err)
	}

	locales := make(map[string]map[string]string)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		language := strings.TrimSuffix(strings.ToLower(entry.Name()), filepath.Ext(entry.Name()))
		localePath := filepath.Join(localesDir, entry.Name())

		content, err := os.ReadFile(localePath)
		if err != nil {
			t.Fatalf("read locale %q: %v", language, err)
		}

		messages := map[string]string{}
		if err := json.Unmarshal(content, &messages); err != nil {
			t.Fatalf("parse locale %q: %v", language, err)
		}
		if len(messages) == 0 {
			t.Fatalf("locale %q is empty", language)
		}
		locales[language] = messages
	}

	if len(locales) == 0 {
		t.Fatal("expected at least one locale")
	}

	return locales
}

func missingKeys(source map[string]string, target map[string]string) []string {
	missing := make([]string, 0)
	for key := range source {
		if _, ok := target[key]; !ok {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	return missing
}
