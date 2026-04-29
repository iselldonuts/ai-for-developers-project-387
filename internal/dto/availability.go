package dto

import "github.com/iselldonuts/ai-for-developers-project-386/internal/model"

type ReplaceAvailabilityInput struct {
	Windows []model.AvailabilityWindow
}
