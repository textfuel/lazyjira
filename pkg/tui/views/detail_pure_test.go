package views

import (
	"testing"
	"time"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
)

func TestCleanWikiMarkup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty stays empty", "", ""},
		{"account mention becomes user", "ping [~accountid:abc-123] now", "ping @user now"},
		{"code block keeps content", "see {code:go}x := 1{code} done", "see x := 1 done"},
		{"link keeps label", "go [docs|http://x] there", "go docs there"},
		{"strips carriage returns", "line\rtext", "linetext"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testkit.AssertEqual(t, "cleaned", cleanWikiMarkup(tt.in), tt.want)
		})
	}
}

func TestWrapText(t *testing.T) {
	t.Parallel()

	t.Run("short line stays intact", func(t *testing.T) {
		t.Parallel()
		lines := wrapText("hello world", 80)
		testkit.AssertSliceEqual(t, "lines", lines, []string{"hello world"})
	})

	t.Run("wraps on word boundary", func(t *testing.T) {
		t.Parallel()
		lines := wrapText("alpha beta gamma", 11)
		if len(lines) < 2 {
			t.Fatalf("expected wrapping, got %v", lines)
		}
		if lines[0] != "alpha beta" {
			t.Errorf("first line = %q, want 'alpha beta'", lines[0])
		}
	})

	t.Run("non positive width defaults", func(t *testing.T) {
		t.Parallel()
		lines := wrapText("short", 0)
		testkit.AssertSliceEqual(t, "lines", lines, []string{"short"})
	})
}

func TestFindURLs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want []string
	}{
		{"plain url", "see https://example.com here", []string{"https://example.com"}},
		{"strips trailing punctuation", "(http://x.com).", []string{"http://x.com"}},
		{"no url", "nothing here", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testkit.AssertSliceEqual(t, "urls", findURLs(tt.in), tt.want)
		})
	}
}

func TestExtractURLs(t *testing.T) {
	t.Parallel()

	t.Run("nil issue returns nil", func(t *testing.T) {
		t.Parallel()
		if groups := ExtractURLs(nil, "https://h"); groups != nil {
			t.Errorf("expected nil, got %v", groups)
		}
	})

	t.Run("groups body links and history", func(t *testing.T) {
		t.Parallel()
		issue := &jira.Issue{
			Key:         testKey,
			Description: "ref https://body.example.com end",
			IssueLinks: []jira.IssueLink{
				{OutwardIssue: &jira.Issue{Key: testKey2}},
			},
			Changelog: []jira.ChangelogEntry{
				{Items: []jira.ChangeItem{{ToString: "moved https://hist.example.com now"}}},
			},
		}

		groups := ExtractURLs(issue, "https://h")

		sections := make(map[string][]string)
		for _, g := range groups {
			sections[g.Section] = g.URLs
		}
		if len(sections["Body"]) != 1 || sections["Body"][0] != "https://body.example.com" {
			t.Errorf("body group = %v", sections["Body"])
		}
		if len(sections["Links"]) != 1 || sections["Links"][0] != "https://h/browse/PLAT-2" {
			t.Errorf("links group = %v", sections["Links"])
		}
		if len(sections["History"]) != 1 {
			t.Errorf("history group = %v", sections["History"])
		}
	})

	t.Run("prefers adf description urls", func(t *testing.T) {
		t.Parallel()
		issue := &jira.Issue{
			Key: testKey,
			DescriptionADF: map[string]any{
				"type": "doc",
				"content": []any{
					map[string]any{"type": "inlineCard", "attrs": map[string]any{"url": "https://card.example.com"}},
				},
			},
		}

		groups := ExtractURLs(issue, "https://h")

		if len(groups) == 0 || groups[0].Section != "Body" || groups[0].URLs[0] != "https://card.example.com" {
			t.Errorf("adf body urls = %v", groups)
		}
	})
}

func TestExtractADFURLs(t *testing.T) {
	t.Parallel()

	adf := map[string]any{
		"type": "doc",
		"content": []any{
			map[string]any{
				"type": adfText,
				"text": "link",
				"marks": []any{
					map[string]any{"type": "link", "attrs": map[string]any{"href": "https://marked.example.com"}},
				},
			},
			map[string]any{"type": "inlineCard", "attrs": map[string]any{"url": "https://card.example.com"}},
		},
	}

	urls := extractADFURLs(adf)

	if len(urls) != 2 {
		t.Fatalf("urls = %v, want 2", urls)
	}
}

func TestTimeAgo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ago  time.Duration
		want string
	}{
		{"seconds is just now", 10 * time.Second, "just now"},
		{"minutes", 5 * time.Minute, "5m ago"},
		{"hours", 3 * time.Hour, "3h ago"},
		{"days", 4 * 24 * time.Hour, "4d ago"},
		{"months", 90 * 24 * time.Hour, "3mo ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := timeAgo(time.Now().Add(-tt.ago))
			testkit.AssertEqual(t, "time ago", got, tt.want)
		})
	}
}

func TestRenderDescriptionPreview(t *testing.T) {
	t.Parallel()

	t.Run("empty text returns nil", func(t *testing.T) {
		t.Parallel()
		if lines := RenderDescriptionPreview("", 40, false, BuiltinRenderer{}); lines != nil {
			t.Errorf("expected nil, got %v", lines)
		}
	})

	t.Run("server mode wraps and strips wiki", func(t *testing.T) {
		t.Parallel()
		lines := RenderDescriptionPreview("h1. Title text", 40, false, BuiltinRenderer{})
		if len(lines) == 0 {
			t.Fatal("expected rendered lines")
		}
	})
}
