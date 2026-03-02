package httpx

import (
	"fmt"
	"html/template"
)

// StatusErrorMarkup renders the shared HTMX status-error wrapper.
func StatusErrorMarkup(message string) string {
	return fmt.Sprintf("<div class=\"status-error\">%s</div>", template.HTMLEscapeString(message))
}
