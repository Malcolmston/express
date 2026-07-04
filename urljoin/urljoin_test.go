package urljoin

import "testing"

func TestURLJoin(t *testing.T) {
	tests := []struct {
		name  string
		parts []string
		want  string
	}{
		{
			name:  "basic",
			parts: []string{"http://www.google.com", "a", "b", "c"},
			want:  "http://www.google.com/a/b/c",
		},
		{
			name:  "leading slash on segment",
			parts: []string{"http://example.com/", "/foo"},
			want:  "http://example.com/foo",
		},
		{
			name:  "mixed slashes",
			parts: []string{"http://example.com", "a/", "/b", "/c"},
			want:  "http://example.com/a/b/c",
		},
		{
			name:  "query strings combined",
			parts: []string{"http://example.com", "?foo=bar", "?bar=baz"},
			want:  "http://example.com?foo=bar&bar=baz",
		},
		{
			name:  "single query",
			parts: []string{"http://example.com/", "foo?bar=baz"},
			want:  "http://example.com/foo?bar=baz",
		},
		{
			// url-join strips the slash before a "#" fragment (its regex
			// matches "#" not followed by "!"), so no slash is kept here.
			name:  "hash fragment",
			parts: []string{"http://example.com", "/#/home"},
			want:  "http://example.com#/home",
		},
		{
			name:  "relative path",
			parts: []string{"/", "foo", "bar"},
			want:  "/foo/bar",
		},
		{
			name:  "empty segments skipped",
			parts: []string{"http://example.com", "", "foo", ""},
			want:  "http://example.com/foo",
		},
		{
			name:  "bare protocol merged",
			parts: []string{"http://", "example.com", "a"},
			want:  "http://example.com/a",
		},
		{
			name:  "file protocol three slashes",
			parts: []string{"file:///", "home", "user"},
			want:  "file:///home/user",
		},
		{
			name:  "collapse duplicate slashes in first",
			parts: []string{"http://example.com//", "foo"},
			want:  "http://example.com/foo",
		},
		{
			name:  "single part",
			parts: []string{"http://example.com"},
			want:  "http://example.com",
		},
		{
			name:  "trailing slash preserved on last",
			parts: []string{"http://example.com", "foo/"},
			want:  "http://example.com/foo/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := URLJoin(tt.parts...)
			if got != tt.want {
				t.Fatalf("URLJoin(%q) = %q, want %q", tt.parts, got, tt.want)
			}
		})
	}
}

func TestURLJoinEmpty(t *testing.T) {
	if got := URLJoin(); got != "" {
		t.Fatalf("URLJoin() = %q, want empty string", got)
	}
}
