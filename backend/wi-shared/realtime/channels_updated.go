package realtime

import "encoding/json"

const TypeChannelsUpdated = "channels.updated"

const ChannelsUpdateReasonDmCreated = "dm.created"

// ChannelsUpdatedEvent notifies group subscribers to refresh channel metadata (sidebar).
type ChannelsUpdatedEvent struct {
	Type      string `json:"type"`
	GroupID   string `json:"group_id"`
	Reason    string `json:"reason,omitempty"`
	ChannelID string `json:"channel_id,omitempty"`
}

// NewChannelsUpdated builds a group fanout payload for channel list changes.
func NewChannelsUpdated(groupPublicID, reason, channelPublicID string) ChannelsUpdatedEvent {
	return ChannelsUpdatedEvent{
		Type:      TypeChannelsUpdated,
		GroupID:   groupPublicID,
		Reason:    reason,
		ChannelID: channelPublicID,
	}
}

// MarshalChannelsUpdated JSON-encodes a channels.updated frame.
func MarshalChannelsUpdated(groupPublicID, reason, channelPublicID string) ([]byte, error) {
	return json.Marshal(NewChannelsUpdated(groupPublicID, reason, channelPublicID))
}
