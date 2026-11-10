package authPb

import (
	"time"

	"github.com/wichat/wichat/backend/ms-user/internal/models"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

// GroupMemberEntries maps member profiles to protobuf.
func GroupMemberEntries(members []models.GroupMemberProfile) []*userv1.GroupMemberEntry {
	out := make([]*userv1.GroupMemberEntry, len(members))
	for i, m := range members {
		out[i] = &userv1.GroupMemberEntry{
			UserId:      m.UserID,
			DisplayName: m.DisplayName,
			Email:       m.Email,
			Role:        m.Role,
			JoinedAt:    m.JoinedAt.UTC().Format(time.RFC3339Nano),
		}
	}

	return out
}

// UserSearchHits maps search rows to protobuf.
func UserSearchHits(users []models.UserSearchHit) []*userv1.UserSearchHit {
	out := make([]*userv1.UserSearchHit, len(users))
	for i, u := range users {
		out[i] = &userv1.UserSearchHit{
			UserId:      u.UserID,
			DisplayName: u.DisplayName,
			Email:       u.Email,
			Username:    u.Username,
		}
	}

	return out
}
