import { Geist, Geist_Mono } from "next/font/google"

import "@workspace/ui/globals.css"
import { Providers } from "@/components/providers"
import { AppShell } from "@/components/app-shell"
import { UsageService } from "@/features/usage/domain/services/usage-service"
import { NotificationService } from "@/features/notifications/domain/services/notification-service"

const fontSans = Geist({
  subsets: ["latin"],
  variable: "--font-sans",
})

const fontMono = Geist_Mono({
  subsets: ["latin"],
  variable: "--font-mono",
})

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  // Fetch de datos en el servidor usando domain services
  // Cada feature tiene su propio service en su domain layer
  const checksData = await UsageService.getChecksData()
  const notificationsData = await NotificationService.getNotificationsData()

  return (
    <html lang="en" suppressHydrationWarning>
      <body
        className={`${fontSans.variable} ${fontMono.variable} font-sans antialiased`}
      >
        <Providers>
          <AppShell 
            checksData={checksData}
            hasNotifications={notificationsData.hasNotifications}
            notificationCount={notificationsData.notificationCount}
          >
            {children}
          </AppShell>
        </Providers>
      </body>
    </html>
  )
}
