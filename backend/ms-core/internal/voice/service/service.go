package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/wichat/wichat/backend/ms-core/internal/kafka/producers"
	"github.com/wichat/wichat/backend/ms-core/internal/models"
	voicerepo "github.com/wichat/wichat/backend/ms-core/internal/voice/repository"
	livekitclient "github.com/wichat/wichat/backend/ms-core/pkg/livekit-client"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	userauth "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/auth"
	userchannel "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/channel"
	userpresence "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/presence"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

const (
	MaxRoomParticipants       = 50
	MaxVideoPublishers        = 25
	TokenRateLimitPerMin      = 10
	DefaultTokenTTL           = 15 * time.Minute
	DefaultMonthlyQuotaMins   = 5000
	IdleWarningAfter          = 28 * time.Minute
	IdleKickAfter             = 30 * time.Minute
	QuotaWarningThresholdPerc = 80.0
)

// Options configures LiveKit and quota limits for voice service.
type Options struct {
	LiveKitURL       string
	MonthlyQuotaMins int64
	EnforceHardCap   bool
	MinioEndpoint    string
	MinioAccessKey   string
	MinioSecretKey   string
	MinioBucket      string
	MinioRegion      string
	MinioUseSSL      bool
}

// Service implements voice/video call use cases.
type Service interface {
	IssueVoiceToken(ctx context.Context, in *models.IssueVoiceTokenInput) (models.IssueVoiceTokenResult, error)
	ModerateParticipant(ctx context.Context, in *models.ModerateParticipantInput) (models.VoiceOccupant, error)
	HandleWebhook(ctx context.Context, authHeader string, rawBody []byte) error
	ListGroupOccupancy(ctx context.Context, accessToken, groupPublicID string) ([]models.ChannelOccupancy, error)
	StartRecording(ctx context.Context, accessToken, groupPublicID, channelPublicID string) (models.CallRecording, error)
	StopRecording(ctx context.Context, accessToken, groupPublicID, channelPublicID, recordingID string) (models.CallRecording, error)
	ListChannelRecordings(ctx context.Context, accessToken, groupPublicID, channelPublicID string, limit int) ([]models.CallRecording, error)
	AcknowledgeIdleWarning(ctx context.Context, accessToken, groupPublicID, channelPublicID string) error
	GetQuotaUsage(ctx context.Context, accessToken, groupPublicID string) (models.QuotaUsageResult, error)
	RunIdleCheck(ctx context.Context) error
}

type service struct {
	pg           voicerepo.PgRepository
	redis        voicerepo.RedisOccupancy
	minio        voicerepo.MinioRepository
	lk           livekitclient.Client
	userChannel  userchannel.Client
	userAuth     userauth.Client
	userPresence userpresence.Client
	events       producers.VoiceEventsPublisher
	opts         Options
	log          logger.Logger
	now          func() time.Time
}

// NewService constructs the voice service.
func NewService(
	pg voicerepo.PgRepository,
	redis voicerepo.RedisOccupancy,
	minio voicerepo.MinioRepository,
	lk livekitclient.Client,
	userChannel userchannel.Client,
	userAuth userauth.Client,
	userPresence userpresence.Client,
	events producers.VoiceEventsPublisher,
	opts Options,
	log logger.Logger,
) Service {
	if opts.MonthlyQuotaMins <= 0 {
		opts.MonthlyQuotaMins = DefaultMonthlyQuotaMins
	}
	return service{
		pg:           pg,
		redis:        redis,
		minio:        minio,
		lk:           lk,
		userChannel:  userChannel,
		userAuth:     userAuth,
		userPresence: userPresence,
		events:       events,
		opts:         opts,
		log:          log,
		now:          func() time.Time { return time.Now().UTC() },
	}
}

func hasAnyPerm(perms []string, targets ...string) bool {
	for _, t := range targets {
		if slices.Contains(perms, t) {
			return true
		}
	}
	return false
}

