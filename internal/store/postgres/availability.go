package postgres

import (
	"context"
	"time"

	"github.com/iselldonuts/ai-for-developers-project-386/internal/model"
	"github.com/iselldonuts/ai-for-developers-project-386/internal/platform/postgres/txmanager"
)

type AvailabilityStore struct {
	txManager txmanager.Manager
}

func NewAvailabilityStore(txManager txmanager.Manager) AvailabilityStore {
	return AvailabilityStore{txManager: txManager}
}

func (s AvailabilityStore) ListAvailabilityWindows(ctx context.Context, ownerID string) ([]model.AvailabilityWindow, error) {
	querier := s.txManager.GetQuerier(ctx)

	rows, err := querier.Query(ctx, `
		select day_of_week, start_time, end_time
		from availability_windows
		where owner_id = $1
		order by
			case day_of_week
				when 'monday' then 1
				when 'tuesday' then 2
				when 'wednesday' then 3
				when 'thursday' then 4
				when 'friday' then 5
				when 'saturday' then 6
				when 'sunday' then 7
			end
	`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	windows := make([]model.AvailabilityWindow, 0)

	for rows.Next() {
		var dayOfWeek string
		var startTime time.Time
		var endTime time.Time

		if err := rows.Scan(&dayOfWeek, &startTime, &endTime); err != nil {
			return nil, err
		}

		windows = append(windows, model.AvailabilityWindow{
			DayOfWeek: model.DayOfWeek(dayOfWeek),
			StartTime: startTime.Format("15:04"),
			EndTime:   endTime.Format("15:04"),
		})
	}

	return windows, rows.Err()
}

func (s AvailabilityStore) ReplaceAvailabilityWindows(ctx context.Context, ownerID string, windows []model.AvailabilityWindow) error {
	querier := s.txManager.GetQuerier(ctx)

	if _, err := querier.Exec(ctx, `delete from availability_windows where owner_id = $1`, ownerID); err != nil {
		return err
	}

	for _, window := range windows {
		startTime, err := time.Parse("15:04", window.StartTime)
		if err != nil {
			return err
		}

		endTime, err := time.Parse("15:04", window.EndTime)
		if err != nil {
			return err
		}

		if _, err := querier.Exec(ctx, `
			insert into availability_windows (owner_id, day_of_week, start_time, end_time)
			values ($1, $2, $3, $4)
		`, ownerID, string(window.DayOfWeek), startTime, endTime); err != nil {
			return err
		}
	}

	return nil
}
