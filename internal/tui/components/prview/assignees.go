package prview

import "strings"

func newAssignees(existing []string, requested []string) []string {
	existingSet := make(map[string]struct{}, len(existing))
	for _, assignee := range existing {
		existingSet[strings.ToLower(strings.TrimSpace(assignee))] = struct{}{}
	}
	added := make([]string, 0, len(requested))
	seen := make(map[string]struct{}, len(requested))
	for _, assignee := range requested {
		trimmed := strings.TrimSpace(assignee)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := existingSet[key]; ok {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		added = append(added, trimmed)
	}
	return added
}

func remainingAssignees(existing []string, remove []string) []string {
	removeSet := make(map[string]struct{}, len(remove))
	for _, assignee := range remove {
		removeSet[strings.ToLower(strings.TrimSpace(assignee))] = struct{}{}
	}
	remaining := make([]string, 0, len(existing))
	for _, assignee := range existing {
		if _, ok := removeSet[strings.ToLower(strings.TrimSpace(assignee))]; ok {
			continue
		}
		remaining = append(remaining, assignee)
	}
	return remaining
}

func assigneesToRemove(existing []string, remove []string) []string {
	existingSet := make(map[string]struct{}, len(existing))
	for _, assignee := range existing {
		existingSet[strings.ToLower(strings.TrimSpace(assignee))] = struct{}{}
	}
	removed := make([]string, 0, len(remove))
	seen := make(map[string]struct{}, len(remove))
	for _, assignee := range remove {
		trimmed := strings.TrimSpace(assignee)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := existingSet[key]; !ok {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		removed = append(removed, trimmed)
	}
	return removed
}
