import type { NotixPosition, ToastData, ToastId } from "../../domain/entities/types";
import type { IToastStore } from "../../domain/ports/store";

export class ReactiveToastStore implements IToastStore {
	private toasts: ToastData[] = [];
	private snapshot: ReadonlyArray<ToastData> = [];
	private readonly listeners = new Set<() => void>();

	getSnapshot(): ReadonlyArray<ToastData> {
		return this.snapshot;
	}

	subscribe(listener: () => void): () => void {
		this.listeners.add(listener);
		return () => this.listeners.delete(listener);
	}

	add(toast: ToastData): void {
		this.toasts = [...this.toasts.filter((t) => t.id !== toast.id), toast];
		this.emit();
	}

	update(id: ToastId, updater: (toast: ToastData) => ToastData): void {
		let changed = false;
		this.toasts = this.toasts.map((t) => {
			if (t.id === id) {
				changed = true;
				return updater(t);
			}
			return t;
		});
		if (changed) this.emit();
	}

	remove(id: ToastId): void {
		const len = this.toasts.length;
		this.toasts = this.toasts.filter((t) => t.id !== id);
		if (this.toasts.length !== len) this.emit();
	}

	clear(position?: NotixPosition): void {
		if (position) {
			this.toasts = this.toasts.filter((t) => t.position !== position);
		} else {
			this.toasts = [];
		}
		this.emit();
	}

	getById(id: ToastId): ToastData | undefined {
		return this.toasts.find((t) => t.id === id);
	}

	private emit(): void {
		this.snapshot = Object.freeze([...this.toasts]);
		for (const fn of this.listeners) fn();
	}
}
