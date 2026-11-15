package validator

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/wi-shared/exception"
)

var hexColorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// RoleName trims and validates a group role name.
func RoleName(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	n := utf8.RuneCountInString(trimmed)
	if n < constants.MinRoleNameRunes || n > constants.MaxRoleNameRunes || hasControlOrInvisible(trimmed, false) {
		return "", exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "name"})
	}

	return trimmed, nil
}

// RoleColor validates an optional 6-digit hex color string.
func RoleColor(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !hexColorPattern.MatchString(trimmed) {
		return "", exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "color"})
	}

	return trimmed, nil
}

var knownPermissionSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(constants.AllKnownPermissions))
	for _, p := range constants.AllKnownPermissions {
		m[p] = struct{}{}
	}
	return m
}()

// RolePermissions validates and deduplicates a list of permission keys.
func RolePermissions(perms []string) ([]string, error) {
	if len(perms) == 0 {
		return []string{}, nil
	}

	seen := make(map[string]struct{}, len(perms))
	out := make([]string, 0, len(perms))
	for _, raw := range perms {
		p := strings.TrimSpace(raw)
		if _, ok := knownPermissionSet[p]; !ok {
			return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "permissions"})
		}
		if _, dup := seen[p]; dup {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out, nil
}
