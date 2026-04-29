# API Contract Notes

## Domain model

- `OwnerProfile`: заранее заданный владелец календаря. В v1 не регистрируется и не аутентифицируется.
- `AvailabilityWindow`: доступные часы владельца по дням недели, из которых вычисляются возможные слоты.
- `EventType`: тип события с `id`, `title`, `description` и обязательным `durationMinutes`.
- `Slot`: вычисляемый свободный интервал для конкретного типа события. Формируется из weekly availability и длительности выбранного типа события.
- `Booking`: подтвержденная запись гостя на конкретный интервал времени.

## Guest scenario

Публичный сценарий гостя в спецификации выражен последовательностью HTTP-операций:

1. `GET /event-types` — посмотреть доступные виды бронирования.
2. `GET /event-types/{eventTypeId}/slots?from=...&to=...` — получить свободные слоты для выбранного типа события.
3. `POST /bookings` — создать бронирование без аккаунта и без логина.

## Coverage checklist

- Owner может создавать типы событий: `POST /owner/event-types`
- Owner может просматривать все будущие встречи одним списком: `GET /owner/bookings/upcoming`
- Owner может управлять доступными часами: `GET /owner/availability`, `PUT /owner/availability`
- Guest может просматривать публичный каталог типов событий: `GET /event-types`
- Guest может выбирать свободный слот в рамках доступности и duration выбранного типа: `GET /event-types/{eventTypeId}/slots`
- Guest может создавать бронирование без регистрации: `POST /bookings`
- Нельзя создать две записи на пересекающееся время, даже для разных типов событий: `POST /bookings` возвращает `409 Conflict`
