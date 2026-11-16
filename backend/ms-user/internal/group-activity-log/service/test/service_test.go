package service_test

import (
	"context"
	"testing"
	"time"

	groupactivitymock "github.com/wichat/wichat/backend/ms-user/internal/group-activity-log/repository/mock"
	groupactivitysvc "github.com/wichat/wichat/backend/ms-user/internal/group-activity-log/service"
	groupmembermock "github.com/wichat/wichat/backend/ms-user/internal/group-member/repository/mock"
	grouprolemock "github.com/wichat/wichat/backend/ms-user/internal/group-role/repository/mock"
	groupmock "github.com/wichat/wichat/backend/ms-user/internal/group/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	testSupport "github.com/wichat/wichat/backend/ms-user/pkg/test-support"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

func TestRecordAndListGroupActivityLogs(t *testing.T) {
	ctx := context.Background()
	st := usermock.NewState()
	authSvc, notify, _, _, _, _ := testSupport.NewAuthService(t, st, nil)
	_, _, sessions := testSupport.TestRepos(st)

	owner := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email:       "audit-owner@example.com",
		DisplayName: "Audit Owner",
		Username:    "audit.owner",
		Password:    "strong-password-123",
		BirthDate:   "1998-01-01",
		Locale:      "en",
	})
	member := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email:       "audit-member@example.com",
		DisplayName: "Audit Member",
		Username:    "audit.member",
		Password:    "strong-password-123",
		BirthDate:   "1998-01-01",
		Locale:      "en",
	})

	groupRepo := groupmock.NewPgRepositoryMock(st)
	memberRepo := groupmembermock.NewPgRepositoryMock(st)
	roleRepo := grouprolemock.NewPgRepositoryMock(st)

	group := models.Group{
		ID:        "g-audit-1",
		PublicID:  "111111111111111111111111",
		Name:      "Audit Group",
		CreatedAt: time.Now().UTC(),
	}
	_ = groupRepo.CreateGroupWithOwner(ctx, group, models.GroupMember{
		GroupID: group.ID, UserID: owner.User.ID, Role: constants.GroupRoleOwner, CreatedAt: time.Now().UTC(),
	})
	_ = memberRepo.InsertGroupMember(ctx, models.GroupMember{
		GroupID: group.ID, UserID: member.User.ID, Role: constants.GroupRoleMember, CreatedAt: time.Now().UTC(),
	})

	actRepo := groupactivitymock.NewPgRepositoryMock()
	log := logger.New(logger.Options{Service: "group-activity-test", Env: "test", Level: "error"})
	svc := groupactivitysvc.NewService(actRepo, groupRepo, memberRepo, roleRepo, sessions, []byte(testSupport.TestJWTSecret), log)

	base := time.Now().UTC().Add(-5 * time.Minute)
	svc.Record(ctx, models.GroupActivityLogEntry{
		GroupID:    group.ID,
		ActorID:    owner.User.ID,
		Category:   constants.GroupAuditCategoryProfile,
		Action:     constants.GroupAuditGroupProfileUpdated,
		TargetType: constants.GroupAuditTargetGroup,
		TargetID:   group.PublicID,
		TargetName: "Audit Group",
		Changes:    map[string]models.ActivityFieldChange{"name": {Old: "Old", New: "Audit Group"}},
		CreatedAt:  base,
	})
	svc.Record(ctx, models.GroupActivityLogEntry{
		GroupID:    group.ID,
		ActorID:    owner.User.ID,
		Category:   constants.GroupAuditCategoryModeration,
		Action:     constants.GroupAuditMemberKicked,
		TargetType: constants.GroupAuditTargetMember,
		TargetID:   member.User.ID,
		TargetName: member.User.DisplayName,
		Metadata:   map[string]any{"reason": "Spam"},
		CreatedAt:  base.Add(time.Minute),
	})

	// Non-privileged member is denied
	if _, err := svc.ListGroupActivityLogs(ctx, member.AccessToken, group.PublicID, "", "", "", 10, ""); err == nil {
		t.Fatal("expected non-privileged member to be denied")
	}

	// Owner can paginate and filter
	page1, err := svc.ListGroupActivityLogs(ctx, owner.AccessToken, group.PublicID, "", "", "", 1, "")
	if err != nil || len(page1.Logs) != 1 || page1.Logs[0].Action != constants.GroupAuditMemberKicked || page1.NextCursor == "" {
		t.Fatalf("unexpected page1: %+v err=%v", page1, err)
	}

	page2, err := svc.ListGroupActivityLogs(ctx, owner.AccessToken, group.PublicID, "", "", "", 1, page1.NextCursor)
	if err != nil || len(page2.Logs) != 1 || page2.Logs[0].Action != constants.GroupAuditGroupProfileUpdated || page2.NextCursor != "" {
		t.Fatalf("unexpected page2: %+v err=%v", page2, err)
	}

	// Invalid cursor returns error
	if _, err := svc.ListGroupActivityLogs(ctx, owner.AccessToken, group.PublicID, "", "", "", 10, "not-valid-base64!"); err == nil {
		t.Fatal("expected error for invalid cursor")
	}
}
