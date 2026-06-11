package ascii

import "testing"

func TestConvert(t *testing.T) {
	t.Parallel()
	cases := []struct{ in, want string }{
		{"", ""},
		{"hello", "hello"},
		{"HELLO", "hello"},
		{"Größe", "groesse"},
		{"über alles", "ueber alles"},
		{"straße", "strasse"},
		{"café", "cafe"},
		{"naïve", "naive"},
		{"jalapeño", "jalapeno"},
		{"keeps punctuation: a/b.c_d", "keeps punctuation: a/b.c_d"},
		{"a_b<c>d|e$f%g`h", "a_b<c>d|e$f%g`h"},
	}
	for _, tc := range cases {
		if got := Convert(tc.in); got != tc.want {
			t.Errorf("Convert(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
