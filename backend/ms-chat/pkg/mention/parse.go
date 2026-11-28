package mention

import (
	"regexp"
	"strings"
)

const (
	mentionUserIDCapture = `([0-9a-f]{24}|[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12})`
	mentionRoleIDCapture = `([0-9a-f]{24})`
)

var (
	mentionPublicIDRe     = regexp.MustCompile(`\[@([^\]]+)\]\(wichat-mention:` + mentionUserIDCapture + `\)`)
	mentionRolePublicIDRe = regexp.MustCompile(`\[@([^\]]+)\]\(wichat-mention-role:` + mentionRoleIDCapture + `\)`)
	mentionPreviewStripRe = regexp.MustCompile(`\[@([^\]]+)\]\(wichat-mention(?:-role)?:[^)]+\)`)
)

// PublicUserIDsFromBody dedupes user public ids from stored message markdown.
func PublicUserIDsFromBody(body string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, m := range mentionPublicIDRe.FindAllStringSubmatch(body, -1) {
		if len(m) < 3 {
			continue
		}
		id := strings.TrimSpace(m[2])
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

// StripCallerPublicID removes the author's public user id from mention targets.
func StripCallerPublicID(ids []string, callerPublicID string) []string {
	callerPublicID = strings.TrimSpace(callerPublicID)
	if callerPublicID == "" || len(ids) == 0 {
		return ids
	}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if strings.TrimSpace(id) != callerPublicID {
			out = append(out, id)
		}
	}
	return out
}

// MergePublicUserIDs combines explicit ids with body-derived ids.
func MergePublicUserIDs(explicit []string, body string) []string {
	seen := make(map[string]struct{})
	var out []string
	add := func(id string) {
		id = strings.TrimSpace(id)
		if id == "" {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	for _, id := range explicit {
		add(id)
	}
	for _, id := range PublicUserIDsFromBody(body) {
		add(id)
	}
	return out
}

// PublicRoleIDsFromBody dedupes role public ids from stored message markdown.
func PublicRoleIDsFromBody(body string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, m := range mentionRolePublicIDRe.FindAllStringSubmatch(body, -1) {
		if len(m) < 3 {
			continue
		}
		id := strings.TrimSpace(m[2])
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

// UnionPublicUserIDs merges mention target user ids with stable dedupe order.
func UnionPublicUserIDs(lists ...[]string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, list := range lists {
		for _, id := range list {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			out = append(out, id)
		}
	}
	return out
}

// StripTokensForPreview replaces stored mention markdown with plain @label for snippets.
func StripTokensForPreview(body string) string {
	return mentionPreviewStripRe.ReplaceAllString(body, "@$1")
}
