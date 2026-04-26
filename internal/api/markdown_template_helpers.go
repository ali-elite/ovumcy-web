package api

import (
	"html/template"
	"regexp"
	"strings"
)

var orderedMarkdownItemPattern = regexp.MustCompile(`^\d+\.\s+(.+)$`)

func renderTemplateMarkdown(raw string) template.HTML {
	lines := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	var builder strings.Builder
	listMode := ""

	closeList := func() {
		switch listMode {
		case "ul":
			builder.WriteString("</ul>")
		case "ol":
			builder.WriteString("</ol>")
		}
		listMode = ""
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			closeList()
			continue
		}

		if text, ok := strings.CutPrefix(trimmed, "### "); ok {
			closeList()
			builder.WriteString(`<h3>`)
			builder.WriteString(renderMarkdownInline(text))
			builder.WriteString(`</h3>`)
			continue
		}
		if text, ok := strings.CutPrefix(trimmed, "## "); ok {
			closeList()
			builder.WriteString(`<h2>`)
			builder.WriteString(renderMarkdownInline(text))
			builder.WriteString(`</h2>`)
			continue
		}
		if text, ok := strings.CutPrefix(trimmed, "# "); ok {
			closeList()
			builder.WriteString(`<h2>`)
			builder.WriteString(renderMarkdownInline(text))
			builder.WriteString(`</h2>`)
			continue
		}

		if text, ok := markdownBulletText(trimmed); ok {
			if listMode != "ul" {
				closeList()
				builder.WriteString("<ul>")
				listMode = "ul"
			}
			builder.WriteString("<li>")
			builder.WriteString(renderMarkdownInline(text))
			builder.WriteString("</li>")
			continue
		}

		if matches := orderedMarkdownItemPattern.FindStringSubmatch(trimmed); len(matches) == 2 {
			if listMode != "ol" {
				closeList()
				builder.WriteString("<ol>")
				listMode = "ol"
			}
			builder.WriteString("<li>")
			builder.WriteString(renderMarkdownInline(matches[1]))
			builder.WriteString("</li>")
			continue
		}

		closeList()
		builder.WriteString("<p>")
		builder.WriteString(renderMarkdownInline(trimmed))
		builder.WriteString("</p>")
	}

	closeList()
	// #nosec G203 -- renderMarkdownInline escapes all user/model text before limited markup is inserted.
	return template.HTML(builder.String())
}

func markdownBulletText(line string) (string, bool) {
	for _, prefix := range []string{"* ", "- "} {
		if text, ok := strings.CutPrefix(line, prefix); ok {
			return text, true
		}
	}
	return "", false
}

func renderMarkdownInline(raw string) string {
	escaped := template.HTMLEscapeString(raw)
	escaped = replaceMarkdownPair(escaped, "`", "<code>", "</code>")
	escaped = replaceMarkdownPair(escaped, "**", "<strong>", "</strong>")
	escaped = replaceMarkdownPair(escaped, "__", "<strong>", "</strong>")
	escaped = replaceMarkdownPair(escaped, "*", "<em>", "</em>")
	escaped = replaceMarkdownPair(escaped, "_", "<em>", "</em>")
	return escaped
}

func replaceMarkdownPair(input string, marker string, openTag string, closeTag string) string {
	var builder strings.Builder
	remaining := input
	for {
		start := strings.Index(remaining, marker)
		if start < 0 {
			builder.WriteString(remaining)
			return builder.String()
		}
		end := strings.Index(remaining[start+len(marker):], marker)
		if end < 0 {
			builder.WriteString(remaining)
			return builder.String()
		}
		end += start + len(marker)
		builder.WriteString(remaining[:start])
		builder.WriteString(openTag)
		builder.WriteString(remaining[start+len(marker) : end])
		builder.WriteString(closeTag)
		remaining = remaining[end+len(marker):]
	}
}
