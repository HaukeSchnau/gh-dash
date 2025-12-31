package providers

import (
	"sort"
	"strings"
)

func FilterInstances(instances []Instance, include, exclude []string) []Instance {
	if len(instances) == 0 {
		return nil
	}

	normalizedInclude := normalizePatterns(include)
	normalizedExclude := normalizePatterns(exclude)

	ordered := append([]Instance(nil), instances...)
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].ID < ordered[j].ID
	})

	var selected []Instance
	seen := make(map[string]struct{}, len(ordered))
	if len(normalizedInclude) > 0 {
		for _, pattern := range normalizedInclude {
			for _, instance := range ordered {
				if _, ok := seen[instance.ID]; ok {
					continue
				}
				if MatchesPattern(instance, pattern) {
					selected = append(selected, instance)
					seen[instance.ID] = struct{}{}
				}
			}
		}
	} else {
		selected = append(selected, ordered...)
	}

	if len(normalizedExclude) == 0 || len(selected) == 0 {
		return selected
	}

	filtered := make([]Instance, 0, len(selected))
	for _, instance := range selected {
		if matchesAny(instance, normalizedExclude) {
			continue
		}
		filtered = append(filtered, instance)
	}

	return filtered
}

func MatchesPattern(provider Instance, pattern string) bool {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return false
	}
	if pattern == string(provider.Kind) || pattern == string(provider.Kind)+":*" {
		return true
	}
	if strings.HasSuffix(pattern, ":*") {
		return strings.HasPrefix(provider.ID, strings.TrimSuffix(pattern, "*"))
	}
	return provider.ID == pattern
}

func matchesAny(provider Instance, patterns []string) bool {
	for _, pattern := range patterns {
		if MatchesPattern(provider, pattern) {
			return true
		}
	}
	return false
}

func normalizePatterns(patterns []string) []string {
	if len(patterns) == 0 {
		return nil
	}
	out := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		out = append(out, pattern)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
