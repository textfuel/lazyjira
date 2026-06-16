package tui

import (
	"slices"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/jira"
)

func TestNormalizeToWords(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"two words", "Jane Doe", []string{"jane", "doe"}},
		{"umlaut", "Müller", []string{"mueller"}},
		{"accent", "José", []string{"jose"}},
		{"sharp-s and umlaut", "Größe", []string{"groesse"}},
		{"comma trimmed", "Doe, Jane", []string{"doe", "jane"}},
		{"underscore to space", "jane_doe", []string{"jane", "doe"}},
		{"empty", "", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := normalizeToWords(tc.in)
			if !slices.Equal(got, tc.want) {
				t.Errorf("normalizeToWords(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestMatchUsers(t *testing.T) {
	t.Parallel()
	users := []jira.User{
		{DisplayName: "Jane Doe", AccountID: "a1"},
		{DisplayName: "John Doe", AccountID: "a2"},
		{DisplayName: "Jürgen Müller", AccountID: "a3"},
	}
	cases := []struct {
		name  string
		token []string
		want  []string // matched account IDs, in user order
	}{
		{"single unique", []string{"jane"}, []string{"a1"}},
		{"single ambiguous", []string{"doe"}, []string{"a1", "a2"}},
		{"multi-word exact", []string{"jane", "doe"}, []string{"a1"}},
		{"transliterated", []string{"mueller"}, []string{"a3"}},
		{"no match", []string{"jose"}, nil},
		{"multi-word no prefix", []string{"jane", "d"}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := matchUsers(tc.token, users)
			ids := make([]string, 0, len(got))
			for _, u := range got {
				ids = append(ids, u.AccountID)
			}
			if !slices.Equal(ids, tc.want) {
				t.Errorf("matchUsers(%v) ids = %v, want %v", tc.token, ids, tc.want)
			}
		})
	}
}

func TestResolveMentions(t *testing.T) {
	t.Parallel()
	users := []jira.User{
		{DisplayName: "Jane Doe", AccountID: "a1"},
		{DisplayName: "John Doe", AccountID: "a2"},
		{DisplayName: "Jürgen Müller", AccountID: "a3"},
		{DisplayName: "Solo One", AccountID: "s1"},
	}
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"underscore token", "hi @Solo_One!", "hi [@Solo One](accountid:s1)!"},
		{"transliterated token", "ping @mueller", "ping [@Jürgen Müller](accountid:a3)"},
		{"umlaut token", "ping @Müller", "ping [@Jürgen Müller](accountid:a3)"},
		{"ambiguous unchanged", "@doe", "@doe"},
		{"no match unchanged", "@nobody", "@nobody"},
		{"email unchanged", "mail me@solo.com", "mail me@solo.com"},
		{"handle unchanged", "git@host", "git@host"},
		{"escaped unchanged", "\\@solo", "\\@solo"},
		{"inline code unchanged", "`@solo`", "`@solo`"},
		{"fenced unchanged", "```\n@solo\n```", "```\n@solo\n```"},
		{"already resolved unchanged", "[@Solo One](accountid:s1)", "[@Solo One](accountid:s1)"},
		{"multiple replaced", "@mueller and @Solo_One", "[@Jürgen Müller](accountid:a3) and [@Solo One](accountid:s1)"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := resolveMentions(tc.in, users)
			if got != tc.want {
				t.Errorf("resolveMentions(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
