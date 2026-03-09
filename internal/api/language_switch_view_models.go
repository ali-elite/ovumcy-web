package api

import (
	"fmt"
	"net/url"
	"strings"
)

type languageSwitchOption struct {
	Code   string
	Href   string
	Label  string
	Active bool
}

func buildLanguageSwitchOptions(messages map[string]string, currentPath string, currentLanguage string, supported []string) []languageSwitchOption {
	options := make([]languageSwitchOption, 0, len(supported))
	for _, code := range supported {
		normalizedCode := strings.TrimSpace(code)
		if normalizedCode == "" {
			continue
		}
		options = append(options, languageSwitchOption{
			Code:   normalizedCode,
			Href:   buildLanguageSwitchHref(normalizedCode, currentPath),
			Label:  localizedLanguageSwitchLabel(messages, normalizedCode),
			Active: normalizedCode == currentLanguage,
		})
	}
	return options
}

func buildLanguageSwitchHref(code string, currentPath string) string {
	return fmt.Sprintf("/lang/%s?next=%s", url.PathEscape(code), url.QueryEscape(currentPath))
}

func localizedLanguageSwitchLabel(messages map[string]string, code string) string {
	key := fmt.Sprintf("lang.%s", code)
	localized := translateMessage(messages, key)
	if localized == key || strings.TrimSpace(localized) == "" {
		return strings.ToUpper(code)
	}
	return localized
}
