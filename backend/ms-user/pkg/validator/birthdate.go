package validator

import (
	"strings"
	"time"

	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/wi-shared/exception"
)

// ParseBirthDate parses YYYY-MM-DD into UTC midnight.
func ParseBirthDate(raw string) (*time.Time, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "birth_date"})
	}
	parsed, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "birth_date"})
	}
	utc := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC)

	return &utc, nil
}

// AgeYears returns full years since birth (UTC, approximate via YearDay).
func AgeYears(birth time.Time) int {
	now := time.Now().UTC()
	years := now.Year() - birth.Year()
	if now.YearDay() < birth.YearDay() {
		years--
	}

	return years
}

// MeetsMinAge reports whether birth satisfies MinAgeYears.
func MeetsMinAge(birth time.Time) bool {
	return AgeYears(birth) >= constants.MinAgeYears
}
