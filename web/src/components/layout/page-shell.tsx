import { AppShell, Box, Container, Group, Tabs, Text } from "@mantine/core";
import type { PropsWithChildren } from "react";
import { useLocation, useNavigate } from "react-router-dom";

interface PageShellProps extends PropsWithChildren {
  size?: string;
}

export function PageShell({ children, size = "md" }: PageShellProps) {
  const location = useLocation();
  const navigate = useNavigate();
  const currentSection = location.pathname.startsWith("/owner") ? "/owner" : "/";

  return (
    <AppShell padding="md" header={{ height: 64 }}>
      <AppShell.Header>
        <Container size="lg" h="100%">
          <Group justify="space-between" h="100%" py="md">
            <Text fw={700}>calcal</Text>
            <Tabs
              variant="pills"
              radius="xl"
              value={currentSection}
              onChange={(value) => {
                if (value) {
                  navigate(value);
                }
              }}
            >
              <Tabs.List>
                <Tabs.Tab value="/">Пользователь</Tabs.Tab>
                <Tabs.Tab value="/owner">Владелец</Tabs.Tab>
              </Tabs.List>
            </Tabs>
          </Group>
        </Container>
      </AppShell.Header>
      <AppShell.Main>
        <Container size={size}>
          <Box py="xl">{children}</Box>
        </Container>
      </AppShell.Main>
    </AppShell>
  );
}
