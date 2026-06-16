package tui

import (
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/textfuel/lazyjira/v2/pkg/ascii"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
)

// mentionScanRe captures a leading boundary (start of string, or a single
// non-word, non-backslash char) in group 1 and the raw @-token in group 2.
// The boundary keeps emails (me@host) and escaped (\@) positions from matching.
var mentionScanRe = regexp.MustCompile(`(^|[^\p{L}\p{N}_\\])@([\p{L}\p{N}_]+)`)

// maskPatterns are blanked out (in order) before scanning so @-tokens inside
// fenced blocks, inline code or already-resolved mentions are never touched.
var maskPatterns = []*regexp.Regexp{
	regexp.MustCompile("(?s)```.*?```"),
	regexp.MustCompile("(?s)~~~.*?~~~"),
	regexp.MustCompile("`[^`]*`"),
	regexp.MustCompile(`\[@[^\]]*\]\(accountid:[^)]*\)`),
}

// normalizeToWords lowercases s, transliterates it to ASCII (German ae/oe/ue/ss
// plus NFD accent stripping), turns '_' into spaces and splits on whitespace.
// Each resulting word has leading and trailing ASCII punctuation (",.") trimmed;
// empty words are dropped.
func normalizeToWords(s string) []string {
	s = strings.ReplaceAll(s, "_", " ")
	s = ascii.Convert(s)
	fields := strings.Fields(s)
	words := make([]string, 0, len(fields))
	for _, f := range fields {
		if f = strings.Trim(f, ",."); f != "" {
			words = append(words, f)
		}
	}
	return words
}

// matchUsers returns every user whose DisplayName matches the token words.
// A single-word token matches when the word is one of the DisplayName words.
// A multi-word token matches only when the words equal the full DisplayName
// word sequence.
func matchUsers(tokenWords []string, users []jira.User) []jira.User {
	var matched []jira.User
	for _, u := range users {
		uw := normalizeToWords(u.DisplayName)
		var ok bool
		if len(tokenWords) == 1 {
			ok = slices.Contains(uw, tokenWords[0])
		} else {
			ok = slices.Equal(tokenWords, uw)
		}
		if ok {
			matched = append(matched, u)
		}
	}
	return matched
}

// hasMentionCandidate reports whether md contains at least one @-token that the
// scanner could replace, letting callers skip the (possibly async) mention
// machinery for ordinary text. May return true for tokens that ultimately stay
// literal (e.g. inside code spans); resolveMentions then leaves them untouched.
func hasMentionCandidate(md string) bool {
	return mentionScanRe.MatchString(md)
}

// resolveMentions replaces every unambiguous @-token in md with an ADF mention
// link [@DisplayName](accountid:ID). Tokens that match zero or multiple users
// stay literal, as do tokens inside masked spans (see maskPatterns).
func resolveMentions(md string, users []jira.User) string {
	masked, spans := maskSpans(md)
	masked = mentionScanRe.ReplaceAllStringFunc(masked, func(m string) string {
		sub := mentionScanRe.FindStringSubmatch(m)
		prefix, rawToken := sub[1], sub[2]
		matches := matchUsers(normalizeToWords(rawToken), users)
		if len(matches) != 1 {
			return m
		}
		u := matches[0]
		return prefix + "[@" + u.DisplayName + "](accountid:" + u.AccountID + ")"
	})
	return restoreSpans(masked, spans)
}

// maskSpans replaces each protected span with a NUL-delimited placeholder and
// returns the masked string plus the original spans, indexed by placeholder.
func maskSpans(md string) (string, []string) {
	var spans []string
	for _, re := range maskPatterns {
		md = re.ReplaceAllStringFunc(md, func(m string) string {
			placeholder := "\x00" + strconv.Itoa(len(spans)) + "\x00"
			spans = append(spans, m)
			return placeholder
		})
	}
	return md, spans
}

// restoreSpans reverses maskSpans, putting every original span back in place.
func restoreSpans(md string, spans []string) string {
	for i, s := range spans {
		md = strings.ReplaceAll(md, "\x00"+strconv.Itoa(i)+"\x00", s)
	}
	return md
}
