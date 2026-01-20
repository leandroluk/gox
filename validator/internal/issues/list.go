// internal/issues/list.go
package issues

type List struct {
	items []Issue
}

func NewList() List {
	return List{items: make([]Issue, 0)}
}

func (list *List) Add(issue Issue) {
	if issue.IsZero() {
		return
	}
	list.items = append(list.items, issue)
}

func (list *List) AddWithLimit(issue Issue, maxIssues int) bool {
	if issue.IsZero() {
		return false
	}
	if maxIssues > 0 && len(list.items) >= maxIssues {
		return true
	}
	list.items = append(list.items, issue)
	if maxIssues > 0 && len(list.items) >= maxIssues {
		return true
	}
	return false
}

func (list *List) Merge(other List) {
	if len(other.items) == 0 {
		return
	}
	list.items = append(list.items, other.items...)
}

func (list *List) MergeWithLimit(other List, maxIssues int) bool {
	if len(other.items) == 0 {
		return false
	}
	if maxIssues <= 0 {
		list.items = append(list.items, other.items...)
		return false
	}

	for _, issue := range other.items {
		if len(list.items) >= maxIssues {
			return true
		}
		if issue.IsZero() {
			continue
		}
		list.items = append(list.items, issue)
	}

	return len(list.items) >= maxIssues
}

func (list *List) Items() []Issue {
	return append([]Issue(nil), list.items...)
}

func (list *List) Len() int {
	return len(list.items)
}

func (list *List) IsEmpty() bool {
	return len(list.items) == 0
}

func (list *List) Clear() {
	list.items = list.items[:0]
}
