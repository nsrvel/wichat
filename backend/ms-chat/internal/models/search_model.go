package models

type MessageSearchHit struct {
	MessageID         string
	GroupID           string
	ChannelID         string
	BodyPreview       string
	AuthorDisplayName string
	CreatedAt         string
}

type FileSearchHit struct {
	AttachmentID string
	MessageID    string
	GroupID      string
	ChannelID    string
	Filename     string
	ContentType  string
	CreatedAt    string
}
