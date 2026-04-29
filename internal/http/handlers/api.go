package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/iselldonuts/ai-for-developers-project-386/internal/dto"
	"github.com/iselldonuts/ai-for-developers-project-386/internal/errs"
	httpapi "github.com/iselldonuts/ai-for-developers-project-386/internal/http/api"
	"github.com/iselldonuts/ai-for-developers-project-386/internal/model"
	"github.com/iselldonuts/ai-for-developers-project-386/internal/service"
)

type APIHandler struct {
	availabilityService service.AvailabilityService
	bookingService      service.BookingService
	eventTypeService    service.EventTypeService
}

func NewAPIHandler(
	availabilityService service.AvailabilityService,
	bookingService service.BookingService,
	eventTypeService service.EventTypeService,
) APIHandler {
	return APIHandler{
		availabilityService: availabilityService,
		bookingService:      bookingService,
		eventTypeService:    eventTypeService,
	}
}

func (h APIHandler) PublicBookingsApiCreate(w http.ResponseWriter, r *http.Request) {
	var requestBody httpapi.CreateBookingRequest

	if err := decodeJSON(r, &requestBody); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	item, err := h.bookingService.Create(r.Context(), dto.CreateBookingInput{
		EventTypeID: requestBody.EventTypeId,
		GuestName:   requestBody.GuestName,
		GuestEmail:  requestBody.GuestEmail,
		StartsAt:    requestBody.StartsAt,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, httpapi.CreatedBookingResponse{
		Booking: mapBooking(item),
	})
}

func (h APIHandler) PublicEventTypesApiList(w http.ResponseWriter, r *http.Request) {
	items, err := h.eventTypeService.List(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, httpapi.EventTypeListResponse{
		Items: mapEventTypes(items),
	})
}

func (h APIHandler) PublicEventTypesApiListSlots(w http.ResponseWriter, r *http.Request, eventTypeID string, params httpapi.PublicEventTypesApiListSlotsParams) {
	items, err := h.bookingService.ListSlots(r.Context(), eventTypeID, params.From, params.To)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, httpapi.SlotListResponse{Items: mapSlots(items)})
}

func (h APIHandler) OwnerAvailabilityApiGetAvailability(w http.ResponseWriter, r *http.Request) {
	snapshot, err := h.availabilityService.Get(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, httpapi.AvailabilityResponse{
		Owner:   mapOwnerProfile(snapshot.Owner),
		Windows: mapAvailabilityWindows(snapshot.Windows),
	})
}

func (h APIHandler) OwnerAvailabilityApiUpdateAvailability(w http.ResponseWriter, r *http.Request) {
	var requestBody httpapi.UpdateAvailabilityRequest

	if err := decodeJSON(r, &requestBody); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	snapshot, err := h.availabilityService.Replace(r.Context(), dto.ReplaceAvailabilityInput{
		Windows: mapRequestAvailabilityWindows(requestBody.Windows),
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, httpapi.UpdatedAvailabilityResponse{
		Availability: httpapi.AvailabilityResponse{
			Owner:   mapOwnerProfile(snapshot.Owner),
			Windows: mapAvailabilityWindows(snapshot.Windows),
		},
	})
}

func (h APIHandler) OwnerBookingsApiListUpcoming(w http.ResponseWriter, r *http.Request) {
	result, err := h.bookingService.ListUpcoming(r.Context(), time.Now())
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, httpapi.UpcomingBookingsResponse{Items: mapUpcomingBookings(result.Items)})
}

func (h APIHandler) OwnerEventTypesApiList(w http.ResponseWriter, r *http.Request) {
	items, err := h.eventTypeService.List(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, httpapi.EventTypeListResponse{
		Items: mapEventTypes(items),
	})
}

