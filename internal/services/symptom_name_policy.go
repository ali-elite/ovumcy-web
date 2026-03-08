package services

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	maxSymptomNameLength = 80
	defaultSymptomIcon   = "✨"
)

var (
	hexSymptomColorPattern = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
)

func normalizeSymptomNameInput(raw string) (string, error) {
	name := normalizeSymptomSpacing(raw)
	if name == "" || utf8.RuneCountInString(name) > maxSymptomNameLength {
		return "", ErrInvalidSymptomName
	}
	if containsInvalidSymptomNameRune(name) {
		return "", ErrInvalidSymptomName
	}
	return name, nil
}

func normalizeSymptomNameKey(raw string) string {
	return strings.ToLower(normalizeSymptomSpacing(raw))
}

func normalizeSymptomSpacing(raw string) string {
	fields := strings.Fields(strings.ToValidUTF8(raw, ""))
	return strings.TrimSpace(strings.Join(fields, " "))
}

func normalizeSymptomIconInput(raw string) string {
	icon := strings.TrimSpace(strings.ToValidUTF8(raw, ""))
	if icon == "" {
		return defaultSymptomIcon
	}
	return icon
}

func normalizeSymptomColorInput(raw string) (string, error) {
	color := strings.TrimSpace(strings.ToUpper(strings.ToValidUTF8(raw, "")))
	if !hexSymptomColorPattern.MatchString(color) {
		return "", ErrInvalidSymptomColor
	}
	return color, nil
}

func containsInvalidSymptomNameRune(value string) bool {
	for _, r := range value {
		if unicode.IsControl(r) {
			return true
		}
		switch r {
		case '<', '>':
			return true
		}
	}
	return false
}
