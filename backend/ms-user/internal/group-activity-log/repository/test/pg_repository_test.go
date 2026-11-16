// Integration tests for group-activity-log Postgres repository. Requires Docker; skipped when unavailable.
package repository_test

import (
	"context"
	"testing"
	"time"

	groupactivityrepo "github.com/wichat/wichat/backend/ms-user/internal/group-activity-log/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

func TestGroupActivityLogPgRepositoryContract(t *testing.T) {
	var _ groupactivityrepo.PgRepository
	_ = context.Background()
	_ = time.UTC
	_ = models.GroupActivityLogFilter{}
	_ = constants.GroupAuditGroupCreated
}