func isVoiceCapableChannel(channelType string) bool {
	switch channelType {
	case "voice_only", "text_and_voice", "dm":
		return true
	default:
		return false
	}
}

func roomNameForChannel(channelType, channelUUID string) string {
	if channelType == "dm" {
		return "dm:" + channelUUID
	}
	return "channel:" + channelUUID
}

func newPublicID24() string {
	var b [12]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func (s service) IssueVoiceToken(ctx context.Context, in *models.IssueVoiceTokenInput) (models.IssueVoiceTokenResult, error) {
	if strings.TrimSpace(in.AccessToken) == "" {
		return models.IssueVoiceTokenResult{}, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}

	acc, err := s.userChannel.AssertGroupChannelAccess(ctx, &userv1.AssertGroupChannelAccessRequest{
		AccessToken: in.AccessToken,
		GroupId:     in.GroupPublicID,
		ChannelId:   in.ChannelPublicID,
	})
	if err != nil {
		return models.IssueVoiceTokenResult{}, err
	}

	if !isVoiceCapableChannel(acc.GetChannelType()) {
		return models.IssueVoiceTokenResult{}, exception.InvalidInput(exception.MsgVoiceChannelNotVoice, map[string]any{"channel_type": acc.GetChannelType()})
	}

	perms := acc.GetPermissions()
	if len(perms) > 0 && !hasAnyPerm(perms, "administrator", "connect") {
		return models.IssueVoiceTokenResult{}, exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	userUUID := acc.GetUserUuid()
	channelUUID := acc.GetChannelUuid()

	if s.redis != nil {
		allowed, rlErr := s.redis.AllowTokenRequest(ctx, userUUID, TokenRateLimitPerMin)
		if rlErr != nil {
			return models.IssueVoiceTokenResult{}, rlErr
		}
		if !allowed {
			return models.IssueVoiceTokenResult{}, exception.RateLimited(exception.MsgRateLimited, nil)
		}
	}

	if s.opts.EnforceHardCap {
		quota, qErr := s.computeQuotaUsage(ctx)
		if qErr == nil && quota.UsedParticipantMinutes >= quota.LimitParticipantMinutes {
			return models.IssueVoiceTokenResult{}, exception.PermissionDenied(exception.MsgVoiceQuotaExceeded, nil)
		}
	}

	roomCap := MaxRoomParticipants
	if limit := int(acc.GetUserLimit()); limit > 0 && limit < MaxRoomParticipants {
		roomCap = limit
	}

	var occupants []models.VoiceOccupant
	var isRecording bool
	alreadyJoined := false
	if s.redis != nil {
		var occErr error
		occupants, occErr = s.redis.ListOccupants(ctx, channelUUID)
		if occErr != nil {
			return models.IssueVoiceTokenResult{}, occErr
		}
		for _, o := range occupants {
			if o.UserID == userUUID {
				alreadyJoined = true
				break
			}
		}
		if !alreadyJoined && len(occupants) >= roomCap {
			return models.IssueVoiceTokenResult{}, exception.Conflict(exception.MsgVoiceRoomFull, map[string]any{"limit": roomCap})
		}
		if meta, mErr := s.redis.GetRoomMeta(ctx, channelUUID); mErr == nil {
			isRecording = meta.IsRecording
		}
	}

	canSpeak := len(perms) == 0 || hasAnyPerm(perms, "administrator", "speak")
	canVideo := len(perms) == 0 || hasAnyPerm(perms, "administrator", "video")
	useVAD := len(perms) == 0 || hasAnyPerm(perms, "administrator", "use_voice_activity")

	displayName := userUUID
	username := ""
	avatarURL := ""
	if s.userAuth != nil {
		if me, meErr := s.userAuth.Me(ctx, &userv1.MeRequest{AccessToken: in.AccessToken}); meErr == nil && me.GetUser() != nil {
			u := me.GetUser()
			if u.GetDisplayName() != "" {
				displayName = u.GetDisplayName()
			}
			username = u.GetUsername()
			avatarURL = u.GetAvatarUrl()
		}
	}

	roomName := roomNameForChannel(acc.GetChannelType(), channelUUID)
	metadata := livekitclient.ParticipantMetadataJSON(displayName, username, avatarURL)
	token, expiresAt, err := s.lk.IssueToken(livekitclient.TokenGrants{
		RoomName:        roomName,
		Identity:        userUUID,
		Name:            displayName,
		Metadata:        metadata,
		TTL:             DefaultTokenTTL,
		CanPublish:      canSpeak || canVideo,
		CanSubscribe:    true,
		CanPublishAudio: canSpeak,
		CanPublishVideo: canVideo,
	})
	if err != nil {
		return models.IssueVoiceTokenResult{}, err
	}

	if s.redis != nil {
		_ = s.redis.SetRoomMeta(ctx, channelUUID, models.RoomMeta{
			GroupPublicID:   in.GroupPublicID,
			ChannelPublicID: in.ChannelPublicID,
			RoomName:        roomName,
			IsRecording:     isRecording,
		})
	}

	bitrate := acc.GetBitrateKbps()
	if bitrate <= 0 {
		bitrate = 64
	}
	videoQuality := acc.GetVideoQuality()
	if videoQuality == "" {
		videoQuality = "auto"
	}

	return models.IssueVoiceTokenResult{
		Token:              token,
		LiveKitURL:         s.opts.LiveKitURL,
		RoomName:           roomName,
		ChannelID:          in.ChannelPublicID,
		ExpiresAt:          expiresAt,
		InitialServerMuted: !canSpeak,
		CanPublishAudio:    canSpeak,
		CanPublishVideo:    canVideo,
		UseVoiceActivity:   useVAD,
		BitrateKbps:        bitrate,
		VideoQuality:       videoQuality,
		IsRecording:        isRecording,
		Occupants:          occupants,
	}, nil
}

func (s service) ListGroupOccupancy(ctx context.Context, accessToken, groupPublicID string) ([]models.ChannelOccupancy, error) {
	if strings.TrimSpace(accessToken) == "" {
		return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}
	if _, err := s.userChannel.ListGroupChannels(ctx, &userv1.ListGroupChannelsRequest{
		AccessToken: accessToken,
		GroupId:     groupPublicID,
	}); err != nil {
		return nil, err
	}
	if s.redis == nil {
		return nil, nil
	}

	channelUUIDs, err := s.redis.ListGroupChannelUUIDs(ctx, groupPublicID)
	if err != nil {
		return nil, err
	}
	out := make([]models.ChannelOccupancy, 0, len(channelUUIDs))
	for _, chUUID := range channelUUIDs {
		occs, err := s.redis.ListOccupants(ctx, chUUID)
		if err != nil || len(occs) == 0 {
			continue
		}
		meta, err := s.redis.GetRoomMeta(ctx, chUUID)
		if err != nil || meta.ChannelPublicID == "" {
			continue
		}
		out = append(out, models.ChannelOccupancy{
			ChannelID:   meta.ChannelPublicID,
			IsRecording: meta.IsRecording,
			Occupants:   occs,
		})
	}
	return out, nil
}

func (s service) AcknowledgeIdleWarning(ctx context.Context, accessToken, groupPublicID, channelPublicID string) error {
	acc, err := s.userChannel.AssertGroupChannelAccess(ctx, &userv1.AssertGroupChannelAccessRequest{
		AccessToken: accessToken,
		GroupId:     groupPublicID,
		ChannelId:   channelPublicID,
	})
	if err != nil {
		return err
	}
	if s.redis == nil {
		return nil
	}
	err = s.redis.TouchOccupantActivity(ctx, acc.GetChannelUuid(), acc.GetUserUuid(), s.now().Unix())
	if errors.Is(err, voicerepo.ErrNotFound) {
		return nil
	}
	return err
}
