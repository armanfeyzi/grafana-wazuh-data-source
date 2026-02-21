package models

import "strings"

func SanitizeStringList(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	out := make([]string, 0, len(values))
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			part = strings.TrimSpace(part)
			if part == "" || part == "$__all" || part == "All" || part == ".*" || strings.HasPrefix(part, "$") {
				continue
			}
			out = append(out, part)
		}
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
