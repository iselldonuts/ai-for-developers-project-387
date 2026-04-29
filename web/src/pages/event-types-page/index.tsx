import { List, Stack, Text, Title } from "@mantine/core";
import { PageShell } from "../../components/layout/page-shell";

export function EventTypesPage() {
  return (
    <PageShell>
      <Stack gap="md">
        <Title order={2}>Event types</Title>
        <Text c="dimmed">
          Placeholder page for future event type listing and management.
        </Text>
        <List spacing="xs">
          <List.Item>Generated API client will be connected from `src/api/generated`.</List.Item>
          <List.Item>Page-level composition stays in `pages/`.</List.Item>
          <List.Item>Reusable UI belongs in `components/`.</List.Item>
        </List>
      </Stack>
    </PageShell>
  );
}
