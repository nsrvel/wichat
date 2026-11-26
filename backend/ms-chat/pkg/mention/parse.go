package mention

import (
	"regexp"
	"strings"
)

const mentionUserIDCapture = `([0-9a-f]{24}|[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12})`

var (
	mentionPublicIDRe     = regexp.MustCompile(`\[@([^\]]+)\]\(wichat-mention:` + mentionUserIDCapture + `\)`)
	mentionPreviewStripRe = regexp.MustCompile(`\[@([^\]]+)\]\(wichat-mention:[^)]+\)`)
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

// StripTokensForPreview replaces stored mention markdown with plain @label for snippets.
func StripTokensForPreview(body string) string {
	return mentionPreviewStripRe.ReplaceAllString(body, "@$1")
}
