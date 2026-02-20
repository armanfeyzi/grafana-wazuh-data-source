package models

import "strings"

func SanitizeStringList(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || value == "$__all" || value == "All" || value == ".*" {
			continue
		}
		out = append(out, value)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func (f QueryFilters) AgentNamesForQuery() []string {
	return SanitizeStringList(f.AgentNames)
}

func (f QueryFilters) SeverityForQuery() []string {
	return SanitizeStringList(f.Severity)
}
