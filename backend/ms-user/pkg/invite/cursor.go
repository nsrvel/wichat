package invite

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/wichat/wichat/backend/ms-user/internal/models"
)

// EncodeMemberCursor serializes a member page cursor for APIs.
func EncodeMemberCursor(c *models.GroupMemberCursor) string {
	if c == nil {
		return ""
	}
	raw := fmt.Sprintf("%s|%s", c.CreatedAt.UTC().Format(time.RFC3339Nano), c.UserID)

	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

// DecodeMemberCursor parses a cursor from ListGroupMembers.
func DecodeMemberCursor(raw string) (*models.GroupMemberCursor, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, err
	}
	parts := strings.SplitN(string(b), "|", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid cursor")
	}
	ts, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return nil, err
	}

	return &models.GroupMemberCursor{CreatedAt: ts, UserID: parts[1]}, nil
}
