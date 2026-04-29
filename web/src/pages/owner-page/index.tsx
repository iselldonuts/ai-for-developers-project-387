import {
  Alert,
  Badge,
  Button,
  Card,
  Divider,
  Grid,
  Group,
  Loader,
  Radio,
  SimpleGrid,
  Stack,
  Switch,
  Table,
  Text,
  TextInput,
  Textarea,
  ThemeIcon,
  Title,
} from "@mantine/core";
import { IconCalendarClock, IconClockHour3, IconSparkles } from "@tabler/icons-react";
import { useEffect, useMemo, useState } from "react";
import { apiClient } from "../../api/client";
import { PageShell } from "../../components/layout/page-shell";
import type {
  AvailabilityResponse,
  AvailabilityWindow,
  DayOfWeek,
  EventType,
} from "../../types";

const orderedDays: Array<{ key: DayOfWeek; shortLabel: string; fullLabel: string }> = [
  { key: "monday", shortLabel: "Пн", fullLabel: "Понедельник" },
  { key: "tuesday", shortLabel: "Вт", fullLabel: "Вторник" },
  { key: "wednesday", shortLabel: "Ср", fullLabel: "Среда" },
  { key: "thursday", shortLabel: "Чт", fullLabel: "Четверг" },
  { key: "friday", shortLabel: "Пт", fullLabel: "Пятница" },
  { key: "saturday", shortLabel: "Сб", fullLabel: "Суббота" },
  { key: "sunday", shortLabel: "Вс", fullLabel: "Воскресенье" },
];

const durationOptions = [
  { value: "15", label: "15 мин" },
  { value: "30", label: "30 мин" },
  { value: "45", label: "45 мин" },
  { value: "60", label: "60 мин" },
] as const;

interface DayAvailabilityFormValue {
  enabled: boolean;
  startTime: string;
  endTime: string;
}

type AvailabilityFormState = Record<DayOfWeek, DayAvailabilityFormValue>;

interface EventTypeFormState {
  title: string;
  description: string;
  durationMinutes: string;
}

const emptyEventTypeForm: EventTypeFormState = {
  title: "",
  description: "",
  durationMinutes: "30",
};

function buildEmptyEventTypeForm(): EventTypeFormState {
  return {
    ...emptyEventTypeForm,
  };
}

function buildInitialAvailabilityForm(): AvailabilityFormState {
  return orderedDays.reduce<AvailabilityFormState>((accumulator, day) => {
    accumulator[day.key] = {
      enabled: false,
      startTime: "09:00",
      endTime: "17:00",
    };
    return accumulator;
  }, {} as AvailabilityFormState);
}

function mapWindowsToForm(windows: AvailabilityWindow[]): AvailabilityFormState {
  const nextForm = buildInitialAvailabilityForm();

  for (const window of windows) {
    nextForm[window.dayOfWeek] = {
      enabled: true,
      startTime: window.startTime,
      endTime: window.endTime,
    };
  }

  return nextForm;
}

function validateAvailabilityForm(form: AvailabilityFormState): string | null {
  for (const day of orderedDays) {
    const value = form[day.key];

    if (!value.enabled) {
      continue;
    }

    if (!value.startTime || !value.endTime) {
      return `Укажите время "с" и "до" для дня ${day.fullLabel.toLowerCase()}.`;
    }

    if (value.startTime >= value.endTime) {
      return `Время начала должно быть раньше времени окончания для дня ${day.fullLabel.toLowerCase()}.`;
    }
  }

  return null;
}

function buildAvailabilityPayload(form: AvailabilityFormState): AvailabilityWindow[] {
  return orderedDays.flatMap((day) => {
    const value = form[day.key];

    if (!value.enabled) {
      return [];
    }

    return [
      {
        dayOfWeek: day.key,
        startTime: value.startTime,
        endTime: value.endTime,
      },
    ];
  });
}

function formatDurationLabel(durationMinutes: number): string {
  return `${durationMinutes} мин`;
}

function formatAvailabilitySummary(windows: AvailabilityWindow[]): string {
  if (windows.length === 0) {
    return "Расписание пока не задано";
  }

  return `${windows.length} ${windows.length === 1 ? "день" : windows.length < 5 ? "дня" : "дней"} в неделю`;
}

