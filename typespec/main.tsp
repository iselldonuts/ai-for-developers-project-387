import "@typespec/http";
import "@typespec/openapi3";

using TypeSpec.Http;
using TypeSpec.OpenAPI;

@service(#{ title: "CalCal API" })
@server("http://localhost:8080", "Local development")
namespace CalCal;

alias DayOfWeek =
  "monday" |
  "tuesday" |
  "wednesday" |
  "thursday" |
  "friday" |
  "saturday" |
  "sunday";

@doc("Single calendar owner profile. It is configured ahead of time and does not authenticate in v1.")
model OwnerProfile {
  id: string;
  displayName: string;
  timezone: string;
}

@doc("Owner-defined availability window for a specific day of week.")
model AvailabilityWindow {
  dayOfWeek: DayOfWeek;
  @doc("Local start time in HH:MM format.")
  startTime: string;
  @doc("Local end time in HH:MM format.")
  endTime: string;
}

@doc("Bookable event type configured by the calendar owner.")
model EventType {
  id: string;
  title: string;
  description: string;
  durationMinutes: int32;
}

@doc("Computed free slot for a specific event type.")
model Slot {
  eventTypeId: string;
  startsAt: utcDateTime;
  endsAt: utcDateTime;
  durationMinutes: int32;
}

@doc("Confirmed booking created by a guest without registration.")
model Booking {
  id: string;
  eventTypeId: string;
  guestName: string;
  guestEmail: string;
  startsAt: utcDateTime;
  endsAt: utcDateTime;
}

model UpcomingBookingItem {
  ...Booking;
  eventTypeTitle: string;
}

model ErrorBody {
  code: string;
  message: string;
}

model BadRequestError {
  @statusCode
  statusCode: 400;
  error: ErrorBody;
}

model NotFoundError {
  @statusCode
  statusCode: 404;
  error: ErrorBody;
}

@doc("Returned when the requested time interval overlaps with an existing booking, even if the event type is different.")
model BookingConflictError {
  @statusCode
  statusCode: 409;
  error: ErrorBody;
}

model EventTypeListResponse {
  items: EventType[];
}

model AvailabilityResponse {
  owner: OwnerProfile;
  windows: AvailabilityWindow[];
}

model UpcomingBookingsResponse {
  items: UpcomingBookingItem[];
}

model SlotListResponse {
  items: Slot[];
}

model BookingResponse {
  booking: Booking;
}

model CreateEventTypeRequest {
  title: string;
  description: string;
  durationMinutes: int32;
}

model CreatedEventTypeResponse {
  @statusCode
  statusCode: 201;
  eventType: EventType;
}

model UpdateAvailabilityRequest {
  windows: AvailabilityWindow[];
}

model UpdatedAvailabilityResponse {
  availability: AvailabilityResponse;
}

model CreateBookingRequest {
  eventTypeId: string;
  guestName: string;
  guestEmail: string;
  startsAt: utcDateTime;
}

model CreatedBookingResponse {
  @statusCode
  statusCode: 201;
  booking: Booking;
}

@tag("Owner")
@route("/owner/event-types")
interface OwnerEventTypesApi {
  @doc("List all event types configured by the owner.")
  @get
  list(): EventTypeListResponse;

  @doc("Create a new event type with a required duration in minutes.")
  @post
  create(@body body: CreateEventTypeRequest): CreatedEventTypeResponse | BadRequestError;
}

@tag("Owner")
@route("/owner/availability")
interface OwnerAvailabilityApi {
  @doc("Read owner availability windows that define when slots can be offered.")
  @get
  getAvailability(): AvailabilityResponse;

  @doc("Replace owner availability windows for the weekly schedule.")
  @put
  updateAvailability(@body body: UpdateAvailabilityRequest): UpdatedAvailabilityResponse | BadRequestError;
}

@tag("Owner")
@route("/owner/bookings")
interface OwnerBookingsApi {
  @doc("List upcoming bookings for all event types in a single owner-facing list.")
  @get
  @route("/upcoming")
  listUpcoming(): UpcomingBookingsResponse;
}

@tag("Public")
@route("/event-types")
interface PublicEventTypesApi {
  @doc("Public catalog of bookable event types visible to guests.")
  @get
  list(): EventTypeListResponse;

  @doc("List computed free slots for the selected event type and date range. Slots are derived from owner availability and exclude conflicting bookings.")
  @get
  @route("/{eventTypeId}/slots")
  listSlots(
    @path eventTypeId: string,
    @query from: utcDateTime,
    @query to: utcDateTime,
  ): SlotListResponse | BadRequestError | NotFoundError;
}

@tag("Public")
@route("/bookings")
interface PublicBookingsApi {
  @doc("Create a booking for the selected start time. The server must reject the request with 409 when the requested interval overlaps any existing booking, regardless of event type.")
  @post
  create(@body body: CreateBookingRequest): CreatedBookingResponse | BadRequestError | NotFoundError | BookingConflictError;
}
