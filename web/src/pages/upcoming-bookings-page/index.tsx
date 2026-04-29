import { Stack, Text, Title } from "@mantine/core";
import { PageShell } from "../../components/layout/page-shell";

export function UpcomingBookingsPage() {
  return (
    <PageShell>
      <Stack gap="md">
        <Title order={2}>Upcoming bookings</Title>
        <Text c="dimmed">
          Placeholder page for owner-facing upcoming bookings list.
        </Text>
      </Stack>
    </PageShell>
  );
}
