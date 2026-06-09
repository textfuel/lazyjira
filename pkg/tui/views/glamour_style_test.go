package views

import "testing"

func TestResolveGlamourStyle_Explicit(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"dark":  "dark",
		"light": "light",
		"notty": "notty",
	}
	for in, want := range cases {
		if got := ResolveGlamourStyle(in); got != want {
			t.Errorf("ResolveGlamourStyle(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestResolveGlamourStyle_AutoDark(t *testing.T) {
	t.Parallel()
	darkBackground := func() bool { return true }
	for _, in := range []string{"", "auto"} {
		if got := resolveGlamourStyle(in, darkBackground); got != "dark" {
			t.Errorf("resolveGlamourStyle(%q) with dark bg = %q, want %q", in, got, "dark")
		}
	}
}

func TestResolveGlamourStyle_AutoLight(t *testing.T) {
	t.Parallel()
	lightBackground := func() bool { return false }
	for _, in := range []string{"", "auto"} {
		if got := resolveGlamourStyle(in, lightBackground); got != "light" {
			t.Errorf("resolveGlamourStyle(%q) with light bg = %q, want %q", in, got, "light")
		}
	}
}

func TestResolveGlamourStyle_UnknownFallsBackToAuto(t *testing.T) {
	t.Parallel()
	darkBackground := func() bool { return true }
	if got := resolveGlamourStyle("dracula", darkBackground); got != "dark" {
		t.Errorf("resolveGlamourStyle(unknown) = %q, want %q", got, "dark")
	}
}
