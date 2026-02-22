import { DEFAULT_TOAST_DURATION, EXIT_DURATION } from "../constants";
import { createToastData, updateToastData } from "../domain/entities/toast";
import type {
	NotixOptions,
	NotixPosition,
	NotixPromiseOptions,
	ToastData,
	ToastId,
} from "../domain/entities/types";
import type { IToastStore } from "../domain/ports/store";
import type { TimerService } from "./timer-service";

export class ToastManager {
	private defaultPosition: NotixPosition = "top-right";
	private globalOptions: Partial<NotixOptions> = {};

	constructor(
		private store: IToastStore,
		private timers: TimerService,
	) {}

	show(options: NotixOptions): ToastId {
		const toast = createToastData(options, {
			position: this.defaultPosition,
			globalOptions: this.globalOptions,
		});

		const existing = this.store.getById(toast.id);
		if (existing && !existing.exiting) {
			const updated = updateToastData(existing, options);
			this.store.update(toast.id, () => updated);
			this.scheduleDismiss(updated);
			return updated.id;
		}

		this.store.add(toast);
		this.scheduleDismiss(toast);
		return toast.id;
	}

	success(options: NotixOptions): ToastId {
		return this.show({ ...options, state: "success" });
	}

	error(options: NotixOptions): ToastId {
		return this.show({ ...options, state: "error" });
	}

	warning(options: NotixOptions): ToastId {
		return this.show({ ...options, state: "warning" });
	}

	info(options: NotixOptions): ToastId {
		return this.show({ ...options, state: "info" });
	}

	loading(options: NotixOptions): ToastId {
		return this.show({ ...options, state: "loading", duration: null });
	}

	action(options: NotixOptions): ToastId {
		return this.show({ ...options, state: "action" });
	}

	promise<T>(
		promise: Promise<T> | (() => Promise<T>),
		opts: NotixPromiseOptions<T>,
	): Promise<T> {
		const id = this.show({
			...opts.loading,
			state: "loading",
			duration: null,
			position: opts.position,
		});

		const p = typeof promise === "function" ? promise() : promise;

		p.then((data) => {
			if (opts.action) {
				const actionOpts =
					typeof opts.action === "function" ? opts.action(data) : opts.action;
				this.update(id, { ...actionOpts, state: "action" });
			} else {
				const successOpts =
					typeof opts.success === "function"
						? opts.success(data)
						: opts.success;
				this.update(id, { ...successOpts, state: "success" });
			}
		}).catch((err) => {
			const errorOpts =
				typeof opts.error === "function" ? opts.error(err) : opts.error;
			this.update(id, { ...errorOpts, state: "error" });
		});

		return p;
	}

	dismiss(id: ToastId): void {
		const toast = this.store.getById(id);
		if (!toast || toast.exiting) return;

		toast.onDismiss?.();
		this.timers.cancel(id);

		this.store.update(id, (t) => ({
			...t,
			exiting: true,
			lifecycle: "exiting" as const,
		}));

		setTimeout(() => {
			this.store.remove(id);
		}, EXIT_DURATION);
	}

	update(id: ToastId, options: Partial<NotixOptions>): void {
		const existing = this.store.getById(id);
		if (!existing) return;

		const updated = updateToastData(existing, options);
		this.store.update(id, () => updated);
		this.scheduleDismiss(updated);
	}

	clear(position?: NotixPosition): void {
		this.timers.cancelAll();
		this.store.clear(position);
	}

	setDefaultPosition(position: NotixPosition): void {
		this.defaultPosition = position;
	}

	setDefaultOptions(options: Partial<NotixOptions>): void {
		this.globalOptions = options;
	}

	pauseTimers(): void {
		this.timers.pause();
	}

	resumeTimers(): void {
		this.timers.resume();
	}

	private scheduleDismiss(toast: ToastData): void {
		const duration = toast.duration ?? DEFAULT_TOAST_DURATION;
		if (duration === null || duration <= 0) return;

		this.timers.schedule(toast.id, duration, () => {
			toast.onAutoClose?.();
			this.dismiss(toast.id);
		});
	}
}
