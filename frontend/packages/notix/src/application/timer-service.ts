interface TimerEntry {
	timerId: number;
	remaining: number;
	startedAt: number;
	callback: () => void;
}

export class TimerService {
	private timers = new Map<string, TimerEntry>();
	private paused = false;

	schedule(key: string, delay: number, callback: () => void): void {
		if (delay <= 0) return;
		this.cancel(key);

		const timerId = window.setTimeout(() => {
			this.timers.delete(key);
			callback();
		}, delay);

		this.timers.set(key, {
			timerId,
			remaining: delay,
			startedAt: Date.now(),
			callback,
		});
	}

	cancel(key: string): void {
		const entry = this.timers.get(key);
		if (entry) {
			clearTimeout(entry.timerId);
			this.timers.delete(key);
		}
	}

	cancelAll(): void {
		for (const entry of this.timers.values()) {
			clearTimeout(entry.timerId);
		}
		this.timers.clear();
	}

	pause(): void {
		if (this.paused) return;
		this.paused = true;

		for (const [key, entry] of this.timers) {
			clearTimeout(entry.timerId);
			const elapsed = Date.now() - entry.startedAt;
			this.timers.set(key, {
				...entry,
				remaining: Math.max(0, entry.remaining - elapsed),
			});
		}
	}

	resume(): void {
		if (!this.paused) return;
		this.paused = false;

		for (const [key, entry] of this.timers) {
			if (entry.remaining <= 0) {
				this.timers.delete(key);
				entry.callback();
				continue;
			}

			const timerId = window.setTimeout(() => {
				this.timers.delete(key);
				entry.callback();
			}, entry.remaining);

			this.timers.set(key, {
				...entry,
				timerId,
				startedAt: Date.now(),
			});
		}
	}

	has(key: string): boolean {
		return this.timers.has(key);
	}

	get isPaused(): boolean {
		return this.paused;
	}
}
