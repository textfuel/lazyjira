package components

import "regexp"

const (
	testTitle        = "Test Title"
	testItem1Label   = "First Item"
	testItem2Label   = "Second Item"
	testItem3Label   = "Third Item"
	testItem1ID      = "id-1"
	testItem2ID      = "id-2"
	testItem3ID      = "id-3"
	testSummaryText  = "My issue summary"
	testProjectKey   = "PROJ"
	testIssueType    = "Story"
	testFieldName    = "Priority"
	testFieldID      = "priority"
	testBranchName1  = "feat/PROJ-1-add-login"
	testBranchName2  = "fix/PROJ-2-fix-crash"
	testQueryText    = "project = PROJ ORDER BY updated DESC"
	testHistoryItem1 = "project = PROJ ORDER BY created DESC"
	testHistoryItem2 = "assignee = currentUser()"
)

var ansiEscapeRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiEscapeRe.ReplaceAllString(s, "")
}