func (h APIHandler) OwnerEventTypesApiCreate(w http.ResponseWriter, r *http.Request) {
	var requestBody httpapi.CreateEventTypeRequest

	if err := decodeJSON(r, &requestBody); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	item, err := h.eventTypeService.Create(r.Context(), dto.CreateEventTypeInput{
		Title:           requestBody.Title,
		Description:     requestBody.Description,
		DurationMinutes: int(requestBody.DurationMinutes),
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, httpapi.CreatedEventTypeResponse{
		EventType: mapEventType(item),
	})
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(target)
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeServiceError(w http.ResponseWriter, err error) {
	var validationErr errs.ValidationError

	switch {
	case errors.As(err, &validationErr):
		writeError(w, http.StatusBadRequest, "bad_request", validationErr.Error())
	case errors.Is(err, errs.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", err.Error())
	case errors.Is(err, errs.ErrConflict):
		writeError(w, http.StatusConflict, "conflict", "requested slot overlaps an existing booking")
	case errors.Is(err, errs.ErrNotImplemented):
		writeError(w, http.StatusNotImplemented, "not_implemented", err.Error())
	default:
		log.Printf("unexpected service error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func writeError(w http.ResponseWriter, statusCode int, code string, message string) {
	writeJSON(w, statusCode, map[string]httpapi.ErrorBody{
		"error": {
			Code:    code,
			Message: message,
		},
	})
}

func mapOwnerProfile(owner model.OwnerProfile) httpapi.OwnerProfile {
	return httpapi.OwnerProfile{
		Id:          owner.ID,
		DisplayName: owner.DisplayName,
		Timezone:    owner.Timezone,
	}
}

func mapAvailabilityWindows(windows []model.AvailabilityWindow) []httpapi.AvailabilityWindow {
	items := make([]httpapi.AvailabilityWindow, 0, len(windows))

	for _, window := range windows {
		items = append(items, httpapi.AvailabilityWindow{
			DayOfWeek: httpapi.AvailabilityWindowDayOfWeek(window.DayOfWeek),
			StartTime: window.StartTime,
			EndTime:   window.EndTime,
		})
	}

	return items
}

func mapRequestAvailabilityWindows(windows []httpapi.AvailabilityWindow) []model.AvailabilityWindow {
	items := make([]model.AvailabilityWindow, 0, len(windows))

	for _, window := range windows {
		items = append(items, model.AvailabilityWindow{
			DayOfWeek: model.DayOfWeek(window.DayOfWeek),
			StartTime: window.StartTime,
			EndTime:   window.EndTime,
		})
	}

	return items
}

func mapEventTypes(items []model.EventType) []httpapi.EventType {
	result := make([]httpapi.EventType, 0, len(items))

	for _, item := range items {
		result = append(result, mapEventType(item))
	}

	return result
}

func mapEventType(item model.EventType) httpapi.EventType {
	return httpapi.EventType{
		Id:              item.ID,
		Title:           item.Title,
		Description:     item.Description,
		DurationMinutes: int32(item.DurationMinutes),
	}
}

func mapSlots(items []model.Slot) []httpapi.Slot {
	result := make([]httpapi.Slot, 0, len(items))

	for _, item := range items {
		result = append(result, httpapi.Slot{
			EventTypeId:     item.EventTypeID,
			StartsAt:        item.StartsAt.UTC(),
			EndsAt:          item.EndsAt.UTC(),
			DurationMinutes: int32(item.DurationMinutes),
		})
	}

	return result
}

func mapBooking(item model.Booking) httpapi.Booking {
	return httpapi.Booking{
		Id:          item.ID,
		EventTypeId: item.EventTypeID,
		GuestName:   item.GuestName,
		GuestEmail:  item.GuestEmail,
		StartsAt:    item.StartsAt.UTC(),
		EndsAt:      item.EndsAt.UTC(),
	}
}

func mapUpcomingBookings(items []dto.BookingListItem) []httpapi.UpcomingBookingItem {
	result := make([]httpapi.UpcomingBookingItem, 0, len(items))

	for _, item := range items {
		result = append(result, httpapi.UpcomingBookingItem{
			Id:             item.BookingID,
			EventTypeId:    item.EventTypeID,
			EventTypeTitle: item.EventTypeTitle,
			GuestName:      item.GuestName,
			GuestEmail:     item.GuestEmail,
			StartsAt:       item.StartsAt.UTC(),
			EndsAt:         item.EndsAt.UTC(),
		})
	}

	return result
}
