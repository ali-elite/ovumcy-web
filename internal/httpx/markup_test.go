package httpx

import "testing"

func TestStatusErrorMarkupEscapesHTML(t *testing.T) {
	got := StatusErrorMarkup(`<script>alert("x")</script>`)
	want := `<div class="status-error">&lt;script&gt;alert(&#34;x&#34;)&lt;/script&gt;</div>`
	if got != want {
		t.Fatalf("unexpected markup: got %q want %q", got, want)
	}
}
