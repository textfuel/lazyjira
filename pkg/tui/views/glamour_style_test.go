package views

import "testing"

func TestResolveGlamourStyle_Explicit(t *testing.T) {
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
	orig := hasDarkBackground
	t.Cleanup(func() { hasDarkBackground = orig })
	hasDarkBackground = func() bool { return true }

	for _, in := range []string{"", "auto"} {
		if got := ResolveGlamourStyle(in); got != "dark" {
			t.Errorf("ResolveGlamourStyle(%q) with dark bg = %q, want %q", in, got, "dark")
		}
	}
}

func TestResolveGlamourStyle_AutoLight(t *testing.T) {
	orig := hasDarkBackground
	t.Cleanup(func() { hasDarkBackground = orig })
	hasDarkBackground = func() bool { return false }

	for _, in := range []string{"", "auto"} {
		if got := ResolveGlamourStyle(in); got != "light" {
			t.Errorf("ResolveGlamourStyle(%q) with light bg = %q, want %q", in, got, "light")
		}
	}
}

func TestResolveGlamourStyle_UnknownFallsBackToAuto(t *testing.T) {
	orig := hasDarkBackground
	t.Cleanup(func() { hasDarkBackground = orig })
	hasDarkBackground = func() bool { return true }

	if got := ResolveGlamourStyle("dracula"); got != "dark" {
		t.Errorf("ResolveGlamourStyle(unknown) = %q, want %q", got, "dark")
	}
}
