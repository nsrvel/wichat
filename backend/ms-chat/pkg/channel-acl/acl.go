package channelAcl

import (
	"context"

	"github.com/wichat/wichat/backend/wi-shared/exception"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

// ChannelLister loads visible channels for ACL filtering.
type ChannelLister interface {
	ListGroupChannels(ctx context.Context, in *userv1.ListGroupChannelsRequest) (*userv1.ListGroupChannelsResponse, error)
}

// AllowedIDs returns channel public IDs the bearer may access in the group.
func AllowedIDs(ctx context.Context, lister ChannelLister, accessToken, groupID string) (map[string]struct{}, error) {
	if lister == nil {
		return nil, exception.Internal(nil)
	}
	out, err := lister.ListGroupChannels(ctx, &userv1.ListGroupChannelsRequest{
		AccessToken: accessToken,
		GroupId:     groupID,
	})
	if err != nil {
		return nil, err
	}
	allowed := make(map[string]struct{}, len(out.GetChannels()))
	for _, ch := range out.GetChannels() {
		if ch.GetId() == "" {
			continue
		}
		allowed[ch.GetId()] = struct{}{}
	}
	return allowed, nil
}
