package constants

import "time"

const (
	DefaultMessageListLimit = 50
	DefaultMentionListLimit = 30
	MaxMessageListLimit     = 100
	MaxGapFillListLimit     = 500
	MaxMessageBodyRunes     = 4096
	MaxAttachmentBytes      = 25 << 20 // 25 MiB

	OutboxPublishBatchSize = 32
	OutboxPollInterval     = 250 * time.Millisecond
)
