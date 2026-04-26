package api

import (
	"strings"
	"testing"
)

func TestRenderTemplateMarkdownSupportsAdviceFormatting(t *testing.T) {
	rendered := string(renderTemplateMarkdown("### Physical Comfort\nUse **heat** and `gentle` support.\n* Bring tea\n* Offer rest\n1. Ask first"))

	expected := []string{
		"<h3>Physical Comfort</h3>",
		"<p>Use <strong>heat</strong> and <code>gentle</code> support.</p>",
		"<ul><li>Bring tea</li><li>Offer rest</li></ul>",
		"<ol><li>Ask first</li></ol>",
	}
	for _, fragment := range expected {
		if !strings.Contains(rendered, fragment) {
			t.Fatalf("expected rendered markdown to include %q, got %q", fragment, rendered)
		}
	}
}

func TestRenderTemplateMarkdownEscapesRawHTML(t *testing.T) {
	rendered := string(renderTemplateMarkdown("Hello <script>alert(1)</script> **safe**"))

	if strings.Contains(rendered, "<script>") {
		t.Fatalf("expected script tag to be escaped, got %q", rendered)
	}
	if !strings.Contains(rendered, "&lt;script&gt;alert(1)&lt;/script&gt;") {
		t.Fatalf("expected escaped script text, got %q", rendered)
	}
	if !strings.Contains(rendered, "<strong>safe</strong>") {
		t.Fatalf("expected markdown emphasis to render, got %q", rendered)
	}
}
