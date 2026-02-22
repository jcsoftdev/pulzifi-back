import { TimerService } from "../application/timer-service";
import { ToastManager } from "../application/toast-manager";
import type {
	NotixOptions,
	NotixPosition,
	NotixPromiseOptions,
	ToastId,
} from "../domain/entities/types";
import { ReactiveToastStore } from "../infrastructure/store/reactive-store";

const store = new ReactiveToastStore();
const timers = new TimerService();
const manager = new ToastManager(store, timers);

export const notix = {
	show: (opts: NotixOptions): ToastId => manager.show(opts),
	success: (opts: NotixOptions): ToastId => manager.success(opts),
	error: (opts: NotixOptions): ToastId => manager.error(opts),
	warning: (opts: NotixOptions): ToastId => manager.warning(opts),
	info: (opts: NotixOptions): ToastId => manager.info(opts),
	loading: (opts: NotixOptions): ToastId => manager.loading(opts),
	action: (opts: NotixOptions): ToastId => manager.action(opts),
	promise: <T,>(
		p: Promise<T> | (() => Promise<T>),
		opts: NotixPromiseOptions<T>,
	): Promise<T> => manager.promise(p, opts),
	dismiss: (id: ToastId): void => manager.dismiss(id),
	update: (id: ToastId, opts: Partial<NotixOptions>): void =>
		manager.update(id, opts),
	clear: (position?: NotixPosition): void => manager.clear(position),
	setDefaultPosition: (position: NotixPosition): void =>
		manager.setDefaultPosition(position),
	setDefaultOptions: (options: Partial<NotixOptions>): void =>
		manager.setDefaultOptions(options),
};

export function getGlobalStore(): ReactiveToastStore {
	return store;
}

export function getGlobalManager(): ToastManager {
	return manager;
}
