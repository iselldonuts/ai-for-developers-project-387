package dto

type ListEventTypesFilter struct {
	IncludeArchived bool
}

type CreateEventTypeInput struct {
	Title           string
	Description     string
	DurationMinutes int
}
