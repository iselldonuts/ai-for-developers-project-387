export type ID = string;

export type DayOfWeek =
  | "monday"
  | "tuesday"
  | "wednesday"
  | "thursday"
  | "friday"
  | "saturday"
  | "sunday";

export interface OwnerProfile {
  id: ID;
  displayName: string;
  timezone: string;
}

export interface AvailabilityWindow {
  dayOfWeek: DayOfWeek;
  startTime: string;
  endTime: string;
}

export interface AvailabilityResponse {
  owner: OwnerProfile;
  windows: AvailabilityWindow[];
}

export interface UpdateAvailabilityRequest {
  windows: AvailabilityWindow[];
}

export interface UpdatedAvailabilityResponse {
  availability: AvailabilityResponse;
}

export interface EventType {
  id: ID;
  title: string;
  description: string;
  durationMinutes: number;
}

export interface EventTypeListResponse {
  items: EventType[];
}

export interface CreateEventTypeRequest {
  title: string;
  description: string;
  durationMinutes: number;
}

export interface CreatedEventTypeResponse {
  eventType: EventType;
}

export interface Slot {
  eventTypeId: ID;
  startsAt: string;
  endsAt: string;
  durationMinutes: number;
}

export interface SlotListResponse {
  items: Slot[];
}

export interface Booking {
  id: ID;
  eventTypeId: ID;
  guestName: string;
  guestEmail: string;
  startsAt: string;
  endsAt: string;
}

export interface UpcomingBookingItem extends Booking {
  eventTypeTitle: string;
}

export interface UpcomingBookingsResponse {
  items: UpcomingBookingItem[];
}

export interface CreateBookingRequest {
  eventTypeId: ID;
  guestName: string;
  guestEmail: string;
  startsAt: string;
}

export interface CreatedBookingResponse {
  booking: Booking;
}

export interface ErrorBody {
  code: string;
  message: string;
}
