package workspace

import "strings"

// FilterList filters selection choices by substring
func FilterList(items []string, query string) []string {
	if query == "" {
		return items
	}
	lowerQuery := strings.ToLower(query)
	var filtered []string
	for _, item := range items {
		if strings.Contains(strings.ToLower(item), lowerQuery) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// FilterWorkspaceNames filters candidate workspace choices by substring
func FilterWorkspaceNames(names []string, query string) []string {
	return FilterList(names, query)
}
