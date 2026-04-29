package service

import (
	"testing"

	"github.com/iselldonuts/ai-for-developers-project-386/internal/model"
)

func TestValidateAvailabilityWindows(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		windows []model.AvailabilityWindow
		wantErr bool
	}{
		{
			name: "valid windows",
			windows: []model.AvailabilityWindow{
				{DayOfWeek: model.DayOfWeekMonday, StartTime: "09:00", EndTime: "17:00"},
				{DayOfWeek: model.DayOfWeekTuesday, StartTime: "10:00", EndTime: "18:00"},
			},
		},
		{
			name: "duplicate day",
			windows: []model.AvailabilityWindow{
				{DayOfWeek: model.DayOfWeekMonday, StartTime: "09:00", EndTime: "17:00"},
				{DayOfWeek: model.DayOfWeekMonday, StartTime: "10:00", EndTime: "18:00"},
			},
			wantErr: true,
		},
		{
			name: "invalid time range",
			windows: []model.AvailabilityWindow{
				{DayOfWeek: model.DayOfWeekMonday, StartTime: "17:00", EndTime: "09:00"},
			},
			wantErr: true,
		},
		{
			name: "invalid day",
			windows: []model.AvailabilityWindow{
				{DayOfWeek: model.DayOfWeek("holiday"), StartTime: "09:00", EndTime: "17:00"},
			},
			wantErr: true,
		},
	}

	for _, testCase := range tests {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := validateAvailabilityWindows(testCase.windows)
			if testCase.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}

			if !testCase.wantErr && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
		})
	}
}
