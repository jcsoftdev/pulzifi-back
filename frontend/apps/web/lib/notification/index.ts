export type {
  INotificationPort,
  NotificationId,
  NotificationLevel,
  NotificationOptions,
  NotificationPromiseOptions,
} from './notification-port'
export { NotificationProvider } from './notification-provider'
export { NotixNotificationAdapter } from './notix-adapter'

// ─── Default singleton ───────────────────────────────────────────────────────
// Swap the adapter here to change the underlying toast library.

import type { INotificationPort } from './notification-port'
import { NotixNotificationAdapter } from './notix-adapter'

export const notification: INotificationPort = new NotixNotificationAdapter()
