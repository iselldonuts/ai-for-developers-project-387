import { appEnv } from "../lib/env";
import type {
  AvailabilityResponse,
  CreateBookingRequest,
  CreateEventTypeRequest,
  CreatedBookingResponse,
  CreatedEventTypeResponse,
  EventTypeListResponse,
  SlotListResponse,
  UpdateAvailabilityRequest,
  UpdatedAvailabilityResponse,
} from "../types";

class ApiError extends Error {
  constructor(
    message: string,
    public readonly status: number,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers);

  if (init?.body != null && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const response = await fetch(`${appEnv.apiBaseUrl}${path}`, {
    ...init,
    headers,
  });

  if (!response.ok) {
    let message = `Request failed with status ${response.status}`;

    try {
      const body = (await response.json()) as { error?: { message?: string } };
      message = body.error?.message ?? message;
    } catch {
      // Ignore JSON parsing errors for non-JSON responses.
    }

    throw new ApiError(message, response.status);
  }

  return (await response.json()) as T;
}

export const apiClient = {
  public: {
    listEventTypes() {
      return request<EventTypeListResponse>("/event-types");
    },
    listSlots(eventTypeId: string, from: string, to: string) {
      const searchParams = new URLSearchParams({
        from,
        to,
      });

      return request<SlotListResponse>(`/event-types/${eventTypeId}/slots?${searchParams.toString()}`);
    },
    createBooking(body: CreateBookingRequest) {
      return request<CreatedBookingResponse>("/bookings", {
        method: "POST",
        body: JSON.stringify(body),
      });
    },
  },
  owner: {
    getAvailability() {
      return request<AvailabilityResponse>("/owner/availability");
    },
    updateAvailability(body: UpdateAvailabilityRequest) {
      return request<UpdatedAvailabilityResponse>("/owner/availability", {
        method: "PUT",
        body: JSON.stringify(body),
      });
    },
    listEventTypes() {
      return request<EventTypeListResponse>("/owner/event-types");
    },
    createEventType(body: CreateEventTypeRequest) {
      return request<CreatedEventTypeResponse>("/owner/event-types", {
        method: "POST",
        body: JSON.stringify(body),
      });
    },
  },
};
