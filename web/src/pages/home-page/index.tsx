import {
  Alert,
  Badge,
  Button,
  Card,
  Grid,
  Group,
  Loader,
  Modal,
  SimpleGrid,
  Stack,
  Text,
  TextInput,
  ThemeIcon,
  Title,
} from "@mantine/core";
import {
  IconCalendarEvent,
  IconCheck,
  IconChevronLeft,
  IconChevronRight,
  IconClockHour3,
} from "@tabler/icons-react";
import { useEffect, useMemo, useState } from "react";
import { apiClient } from "../../api/client";
import { PageShell } from "../../components/layout/page-shell";
import type { EventType, Slot } from "../../types";

interface CalendarDay {
  date: Date;
  key: string;
  isCurrentMonth: boolean;
}

interface BookingFormState {
  guestName: string;
  guestEmail: string;
}

const weekDayLabels = ["Пн", "Вт", "Ср", "Чт", "Пт", "Сб", "Вс"];

function getMonthStart(value: Date): Date {
  return new Date(value.getFullYear(), value.getMonth(), 1);
}

function addMonths(value: Date, amount: number): Date {
  return new Date(value.getFullYear(), value.getMonth() + amount, 1);
}

function toDateKey(value: Date): string {
  const year = value.getFullYear();
  const month = `${value.getMonth() + 1}`.padStart(2, "0");
  const day = `${value.getDate()}`.padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function startOfCalendarGrid(value: Date): Date {
  const firstDay = getMonthStart(value);
  const offset = (firstDay.getDay() + 6) % 7;
  return new Date(firstDay.getFullYear(), firstDay.getMonth(), firstDay.getDate() - offset);
}

function buildCalendarDays(value: Date): CalendarDay[] {
  const month = value.getMonth();
  const start = startOfCalendarGrid(value);

  return Array.from({ length: 42 }, (_, index) => {
    const date = new Date(start.getFullYear(), start.getMonth(), start.getDate() + index);
    return {
      date,
      key: toDateKey(date),
      isCurrentMonth: date.getMonth() === month,
    };
  });
}

function formatMonthTitle(value: Date): string {
  return new Intl.DateTimeFormat("ru-RU", {
    month: "long",
    year: "numeric",
  }).format(value);
}

function formatSlotTime(value: string): string {
  return new Intl.DateTimeFormat("ru-RU", {
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}

function formatSlotDateTime(value: string): string {
  return new Intl.DateTimeFormat("ru-RU", {
    dateStyle: "full",
    timeStyle: "short",
  }).format(new Date(value));
}

function formatDuration(durationMinutes: number): string {
  return `${durationMinutes} мин`;
}

function formatSelectedDate(value: string): string {
  return new Intl.DateTimeFormat("ru-RU", {
    day: "numeric",
    month: "long",
    year: "numeric",
  }).format(new Date(value));
}

export function HomePage() {
  const [eventTypes, setEventTypes] = useState<EventType[]>([]);
  const [eventTypesError, setEventTypesError] = useState<string | null>(null);
  const [isLoadingEventTypes, setIsLoadingEventTypes] = useState(true);

  const [currentMonth, setCurrentMonth] = useState(() => getMonthStart(new Date()));
  const [selectedEventTypeId, setSelectedEventTypeId] = useState<string | null>(null);
  const [slots, setSlots] = useState<Slot[]>([]);
  const [slotsError, setSlotsError] = useState<string | null>(null);
  const [isLoadingSlots, setIsLoadingSlots] = useState(false);
  const [selectedDateKey, setSelectedDateKey] = useState<string | null>(null);
  const [selectedSlot, setSelectedSlot] = useState<Slot | null>(null);
  const [bookingForm, setBookingForm] = useState<BookingFormState>({
    guestName: "",
    guestEmail: "",
  });
  const [bookingError, setBookingError] = useState<string | null>(null);
  const [bookingSuccess, setBookingSuccess] = useState<string | null>(null);
  const [isBookingModalOpen, setIsBookingModalOpen] = useState(false);
  const [isCreatingBooking, setIsCreatingBooking] = useState(false);

  useEffect(() => {
    let cancelled = false;

    async function loadEventTypes() {
      setIsLoadingEventTypes(true);
      setEventTypesError(null);

      try {
        const response = await apiClient.public.listEventTypes();

        if (cancelled) {
          return;
        }

        setEventTypes(response.items);
        setSelectedEventTypeId((current) => current ?? response.items[0]?.id ?? null);
      } catch (error) {
        if (!cancelled) {
          setEventTypesError(
            error instanceof Error ? error.message : "Не удалось загрузить типы событий.",
          );
        }
      } finally {
        if (!cancelled) {
          setIsLoadingEventTypes(false);
        }
      }
    }

    void loadEventTypes();

    return () => {
      cancelled = true;
    };
  }, []);

  const monthRange = useMemo(() => {
    const from = getMonthStart(currentMonth);
    const to = new Date(currentMonth.getFullYear(), currentMonth.getMonth() + 1, 1);
    return { from, to };
  }, [currentMonth]);

  useEffect(() => {
    if (!selectedEventTypeId) {
      setSlots([]);
      setSelectedDateKey(null);
      return;
    }

    const eventTypeId = selectedEventTypeId;

    let cancelled = false;

    async function loadSlots() {
      setIsLoadingSlots(true);
      setSlotsError(null);
      setBookingSuccess(null);

      try {
        const response = await apiClient.public.listSlots(
          eventTypeId,
          monthRange.from.toISOString(),
          monthRange.to.toISOString(),
        );

        if (cancelled) {
          return;
        }

        setSlots(response.items);
        setSelectedDateKey((current) => {
          if (current && response.items.some((slot) => toDateKey(new Date(slot.startsAt)) === current)) {
            return current;
          }

          return null;
        });
      } catch (error) {
        if (!cancelled) {
          setSlotsError(error instanceof Error ? error.message : "Не удалось загрузить слоты.");
          setSlots([]);
          setSelectedDateKey(null);
        }
      } finally {
        if (!cancelled) {
          setIsLoadingSlots(false);
        }
      }
    }

    void loadSlots();

    return () => {
      cancelled = true;
    };
  }, [monthRange, selectedEventTypeId]);

  const selectedEventType = useMemo(
    () => eventTypes.find((item) => item.id === selectedEventTypeId) ?? null,
    [eventTypes, selectedEventTypeId],
  );

  const slotsByDate = useMemo(() => {
    const grouped = new Map<string, Slot[]>();

    for (const slot of slots) {
      const key = toDateKey(new Date(slot.startsAt));
      const items = grouped.get(key) ?? [];
      items.push(slot);
      grouped.set(key, items);
    }

    return grouped;
  }, [slots]);

  const calendarDays = useMemo(() => buildCalendarDays(currentMonth), [currentMonth]);
  const selectedDateSlots = selectedDateKey ? slotsByDate.get(selectedDateKey) ?? [] : [];

  async function refreshSlots() {
    if (!selectedEventTypeId) {
      return;
    }

    const eventTypeId = selectedEventTypeId;

    const response = await apiClient.public.listSlots(
      eventTypeId,
      monthRange.from.toISOString(),
      monthRange.to.toISOString(),
    );

    setSlots(response.items);
    setSelectedDateKey((current) => {
      if (current && response.items.some((slot) => toDateKey(new Date(slot.startsAt)) === current)) {
        return current;
      }

      return null;
    });
  }

  async function handleBookingSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!selectedSlot) {
      return;
    }

    setIsCreatingBooking(true);
    setBookingError(null);
    setBookingSuccess(null);

    try {
      await apiClient.public.createBooking({
        eventTypeId: selectedSlot.eventTypeId,
        guestName: bookingForm.guestName.trim(),
        guestEmail: bookingForm.guestEmail.trim(),
        startsAt: selectedSlot.startsAt,
      });

      await refreshSlots();
      setBookingSuccess("Бронирование подтверждено. Слот больше недоступен.");
      setIsBookingModalOpen(false);
      setSelectedSlot(null);
      setBookingForm({ guestName: "", guestEmail: "" });
    } catch (error) {
      const message = error instanceof Error ? error.message : "Не удалось создать бронирование.";
      setBookingError(message);

      if (error instanceof Error && "status" in error && error.status === 409) {
        await refreshSlots();
      }
    } finally {
      setIsCreatingBooking(false);
    }
  }

  return (
    <PageShell size="lg">
      <Stack gap="xl">
        <Stack gap={6}>
          <Title order={1}>Онлайн-запись</Title>
          <Text c="dimmed" maw={720}>
            Выберите формат встречи, затем дату и свободное время. После подтверждения
            выбранный слот сразу исчезает из доступных.
          </Text>
        </Stack>

        {bookingSuccess ? <Alert color="teal">{bookingSuccess}</Alert> : null}
        {eventTypesError ? <Alert color="red">{eventTypesError}</Alert> : null}

        {isLoadingEventTypes ? (
          <Group justify="center" py="xl">
            <Loader size="sm" />
          </Group>
        ) : eventTypes.length === 0 ? (
          <Alert color="blue" variant="light">
            Владелец пока не добавил типы событий.
          </Alert>
        ) : (
          <Stack gap="lg">
            <Card withBorder radius="lg" padding="lg">
              <Stack gap="md">
                <Group gap="sm">
                  <ThemeIcon size="lg" radius="md" variant="light" color="blue">
                    <IconCalendarEvent size={18} />
                  </ThemeIcon>
                  <div>
                    <Title order={3}>Типы событий</Title>
                    <Text size="sm" c="dimmed">
                      Сначала выберите формат записи
                    </Text>
                  </div>
                </Group>

                <SimpleGrid cols={{ base: 1, sm: 2, lg: 3 }}>
                  {eventTypes.map((eventType) => {
                    const isActive = eventType.id === selectedEventTypeId;

                    return (
                      <Card
                        key={eventType.id}
                        withBorder
                        radius="lg"
                        padding="lg"
                        style={{
                          cursor: "pointer",
                          borderColor: isActive ? "var(--mantine-color-blue-5)" : undefined,
                          background: isActive ? "rgba(59, 130, 246, 0.06)" : undefined,
                        }}
                        onClick={() => {
                          setSelectedEventTypeId(eventType.id);
                          setSelectedDateKey(null);
                          setSelectedSlot(null);
                          setBookingError(null);
                          setBookingSuccess(null);
                        }}
                      >
                        <Stack gap="md">
                          <Group justify="space-between" align="flex-start">
                            <Text fw={700}>{eventType.title}</Text>
                            <Badge variant={isActive ? "filled" : "light"}>
                              {formatDuration(eventType.durationMinutes)}
                            </Badge>
                          </Group>
                          <Text size="sm" c={eventType.description ? "inherit" : "dimmed"}>
                            {eventType.description || "Описание не добавлено"}
                          </Text>
                        </Stack>
                      </Card>
                    );
                  })}
                </SimpleGrid>
              </Stack>
            </Card>

            <Grid gutter="lg" align="stretch">
              <Grid.Col span={{ base: 12, lg: 7 }}>
                <Card withBorder radius="lg" padding="lg">
                  <Stack gap="md">
                    <Group justify="space-between">
                      <Group gap="sm">
                        <ThemeIcon size="lg" radius="md" variant="light" color="grape">
                          <IconCalendarEvent size={18} />
                        </ThemeIcon>
                        <div>
                          <Title order={3}>Календарь</Title>
                          <Text size="sm" c="dimmed">
                            {selectedEventType
                              ? `Свободные даты для "${selectedEventType.title}"`
                              : "Выберите тип события"}
                          </Text>
                        </div>
                      </Group>
                      <Group gap="xs">
                        <Button
                          variant="subtle"
                          size="compact-md"
                          onClick={() => setCurrentMonth((value) => addMonths(value, -1))}
                        >
                          <IconChevronLeft size={16} />
                        </Button>
                        <Text fw={600} tt="capitalize">
                          {formatMonthTitle(currentMonth)}
                        </Text>
                        <Button
                          variant="subtle"
                          size="compact-md"
                          onClick={() => setCurrentMonth((value) => addMonths(value, 1))}
                        >
                          <IconChevronRight size={16} />
                        </Button>
                      </Group>
                    </Group>

                    {slotsError ? <Alert color="red">{slotsError}</Alert> : null}

                    <SimpleGrid cols={7} spacing="xs" verticalSpacing="xs">
                      {weekDayLabels.map((label) => (
                        <Text key={label} ta="center" size="sm" fw={600} c="dimmed">
                          {label}
                        </Text>
                      ))}

                      {calendarDays.map((day) => {
                        const isSelected = day.key === selectedDateKey;
                        const slotsCount = slotsByDate.get(day.key)?.length ?? 0;
                        const isAvailable = slotsCount > 0;

                        return (
                          <Button
                            key={day.key}
                            variant={isSelected ? "filled" : "light"}
                            color={isAvailable ? "blue" : "gray"}
                            disabled={!isAvailable}
                            h={82}
                            styles={{
                              root: {
                                opacity: day.isCurrentMonth ? 1 : 0.4,
                              },
                            }}
                            onClick={() => setSelectedDateKey(day.key)}
                          >
                            <Stack gap={6} align="center">
                              <Text size="sm" fw={700} lh={1}>
                                {day.date.getDate()}
                              </Text>
                              <Text size="10px" lh={1.1} c={isSelected ? "white" : "dimmed"}>
                                {slotsCount > 0 ? `${slotsCount} слот.` : "Нет"}
                              </Text>
                            </Stack>
                          </Button>
                        );
                      })}
                    </SimpleGrid>

                    {isLoadingSlots ? (
                      <Group justify="center" py="xl">
                        <Loader size="sm" />
                      </Group>
                    ) : null}
                  </Stack>
                </Card>
              </Grid.Col>

              <Grid.Col span={{ base: 12, lg: 5 }}>
                <Card withBorder radius="lg" padding="lg">
                  <Stack gap="md">
                    <Group gap="sm">
                      <ThemeIcon size="lg" radius="md" variant="light" color="orange">
                        <IconClockHour3 size={18} />
                      </ThemeIcon>
                      <div>
                        <Title order={3}>Доступные слоты</Title>
                        <Text size="sm" c="dimmed">
                          {selectedDateKey ? `На ${formatSelectedDate(selectedDateKey)}` : "Выберите доступную дату"}
                        </Text>
                      </div>
                    </Group>

                    {selectedDateKey == null ? (
                      <Alert color="blue" variant="light">
                        Выберите доступную дату в календаре, чтобы увидеть свободные слоты.
                      </Alert>
                    ) : selectedDateSlots.length === 0 ? (
                      <Alert color="blue" variant="light">
                        На эту дату свободных слотов нет.
                      </Alert>
                    ) : (
                      <SimpleGrid cols={{ base: 2, sm: 3, lg: 2 }}>
                        {selectedDateSlots.map((slot) => (
                          <Button
                            key={slot.startsAt}
                            variant="light"
                            onClick={() => {
                              setSelectedSlot(slot);
                              setBookingError(null);
                              setIsBookingModalOpen(true);
                            }}
                          >
                            {formatSlotTime(slot.startsAt)}
                          </Button>
                        ))}
                      </SimpleGrid>
                    )}
                  </Stack>
                </Card>
              </Grid.Col>
            </Grid>
          </Stack>
        )}
      </Stack>

      <Modal
        opened={isBookingModalOpen}
        onClose={() => {
          setIsBookingModalOpen(false);
          setSelectedSlot(null);
          setBookingError(null);
        }}
        title="Подтверждение бронирования"
        centered
      >
        <Stack gap="md">
          {selectedSlot && selectedEventType ? (
            <Card radius="md" withBorder padding="md">
              <Stack gap={6}>
                <Group justify="space-between" align="flex-start">
                  <Text fw={700}>{selectedEventType.title}</Text>
                  <Badge variant="light">{formatDuration(selectedEventType.durationMinutes)}</Badge>
                </Group>
                <Text size="sm" c="dimmed">
                  {formatSlotDateTime(selectedSlot.startsAt)}
                </Text>
              </Stack>
            </Card>
          ) : null}

          {bookingError ? <Alert color="red">{bookingError}</Alert> : null}

          <form onSubmit={handleBookingSubmit}>
            <Stack gap="md">
              <TextInput
                label="Ваше имя"
                value={bookingForm.guestName}
                onChange={(event) => {
                  const value = event.currentTarget.value;

                  setBookingForm((current) => ({
                    ...current,
                    guestName: value,
                  }));
                }}
                required
              />
              <TextInput
                label="Email"
                type="email"
                value={bookingForm.guestEmail}
                onChange={(event) => {
                  const value = event.currentTarget.value;

                  setBookingForm((current) => ({
                    ...current,
                    guestEmail: value,
                  }));
                }}
                required
              />
              <Button type="submit" loading={isCreatingBooking} leftSection={<IconCheck size={16} />}>
                Забронировать слот
              </Button>
            </Stack>
          </form>
        </Stack>
      </Modal>
    </PageShell>
  );
}
