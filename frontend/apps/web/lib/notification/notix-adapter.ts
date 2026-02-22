import { notix } from "@workspace/notix";
import type {
	INotificationPort,
	NotificationId,
	NotificationOptions,
	NotificationPromiseOptions,
} from "./notification-port";

export class NotixNotificationAdapter implements INotificationPort {
	show(options: NotificationOptions): NotificationId {
		return notix.show({
			title: options.title,
			description: options.description,
			state: options.level,
			duration: options.duration,
		});
	}

	success(options: NotificationOptions): NotificationId {
		return notix.success({
			title: options.title,
			description: options.description,
			duration: options.duration,
		});
	}

	error(options: NotificationOptions): NotificationId {
		return notix.error({
			title: options.title,
			description: options.description,
			duration: options.duration,
		});
	}

	warning(options: NotificationOptions): NotificationId {
		return notix.warning({
			title: options.title,
			description: options.description,
			duration: options.duration,
		});
	}

	info(options: NotificationOptions): NotificationId {
		return notix.info({
			title: options.title,
			description: options.description,
			duration: options.duration,
		});
	}

	loading(options: NotificationOptions): NotificationId {
		return notix.loading({
			title: options.title,
			description: options.description,
		});
	}

	action(options: NotificationOptions): NotificationId {
		return notix.action({
			title: options.title,
			description: options.description,
			duration: options.duration,
		});
	}

	promise<T>(
		promise: Promise<T> | (() => Promise<T>),
		options: NotificationPromiseOptions<T>,
	): Promise<T> {
		return notix.promise(promise, {
			loading: { title: options.loading.title },
			success:
				typeof options.success === "function"
					? (data: T) => {
							const opts = options.success as (data: T) => NotificationOptions;
							const result = opts(data);
							return { title: result.title, description: result.description };
						}
					: {
							title: options.success.title,
							description: options.success.description,
						},
			error:
				typeof options.error === "function"
					? (err: unknown) => {
							const opts = options.error as (err: unknown) => NotificationOptions;
							const result = opts(err);
							return { title: result.title, description: result.description };
						}
					: {
							title: options.error.title,
							description: options.error.description,
						},
		});
	}

	dismiss(id: NotificationId): void {
		notix.dismiss(id);
	}

	clear(): void {
		notix.clear();
	}
}
