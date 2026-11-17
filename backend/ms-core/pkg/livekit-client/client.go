package livekitclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	"github.com/livekit/protocol/webhook"
	"github.com/twitchtv/twirp"
)

// TokenGrants specifies LiveKit participant JWT grants.
type TokenGrants struct {
	RoomName        string
	Identity        string
	Name            string
	Metadata        string
	TTL             time.Duration
	CanPublish      bool
	CanSubscribe    bool
	CanPublishAudio bool
	CanPublishVideo bool
}

// ParticipantModeration updates a participant's SFU permissions.
type ParticipantModeration struct {
	RoomName     string
	Identity     string
	CanPublish   bool
	CanSubscribe bool
	AudioMuted   *bool
}

// EgressS3Config holds S3/MinIO destination settings for room composite egress.
type EgressS3Config struct {
	Endpoint string
	Bucket   string
	Region   string
	Key      string
	Secret   string
	Filepath string
}

// Client abstracts LiveKit JWT signing, RoomService, Egress, and Webhook verification.
type Client interface {
	IssueToken(grants TokenGrants) (string, time.Time, error)
	VerifyWebhook(authHeader string, rawBody []byte) (*livekit.WebhookEvent, error)
	UpdateParticipantPermissions(ctx context.Context, mod ParticipantModeration) error
	MuteAudioTracks(ctx context.Context, roomName, identity string, muted bool) error
	MuteVideoTracks(ctx context.Context, roomName, identity string, muted bool) error
	RemoveParticipant(ctx context.Context, roomName, identity string) error
	ListParticipants(ctx context.Context, roomName string) ([]*livekit.ParticipantInfo, error)
	StartRoomCompositeEgress(ctx context.Context, roomName string, s3 EgressS3Config) (egressID string, err error)
	StopEgress(ctx context.Context, egressID string) error
}

type sdkClient struct {
	url       string
	apiKey    string
	apiSecret string
	rooms     livekit.RoomService
	egress    livekit.Egress
}

// New constructs a LiveKit protocol client wrapper.
func New(url, apiKey, apiSecret string) Client {
	httpURL := toHTTPURL(url)
	var rooms livekit.RoomService
	var egress livekit.Egress
	if httpURL != "" && apiKey != "" && apiSecret != "" {
		httpClient := &http.Client{Timeout: 10 * time.Second}
		rooms = livekit.NewRoomServiceProtobufClient(httpURL, httpClient)
		egress = livekit.NewEgressProtobufClient(httpURL, httpClient)
	}

	return &sdkClient{
		url:       strings.TrimSpace(url),
		apiKey:    strings.TrimSpace(apiKey),
		apiSecret: strings.TrimSpace(apiSecret),
		rooms:     rooms,
		egress:    egress,
	}
}

func (c *sdkClient) adminContext(ctx context.Context, roomName string) (context.Context, error) {
	at := auth.NewAccessToken(c.apiKey, c.apiSecret).
		SetValidFor(2 * time.Minute).
		SetVideoGrant(&auth.VideoGrant{
			RoomAdmin:  true,
			RoomList:   true,
			RoomRecord: true,
			Room:       roomName,
		})
	jwt, err := at.ToJWT()
	if err != nil {
		return nil, err
	}
	return twirp.WithHTTPRequestHeaders(ctx, http.Header{
		"Authorization": []string{"Bearer " + jwt},
	})
}

func toHTTPURL(raw string) string {
	u := strings.TrimSpace(raw)
	if strings.HasPrefix(u, "wss://") {
		return "https://" + strings.TrimPrefix(u, "wss://")
	}
	if strings.HasPrefix(u, "ws://") {
		return "http://" + strings.TrimPrefix(u, "ws://")
	}
	return u
}

func (c *sdkClient) IssueToken(grants TokenGrants) (string, time.Time, error) {
	if c.apiKey == "" || c.apiSecret == "" {
		return "", time.Time{}, errors.New("livekit credentials are not configured")
	}

	ttl := grants.TTL
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	expiresAt := time.Now().UTC().Add(ttl)

	var sources []string
	if grants.CanPublishAudio {
		sources = append(sources, "microphone")
	}
	if grants.CanPublishVideo {
		sources = append(sources, "camera", "screen_share", "screen_share_audio")
	}

	canPublish := grants.CanPublish && len(sources) > 0
	canSubscribe := grants.CanSubscribe
	videoGrant := &auth.VideoGrant{
		RoomJoin:          true,
		Room:              grants.RoomName,
		CanPublish:        &canPublish,
		CanSubscribe:      &canSubscribe,
		CanPublishSources: sources,
	}

	at := auth.NewAccessToken(c.apiKey, c.apiSecret).
		SetIdentity(grants.Identity).
		SetName(grants.Name).
		SetMetadata(grants.Metadata).
		SetValidFor(ttl).
		SetVideoGrant(videoGrant)

	jwt, err := at.ToJWT()
	if err != nil {
		return "", time.Time{}, err
	}

	return jwt, expiresAt, nil
}

