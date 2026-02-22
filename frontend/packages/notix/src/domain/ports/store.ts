import type { NotixPosition, ToastData, ToastId } from "../entities/types";

export interface IToastStore {
	getSnapshot(): ReadonlyArray<ToastData>;
	subscribe(listener: () => void): () => void;
	add(toast: ToastData): void;
	update(id: ToastId, updater: (toast: ToastData) => ToastData): void;
	remove(id: ToastId): void;
	clear(position?: NotixPosition): void;
	getById(id: ToastId): ToastData | undefined;
}
