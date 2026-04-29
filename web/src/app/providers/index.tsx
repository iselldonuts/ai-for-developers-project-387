import { MantineProvider } from "@mantine/core";
import type { PropsWithChildren } from "react";
import "@mantine/core/styles.css";

export function AppProviders({ children }: PropsWithChildren) {
  return (
    <MantineProvider
      theme={{
        primaryColor: "blue",
        fontFamily: "ui-sans-serif, system-ui, sans-serif",
      }}
    >
      {children}
    </MantineProvider>
  );
}
