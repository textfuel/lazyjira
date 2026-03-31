Output e2e/golden/01_create_issue.gif

Set Shell bash
Set Width 1200
Set Height 600
Set FontSize 14
Set TypingSpeed 0ms
Set Theme "Catppuccin Mocha"

Type "./lazyjira --demo"
Enter
Sleep 1s

# select project and issue
Type "4"
Sleep 200ms
Type "j"
Sleep 200ms
Space
Sleep 400ms

# focus issues panel, press n to create
Type "2"
Sleep 200ms
Type "n"
Sleep 400ms

# type picker: select Bug
Type "j"
Sleep 200ms
Enter
Sleep 600ms

# create form opens, summary focused
# type the summary
Set TypingSpeed 40ms
Type "Login page crashes on expired token refresh"
Set TypingSpeed 0ms
Sleep 300ms

# tab to description
Tab
Sleep 200ms

# tab to fields
Tab
Sleep 200ms

# edit priority: press e, select High
Type "e"
Sleep 300ms
Type "j"
Sleep 200ms
Enter
Sleep 300ms

# move to assignee, edit with enter
Type "j"
Sleep 200ms
Enter
Sleep 300ms

# pick Alice Chen
Type "j"
Sleep 200ms
Type "j"
Sleep 200ms
Enter
Sleep 300ms

# move to labels, edit with space
Type "j"
Sleep 200ms
Space
Sleep 300ms

# toggle two labels
Space
Sleep 200ms
Type "j"
Sleep 200ms
Space
Sleep 200ms

# confirm checklist
Enter
Sleep 300ms

# scroll down to see all fields
Type "j"
Sleep 200ms

# go back to summary to review
Tab
Sleep 200ms
Sleep 300ms

# tab to description
Tab
Sleep 200ms

# tab to fields
Tab
Sleep 200ms
Sleep 300ms

# submit with enter from summary
Tab
Sleep 200ms
Enter
Sleep 600ms

# issue created, back to issues list
Sleep 500ms
Type "q"
Sleep 300ms
