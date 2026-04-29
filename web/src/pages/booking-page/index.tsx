import { Stack, Text, Title } from "@mantine/core";
import { PageShell } from "../../components/layout/page-shell";

export function BookingPage() {
  return (
    <PageShell>
      <Stack gap="md">
        <Title order={2}>Booking</Title>
        <Text c="dimmed">
          Placeholder page for slot selection and booking creation flow.
        </Text>
      </Stack>
    </PageShell>
  );
}
