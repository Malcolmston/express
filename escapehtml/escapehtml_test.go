package escapehtml

import "testing"

func TestEscape(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"no special chars", "no special chars"},
		{"&", "&amp;"},
		{"<", "&lt;"},
		{">", "&gt;"},
		{"\"", "&quot;"},
		{"'", "&#39;"},
		{"<script>alert('x & y')</script>",
			"&lt;script&gt;alert(&#39;x &amp; y&#39;)&lt;/script&gt;"},
		{"a & b < c > d", "a &amp; b &lt; c &gt; d"},
		{"\"quoted\"", "&quot;quoted&quot;"},
	}
	for _, tt := range tests {
		if got := Escape(tt.in); got != tt.want {
			t.Fatalf("Escape(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestEscapeNoOpReturnsInput(t *testing.T) {
	in := "plain unicode € text"
	if got := Escape(in); got != in {
		t.Fatalf("Escape(%q) = %q, want unchanged", in, got)
	}
}
