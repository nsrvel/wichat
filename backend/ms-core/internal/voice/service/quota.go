package service

import (
	"context"
	"time"

	"github.com/wichat/wichat/backend/ms-core/internal/models"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

func (s service) computeQuotaUsage(ctx context.Context) (models.QuotaUsageResult, error) {
	now := s.now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	nextMonth := monthStart.AddDate(0, 1, 0)

	totalSeconds, err := s.pg.SumParticipantSecondsBetween(ctx, monthStart, nextMonth)
	if err != nil {
		return models.QuotaUsageResult{}, err
	}

	usedMins := totalSeconds / 60
	limit := s.opts.MonthlyQuotaMins
	if limit <= 0 {
		limit = DefaultMonthlyQuotaMins
	}
	pct := (float64(usedMins) / float64(limit)) * 100.0
	return models.QuotaUsageResult{
		Month:                   monthStart.Format("2006-01"),
		UsedParticipantMinutes:  usedMins,
		LimitParticipantMinutes: limit,
		UsagePercent:            pct,
		WarningActive:           pct >= QuotaWarningThresholdPerc,
	}, nil
}

func (s service) GetQuotaUsage(ctx context.Context, accessToken, groupPublicID string) (models.QuotaUsageResult, error) {
	if accessToken == "" {
		return models.QuotaUsageResult{}, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}
	if groupPublicID != "" && s.userChannel != nil {
		if _, err := s.userChannel.ListGroupChannels(ctx, &userv1.ListGroupChannelsRequest{
			AccessToken: accessToken,
			GroupId:     groupPublicID,
		}); err != nil {
			return models.QuotaUsageResult{}, err
		}
	}
	return s.computeQuotaUsage(ctx)
}
