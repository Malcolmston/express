package nodepath

import "testing"

func TestBasename(t *testing.T) {
	tests := []struct{ in, want string }{
		{"/foo/bar/baz/asdf/quux.html", "quux.html"},
		{"/foo/bar/baz/asdf/quux/", "quux"},
		{"/foo/bar/", "bar"},
		{"foo", "foo"},
		{"/", ""},
		{"", ""},
		{"////", ""},
	}
	for _, tt := range tests {
		if got := Basename(tt.in); got != tt.want {
			t.Errorf("Basename(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestBasenameExt(t *testing.T) {
	tests := []struct{ in, ext, want string }{
		{"/foo/bar/baz/asdf/quux.html", ".html", "quux"},
		{"index.html", ".html", "index"},
		{"index.html", ".txt", "index.html"},
		{"foo.txt", ".txt", "foo"},
	}
	for _, tt := range tests {
		if got := BasenameExt(tt.in, tt.ext); got != tt.want {
			t.Errorf("BasenameExt(%q,%q) = %q, want %q", tt.in, tt.ext, got, tt.want)
		}
	}
}

func TestDirname(t *testing.T) {
	tests := []struct{ in, want string }{
		{"/foo/bar/baz/asdf/quux", "/foo/bar/baz/asdf"},
		{"/foo/bar/", "/foo"},
		{"foo", "."},
		{"/", "/"},
		{"", "."},
		{"foo/bar", "foo"},
	}
	for _, tt := range tests {
		if got := Dirname(tt.in); got != tt.want {
			t.Errorf("Dirname(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestExtname(t *testing.T) {
	tests := []struct{ in, want string }{
		{"index.html", ".html"},
		{"index.coffee.md", ".md"},
		{"index.", "."},
		{"index", ""},
		{".index", ""},
		{".index.md", ".md"},
		{"/path/to/file.txt", ".txt"},
	}
	for _, tt := range tests {
		if got := Extname(tt.in); got != tt.want {
			t.Errorf("Extname(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct{ in, want string }{
		{"/foo/bar//baz/asdf/quux/..", "/foo/bar/baz/asdf"},
		{"foo/bar/../baz", "foo/baz"},
		{"/foo/../..", "/"},
		{"a//b//../b", "a/b"},
		{"./fixtures///b/../b/c.js", "fixtures/b/c.js"},
		{"", "."},
		{"foo/", "foo/"},
	}
	for _, tt := range tests {
		if got := Normalize(tt.in); got != tt.want {
			t.Errorf("Normalize(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestJoin(t *testing.T) {
	tests := []struct {
		in   []string
		want string
	}{
		{[]string{"/foo", "bar", "baz/asdf", "quux", ".."}, "/foo/bar/baz/asdf"},
		{[]string{"foo", "../../bar"}, "../bar"},
		{[]string{}, "."},
		{[]string{"", ""}, "."},
		{[]string{"a", "b", "c"}, "a/b/c"},
		{[]string{"/a", "", "b"}, "/a/b"},
	}
	for _, tt := range tests {
		if got := Join(tt.in...); got != tt.want {
			t.Errorf("Join(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestIsAbsolute(t *testing.T) {
	if !IsAbsolute("/foo/bar") || IsAbsolute("foo/bar") || IsAbsolute("") {
		t.Error("IsAbsolute")
	}
}

func TestResolve(t *testing.T) {
	tests := []struct {
		in   []string
		want string
	}{
		{[]string{"/foo/bar", "./baz"}, "/foo/bar/baz"},
		{[]string{"/foo/bar", "/tmp/file/"}, "/tmp/file"},
		{[]string{"/a/b", "c", "..", "d"}, "/a/b/d"},
	}
	for _, tt := range tests {
		if got := Resolve(tt.in...); got != tt.want {
			t.Errorf("Resolve(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestRelative(t *testing.T) {
	tests := []struct {
		from, to, want string
	}{
		{"/data/orandea/test/aaa", "/data/orandea/impl/bbb", "../../impl/bbb"},
		{"/a/b/c", "/a/b/c", ""},
		{"/a/b", "/a/b/c/d", "c/d"},
		{"/a/b/c/d", "/a/b", "../.."},
	}
	for _, tt := range tests {
		if got := Relative(tt.from, tt.to); got != tt.want {
			t.Errorf("Relative(%q,%q) = %q, want %q", tt.from, tt.to, got, tt.want)
		}
	}
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		in         string
		root, dir  string
		base, name string
		ext        string
	}{
		{"/home/user/dir/file.txt", "/", "/home/user/dir", "file.txt", "file", ".txt"},
		{"foo.txt", "", "", "foo.txt", "foo", ".txt"},
		{"/", "/", "/", "", "", ""},
	}
	for _, tt := range tests {
		p := Parse(tt.in)
		if p.Root != tt.root || p.Dir != tt.dir || p.Base != tt.base || p.Name != tt.name || p.Ext != tt.ext {
			t.Errorf("Parse(%q) = %+v", tt.in, p)
		}
		if got := Format(p); got != tt.in {
			t.Errorf("Format(Parse(%q)) = %q", tt.in, got)
		}
	}
}

func BenchmarkNormalize(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Normalize("/foo/bar//baz/asdf/quux/..")
	}
}
