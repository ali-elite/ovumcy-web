package httpx

import (
	"fmt"
	"html/template"
)

// StatusOKMarkup renders the shared non-dismissible status-ok wrapper.
func StatusOKMarkup(message string) string {
	return fmt.Sprintf(
		"<div class=\"status-ok\"><span class=\"toast-message-wrap\"><span class=\"toast-icon\" aria-hidden=\"true\">✓</span><span class=\"toast-message\">%s</span></span></div>",
		template.HTMLEscapeString(message),
	)
}

// StatusOKTemplateHTML returns trusted shared success markup after escaping message content.
func StatusOKTemplateHTML(message string) template.HTML {
	return trustedEscapedHTML(StatusOKMarkup(message))
}

// StatusErrorMarkup renders the shared HTMX status-error wrapper.
func StatusErrorMarkup(message string) string {
	return fmt.Sprintf("<div class=\"status-error\">%s</div>", template.HTMLEscapeString(message))
}

// DismissibleStatusOKMarkup renders the shared HTMX dismissible status-ok wrapper.
func DismissibleStatusOKMarkup(message string, closeLabel string) string {
	return fmt.Sprintf(
		"<div class=\"status-ok\"><div class=\"toast-body\"><span class=\"toast-message-wrap\"><span class=\"toast-icon\" aria-hidden=\"true\">✓</span><span class=\"toast-message\">%s</span></span><button type=\"button\" class=\"toast-close\" data-dismiss-status aria-label=\"%s\">×</button></div></div>",
		template.HTMLEscapeString(message),
		template.HTMLEscapeString(closeLabel),
	)
}

// DismissibleStatusOKTemplateHTML returns trusted dismissible success markup after escaping message content.
func DismissibleStatusOKTemplateHTML(message string, closeLabel string) template.HTML {
	return trustedEscapedHTML(DismissibleStatusOKMarkup(message, closeLabel))
}

// #nosec G203 -- shared status markup is built from already escaped strings in this package.
func trustedEscapedHTML(markup string) template.HTML {
	return template.HTML(markup)
}
