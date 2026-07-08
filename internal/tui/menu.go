package tui

import "strings"

// listItem is one row of a selectList.
type listItem struct {
	label   string
	checked bool // only meaningful when the list is in multi-select mode
}

// selectList is a minimal up/down/enter (and, in multi mode, space-to-toggle)
// menu — the two interactive screens (action menu, module picker) are both
// just this. Not pulling in the bubbles list component: fuzzy filtering and
// pagination aren't needed here, and this is ~40 lines.
type selectList struct {
	items  []listItem
	cursor int
	multi  bool
}

func newSelectList(labels []string, multi bool) selectList {
	items := make([]listItem, len(labels))
	for i, l := range labels {
		items[i] = listItem{label: l}
	}
	return selectList{items: items, multi: multi}
}

func (l *selectList) up() {
	if l.cursor > 0 {
		l.cursor--
	}
}

func (l *selectList) down() {
	if l.cursor < len(l.items)-1 {
		l.cursor++
	}
}

// toggle flips the checked state of the item under the cursor. No-op in
// single-select mode, where enter acts immediately instead.
func (l *selectList) toggle() {
	if l.multi && len(l.items) > 0 {
		l.items[l.cursor].checked = !l.items[l.cursor].checked
	}
}

// selectedIndices returns checked item indices (multi mode only).
func (l selectList) selectedIndices() []int {
	var out []int
	for i, it := range l.items {
		if it.checked {
			out = append(out, i)
		}
	}
	return out
}

func (l selectList) View() string {
	var b strings.Builder
	for i, it := range l.items {
		cursor := "  "
		if i == l.cursor {
			cursor = "> "
		}
		box := ""
		if l.multi {
			box = "[ ] "
			if it.checked {
				box = "[x] "
			}
		}
		b.WriteString(cursor + box + it.label + "\n")
	}
	return b.String()
}
