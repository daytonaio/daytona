package workspace

import "fmt"

// SelectWorkspace filters candidate workspaces and returns matching items or an error
func SelectWorkspace(workspaces []string, query string) ([]string, error) {
	filtered := FilterList(workspaces, query)
	if len(filtered) == 0 {
		return nil, fmt.Errorf("no workspace matching filter '%s'", query)
	}
	return filtered, nil
}