export function OwnerPage() {
  const [ownerProfile, setOwnerProfile] = useState<AvailabilityResponse["owner"] | null>(null);
  const [availabilityForm, setAvailabilityForm] = useState<AvailabilityFormState>(
    buildInitialAvailabilityForm,
  );
  const [availabilityError, setAvailabilityError] = useState<string | null>(null);
  const [availabilitySuccess, setAvailabilitySuccess] = useState<string | null>(null);
  const [isLoadingAvailability, setIsLoadingAvailability] = useState(true);
  const [isSavingAvailability, setIsSavingAvailability] = useState(false);

  const [eventTypes, setEventTypes] = useState<EventType[]>([]);
  const [eventTypeForm, setEventTypeForm] = useState<EventTypeFormState>(buildEmptyEventTypeForm);
  const [eventTypesError, setEventTypesError] = useState<string | null>(null);
  const [eventTypesSuccess, setEventTypesSuccess] = useState<string | null>(null);
  const [isLoadingEventTypes, setIsLoadingEventTypes] = useState(true);
  const [isCreatingEventType, setIsCreatingEventType] = useState(false);

  useEffect(() => {
    let isCancelled = false;

    async function loadOwnerData() {
      setIsLoadingAvailability(true);
      setIsLoadingEventTypes(true);
      setAvailabilityError(null);
      setEventTypesError(null);

      try {
        const [availabilityResponse, eventTypesResponse] = await Promise.all([
          apiClient.owner.getAvailability(),
          apiClient.owner.listEventTypes(),
        ]);

        if (isCancelled) {
          return;
        }

        setOwnerProfile(availabilityResponse.owner);
        setAvailabilityForm(mapWindowsToForm(availabilityResponse.windows));
        setEventTypes(eventTypesResponse.items);
      } catch (error) {
        const message =
          error instanceof Error ? error.message : "Не удалось загрузить данные владельца.";

        if (!isCancelled) {
          setAvailabilityError(message);
          setEventTypesError(message);
        }
      } finally {
        if (!isCancelled) {
          setIsLoadingAvailability(false);
          setIsLoadingEventTypes(false);
        }
      }
    }

    void loadOwnerData();

    return () => {
      isCancelled = true;
    };
  }, []);

  const enabledDaysCount = useMemo(
    () => orderedDays.filter((day) => availabilityForm[day.key].enabled).length,
    [availabilityForm],
  );

  async function handleAvailabilitySubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const validationError = validateAvailabilityForm(availabilityForm);
    if (validationError) {
      setAvailabilityError(validationError);
      setAvailabilitySuccess(null);
      return;
    }

    setIsSavingAvailability(true);
    setAvailabilityError(null);
    setAvailabilitySuccess(null);

    try {
      const windows = buildAvailabilityPayload(availabilityForm);
      const response = await apiClient.owner.updateAvailability({ windows });

      setOwnerProfile(response.availability.owner);
      setAvailabilityForm(mapWindowsToForm(response.availability.windows));
      setAvailabilitySuccess("Расписание владельца сохранено.");
    } catch (error) {
      setAvailabilityError(
        error instanceof Error ? error.message : "Не удалось сохранить расписание владельца.",
      );
    } finally {
      setIsSavingAvailability(false);
    }
  }

  async function handleEventTypeSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!eventTypeForm.title.trim()) {
      setEventTypesError("Укажите название события.");
      setEventTypesSuccess(null);
      return;
    }

    if (!eventTypeForm.durationMinutes) {
      setEventTypesError("Выберите длительность события.");
      setEventTypesSuccess(null);
      return;
    }

    setIsCreatingEventType(true);
    setEventTypesError(null);
    setEventTypesSuccess(null);

    try {
      const response = await apiClient.owner.createEventType({
        title: eventTypeForm.title.trim(),
        description: eventTypeForm.description.trim(),
        durationMinutes: Number(eventTypeForm.durationMinutes),
      });

      setEventTypes((current) => [...current, response.eventType]);
      setEventTypeForm(buildEmptyEventTypeForm());
      setEventTypesSuccess("Новый тип события создан.");
    } catch (error) {
      setEventTypesError(
        error instanceof Error ? error.message : "Не удалось создать тип события.",
      );
    } finally {
      setIsCreatingEventType(false);
    }
  }

  return (
    <PageShell size="lg">
      <Stack gap="xl">
        <Group justify="space-between" align="flex-start">
          <div>
            <Title order={1}>Owner</Title>
            <Text c="dimmed" maw={720} mt={6}>
              Настройте базовую недельную доступность владельца и создайте типы
              событий, которые будут использовать это расписание при построении
              слотов.
            </Text>
          </div>
          <Badge color="blue" variant="light" size="lg">
            {ownerProfile?.timezone ?? "Timezone loading"}
          </Badge>
        </Group>

        <SimpleGrid cols={{ base: 1, lg: 2 }} spacing="lg" verticalSpacing="lg">
          <Card radius="lg" shadow="sm" padding="lg" withBorder>
            <Stack gap="lg">
              <Group justify="space-between" align="center">
                <Group gap="sm">
                  <ThemeIcon size="lg" radius="md" variant="light" color="blue">
                    <IconCalendarClock size={20} />
                  </ThemeIcon>
                  <div>
                    <Title order={3}>Доступность</Title>
                    <Text size="sm" c="dimmed">
                      Общий недельный график владельца
                    </Text>
                  </div>
                </Group>
                <Badge variant="dot" color={enabledDaysCount > 0 ? "teal" : "gray"}>
                  {enabledDaysCount > 0 ? `${enabledDaysCount} активн.` : "Не настроено"}
                </Badge>
              </Group>

              <Text size="sm" c="dimmed">
                Сейчас все типы событий используют одно общее расписание. Позже
                сюда можно будет добавить пресеты доступности.
              </Text>

              {availabilityError ? <Alert color="red">{availabilityError}</Alert> : null}
              {availabilitySuccess ? <Alert color="teal">{availabilitySuccess}</Alert> : null}

              {isLoadingAvailability ? (
                <Group justify="center" py="xl">
                  <Loader size="sm" />
                </Group>
              ) : (
                <form onSubmit={handleAvailabilitySubmit}>
                  <Stack gap="sm">
                    {orderedDays.map((day, index) => {
                      const value = availabilityForm[day.key];

                      return (
                        <div key={day.key}>
                          <Grid align="center" gutter="sm">
                            <Grid.Col span={{ base: 12, sm: 4 }}>
                              <Group justify="space-between" wrap="nowrap">
                                <div>
                                  <Text fw={600}>{day.shortLabel}</Text>
                                  <Text size="xs" c="dimmed">
                                    {day.fullLabel}
                                  </Text>
                                </div>
                                <Switch
                                  checked={value.enabled}
                                  onChange={(event) => {
                                    const checked = event.currentTarget.checked;
                                    setAvailabilityForm((current) => ({
                                      ...current,
                                      [day.key]: {
                                        ...current[day.key],
                                        enabled: checked,
                                      },
                                    }));
                                  }}
                                  aria-label={`Доступность на ${day.fullLabel.toLowerCase()}`}
                                />
                              </Group>
                            </Grid.Col>
                            <Grid.Col span={{ base: 6, sm: 4 }}>
                              <TextInput
                                label="С"
                                type="time"
                                value={value.startTime}
                                disabled={!value.enabled}
                                onChange={(event) => {
                                  const nextValue = event.currentTarget.value;
                                  setAvailabilityForm((current) => ({
                                    ...current,
                                    [day.key]: {
                                      ...current[day.key],
                                      startTime: nextValue,
                                    },
                                  }));
                                }}
                              />
                            </Grid.Col>
                            <Grid.Col span={{ base: 6, sm: 4 }}>
                              <TextInput
                                label="До"
                                type="time"
                                value={value.endTime}
                                disabled={!value.enabled}
                                onChange={(event) => {
                                  const nextValue = event.currentTarget.value;
                                  setAvailabilityForm((current) => ({
                                    ...current,
                                    [day.key]: {
                                      ...current[day.key],
                                      endTime: nextValue,
                                    },
                                  }));
                                }}
                              />
                            </Grid.Col>
                          </Grid>
                          {index < orderedDays.length - 1 ? <Divider my="sm" /> : null}
                        </div>
                      );
                    })}

                    <Group justify="space-between" mt="md">
                      <Text size="sm" c="dimmed">
                        {formatAvailabilitySummary(buildAvailabilityPayload(availabilityForm))}
                      </Text>
                      <Button type="submit" loading={isSavingAvailability}>
                        Сохранить расписание
                      </Button>
                    </Group>
                  </Stack>
                </form>
              )}
            </Stack>
          </Card>

          <Card radius="lg" shadow="sm" padding="lg" withBorder>
            <Stack gap="lg">
              <Group justify="space-between" align="center">
                <Group gap="sm">
                  <ThemeIcon size="lg" radius="md" variant="light" color="orange">
                    <IconSparkles size={20} />
                  </ThemeIcon>
                  <div>
                    <Title order={3}>Типы событий</Title>
                    <Text size="sm" c="dimmed">
                      Создание публичных вариантов встречи
                    </Text>
                  </div>
                </Group>
                <Badge color="orange" variant="light">
                  {eventTypes.length} шт.
                </Badge>
              </Group>

              <Text size="sm" c="dimmed">
                Пользователь будет выбирать тип события, а слоттер построит слоты
                по общей доступности владельца и длительности события.
              </Text>

              {eventTypesError ? <Alert color="red">{eventTypesError}</Alert> : null}
              {eventTypesSuccess ? <Alert color="teal">{eventTypesSuccess}</Alert> : null}

              <form onSubmit={handleEventTypeSubmit}>
                <Stack gap="md">
                  <TextInput
                    label="Название"
                    placeholder="Например, Intro call"
                    value={eventTypeForm.title}
                    onChange={(event) => {
                      const nextValue = event.currentTarget.value;
                      setEventTypeForm((current) => ({
                        ...current,
                        title: nextValue,
                      }));
                    }}
                    required
                  />

                  <Textarea
                    label="Описание"
                    placeholder="Коротко объясните, для чего этот формат встречи"
                    minRows={3}
                    value={eventTypeForm.description}
                    onChange={(event) => {
                      const nextValue = event.currentTarget.value;
                      setEventTypeForm((current) => ({
                        ...current,
                        description: nextValue,
                      }));
                    }}
                  />

                  <Radio.Group
                    label="Длительность"
                    description="Выберите обязательное время для этого типа события"
                    value={eventTypeForm.durationMinutes}
                    onChange={(value) =>
                      setEventTypeForm((current) => ({
                        ...current,
                        durationMinutes: value,
                      }))
                    }
                  >
                    <Group mt="xs">
                      {durationOptions.map((option) => (
                        <Radio
                          key={option.value}
                          value={option.value}
                          label={option.label}
                        />
                      ))}
                    </Group>
                  </Radio.Group>

                  <Group justify="flex-end">
                    <Button type="submit" loading={isCreatingEventType}>
                      Создать событие
                    </Button>
                  </Group>
                </Stack>
              </form>

              <Divider />

              {isLoadingEventTypes ? (
                <Group justify="center" py="xl">
                  <Loader size="sm" />
                </Group>
              ) : eventTypes.length === 0 ? (
                <Alert color="blue" variant="light">
                  Типы событий еще не созданы.
                </Alert>
              ) : (
                <Table.ScrollContainer minWidth={520}>
                  <Table verticalSpacing="md" striped highlightOnHover>
                    <Table.Thead>
                      <Table.Tr>
                        <Table.Th>Событие</Table.Th>
                        <Table.Th>Описание</Table.Th>
                        <Table.Th>Длительность</Table.Th>
                      </Table.Tr>
                    </Table.Thead>
                    <Table.Tbody>
                      {eventTypes.map((eventType) => (
                        <Table.Tr key={eventType.id}>
                          <Table.Td>
                            <Group gap="sm" wrap="nowrap">
                              <ThemeIcon size="md" radius="xl" variant="light" color="orange">
                                <IconClockHour3 size={16} />
                              </ThemeIcon>
                              <Text fw={600}>{eventType.title}</Text>
                            </Group>
                          </Table.Td>
                          <Table.Td>
                            <Text size="sm" c={eventType.description ? "inherit" : "dimmed"}>
                              {eventType.description || "Без описания"}
                            </Text>
                          </Table.Td>
                          <Table.Td>
                            <Badge variant="light">{formatDurationLabel(eventType.durationMinutes)}</Badge>
                          </Table.Td>
                        </Table.Tr>
                      ))}
                    </Table.Tbody>
                  </Table>
                </Table.ScrollContainer>
              )}
            </Stack>
          </Card>
        </SimpleGrid>
      </Stack>
    </PageShell>
  );
}
