export type {
	INotificationPort,
	NotificationId,
	NotificationLevel,
	NotificationOptions,
	NotificationPromiseOptions,
} from "./notification-port";

export { NotixNotificationAdapter } from "./notix-adapter";
export { NotificationProvider } from "./notification-provider";

// ─── Default singleton ───────────────────────────────────────────────────────
// Swap the adapter here to change the underlying toast library.

import { NotixNotificationAdapter } from "./notix-adapter";
import type { INotificationPort } from "./notification-port";

export const notification: INotificationPort = new NotixNotificationAdapter();