func (c *sdkClient) VerifyWebhook(authHeader string, rawBody []byte) (*livekit.WebhookEvent, error) {
	if c.apiKey == "" || c.apiSecret == "" {
		return nil, errors.New("livekit credentials are not configured")
	}
	req, err := http.NewRequest(http.MethodPost, "/api/v1/webhooks/livekit", bytes.NewReader(rawBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/webhook+json")

	provider := auth.NewSimpleKeyProvider(c.apiKey, c.apiSecret)
	return webhook.ReceiveWebhookEvent(req, provider)
}

func (c *sdkClient) UpdateParticipantPermissions(ctx context.Context, mod ParticipantModeration) error {
	if c.rooms == nil {
		return nil
	}
	actx, err := c.adminContext(ctx, mod.RoomName)
	if err != nil {
		return err
	}
	_, err = c.rooms.UpdateParticipant(actx, &livekit.UpdateParticipantRequest{
		Room:     mod.RoomName,
		Identity: mod.Identity,
		Permission: &livekit.ParticipantPermission{
			CanPublish:   mod.CanPublish,
			CanSubscribe: mod.CanSubscribe,
		},
	})
	return err
}

func (c *sdkClient) MuteAudioTracks(ctx context.Context, roomName, identity string, muted bool) error {
	return c.muteTracksByType(ctx, roomName, identity, livekit.TrackType_AUDIO, muted)
}

func (c *sdkClient) MuteVideoTracks(ctx context.Context, roomName, identity string, muted bool) error {
	return c.muteTracksByType(ctx, roomName, identity, livekit.TrackType_VIDEO, muted)
}

func (c *sdkClient) muteTracksByType(ctx context.Context, roomName, identity string, kind livekit.TrackType, muted bool) error {
	if c.rooms == nil {
		return nil
	}
	actx, err := c.adminContext(ctx, roomName)
	if err != nil {
		return err
	}
	p, err := c.rooms.GetParticipant(actx, &livekit.RoomParticipantIdentity{
		Room:     roomName,
		Identity: identity,
	})
	if err != nil {
		return err
	}
	for _, tr := range p.GetTracks() {
		if tr.GetType() == kind {
			_, _ = c.rooms.MutePublishedTrack(actx, &livekit.MuteRoomTrackRequest{
				Room:     roomName,
				Identity: identity,
				TrackSid: tr.GetSid(),
				Muted:    muted,
			})
		}
	}
	return nil
}

func (c *sdkClient) RemoveParticipant(ctx context.Context, roomName, identity string) error {
	if c.rooms == nil {
		return nil
	}
	actx, err := c.adminContext(ctx, roomName)
	if err != nil {
		return err
	}
	_, err = c.rooms.RemoveParticipant(actx, &livekit.RoomParticipantIdentity{
		Room:     roomName,
		Identity: identity,
	})
	return err
}

func (c *sdkClient) ListParticipants(ctx context.Context, roomName string) ([]*livekit.ParticipantInfo, error) {
	if c.rooms == nil {
		return nil, nil
	}
	actx, err := c.adminContext(ctx, roomName)
	if err != nil {
		return nil, err
	}
	res, err := c.rooms.ListParticipants(actx, &livekit.ListParticipantsRequest{Room: roomName})
	if err != nil {
		return nil, err
	}
	return res.GetParticipants(), nil
}

func (c *sdkClient) StartRoomCompositeEgress(ctx context.Context, roomName string, s3 EgressS3Config) (string, error) {
	if c.egress == nil {
		return "", errors.New("livekit egress client unavailable")
	}
	actx, err := c.adminContext(ctx, roomName)
	if err != nil {
		return "", err
	}
	region := s3.Region
	if region == "" {
		region = "us-east-1"
	}
	res, err := c.egress.StartRoomCompositeEgress(actx, &livekit.RoomCompositeEgressRequest{
		RoomName: roomName,
		Layout:   "grid",
		FileOutputs: []*livekit.EncodedFileOutput{
			{
				FileType: livekit.EncodedFileType_MP4,
				Filepath: s3.Filepath,
				Output: &livekit.EncodedFileOutput_S3{
					S3: &livekit.S3Upload{
						AccessKey:      s3.Key,
						Secret:         s3.Secret,
						Region:         region,
						Endpoint:       s3.Endpoint,
						Bucket:         s3.Bucket,
						ForcePathStyle: true,
					},
				},
			},
		},
	})
	if err != nil {
		return "", err
	}
	return res.GetEgressId(), nil
}

func (c *sdkClient) StopEgress(ctx context.Context, egressID string) error {
	if c.egress == nil {
		return errors.New("livekit egress client unavailable")
	}
	actx, err := c.adminContext(ctx, "")
	if err != nil {
		return err
	}
	_, err = c.egress.StopEgress(actx, &livekit.StopEgressRequest{EgressId: egressID})
	return err
}

// ParticipantMetadataJSON encodes participant metadata for LiveKit JWT.
func ParticipantMetadataJSON(displayName, username, avatarURL string) string {
	b, _ := json.Marshal(struct {
		DisplayName string `json:"display_name"`
		Username    string `json:"username"`
		AvatarURL   string `json:"avatar_url,omitempty"`
	}{
		DisplayName: displayName,
		Username:    username,
		AvatarURL:   avatarURL,
	})
	return string(b)
}
