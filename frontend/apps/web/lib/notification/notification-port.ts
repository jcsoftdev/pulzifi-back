import type { ReactNode } from "react";

export type NotificationLevel =
	| "success"
	| "error"
	| "warning"
	| "info"
	| "loading"
	| "action";

export interface NotificationOptions {
	title: string;
	description?: ReactNode | string;
	level?: NotificationLevel;
	duration?: number | null;
}

export interface NotificationPromiseOptions<T = unknown> {
	loading: Pick<NotificationOptions, "title">;
	success: NotificationOptions | ((data: T) => NotificationOptions);
	error: NotificationOptions | ((err: unknown) => NotificationOptions);
}

export type NotificationId = string;

export interface INotificationPort {
	show(options: NotificationOptions): NotificationId;
	success(options: NotificationOptions): NotificationId;
	error(options: NotificationOptions): NotificationId;
	warning(options: NotificationOptions): NotificationId;
	info(options: NotificationOptions): NotificationId;
	loading(options: NotificationOptions): NotificationId;
	action(options: NotificationOptions): NotificationId;
	promise<T>(
		promise: Promise<T> | (() => Promise<T>),
		options: NotificationPromiseOptions<T>,
	): Promise<T>;
	dismiss(id: NotificationId): void;
	clear(): void;
}
