package objecturl

import "strings"

// Public returns an absolute URL when publicBase is set; otherwise a root-relative path clients can resolve.
func Public(stored, publicBase string) string {
	stored = strings.TrimSpace(stored)
	if stored == "" {
		return ""
	}

	lower := strings.ToLower(stored)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return stored
	}

	path := stored
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	base := strings.TrimRight(strings.TrimSpace(publicBase), "/")
	if base == "" {
		return path
	}

	return base + path
}
