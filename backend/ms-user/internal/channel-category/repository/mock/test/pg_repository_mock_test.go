package mock_test

import (
	"context"
	"testing"
	"time"

	categorymock "github.com/wichat/wichat/backend/ms-user/internal/channel-category/repository/mock"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
)

func TestPgRepositoryMockCategoryCRUDAndPermissions(t *testing.T) {
	t.Parallel()

	st := usermock.NewState()
	repo := categorymock.NewPgRepositoryMock(st)
	ctx := context.Background()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	cat, err := repo.CreateCategory(ctx, "g1", "Engineering", false, "u1", now)
	if err != nil || cat.Name != "Engineering" {
		t.Fatalf("CreateCategory: %+v err=%v", cat, err)
	}

	cats, err := repo.ListCategoriesByGroupID(ctx, "g1")
	if err != nil || len(cats) != 1 || cats[0].ID != cat.ID {
		t.Fatalf("ListCategoriesByGroupID: %+v err=%v", cats, err)
	}

	newName := "Core Eng"
	isPriv := true
	updated, err := repo.UpdateCategory(ctx, "g1", cat.PublicID, &newName, &isPriv, now.Add(time.Minute))
	if err != nil || updated.Name != "Core Eng" || !updated.IsPrivate {
		t.Fatalf("UpdateCategory: %+v err=%v", updated, err)
	}

	if _, err := repo.UpsertCategoryPermissionOverride(ctx, "g1", cat.PublicID, "role", "r1", []string{"view_channels"}, nil, now); err != nil {
		t.Fatalf("UpsertCategoryPermissionOverride: %v", err)
	}
	all, err := repo.ListGroupCategoryPermissionOverrides(ctx, "g1")
	if err != nil || len(all) != 1 {
		t.Fatalf("ListGroupCategoryPermissionOverrides: %+v err=%v", all, err)
	}
	if err := repo.DeleteCategoryPermissionOverride(ctx, "g1", cat.PublicID, "role", "r1"); err != nil {
		t.Fatalf("DeleteCategoryPermissionOverride: %v", err)
	}

	if err := repo.SetCategoryPositions(ctx, "g1", []string{cat.ID}); err != nil {
		t.Fatalf("SetCategoryPositions: %v", err)
	}
	if err := repo.DeleteCategory(ctx, "g1", cat.PublicID); err != nil {
		t.Fatalf("DeleteCategory: %v", err)
	}
}
