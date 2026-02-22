import {
	AUTO_COLLAPSE_DELAY,
	AUTO_EXPAND_DELAY,
	DEFAULT_TOAST_DURATION,
} from "../../constants";
import type {
	NotixOptions,
	NotixPosition,
	ToastData,
	ToastState,
} from "./types";

let idCounter = 0;
const generateId = () =>
	`${++idCounter}-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`;

function resolveAutopilot(
	opts: NotixOptions,
	duration: number | null,
): { expandDelayMs?: number; collapseDelayMs?: number } {
	if (opts.autopilot === false || !duration || duration <= 0) return {};
	const cfg = typeof opts.autopilot === "object" ? opts.autopilot : undefined;
	const clamp = (v: number) => Math.min(duration, Math.max(0, v));
	return {
		expandDelayMs: clamp(cfg?.expand ?? AUTO_EXPAND_DELAY),
		collapseDelayMs: clamp(cfg?.collapse ?? AUTO_COLLAPSE_DELAY),
	};
}

export function createToastData(
	options: NotixOptions,
	defaults: {
		position: NotixPosition;
		globalOptions?: Partial<NotixOptions>;
	},
): ToastData {
	const merged = { ...defaults.globalOptions, ...options };
	const duration = merged.duration ?? DEFAULT_TOAST_DURATION;
	const auto = resolveAutopilot(merged, duration);
	const triggerRect =
		options.triggerRect ??
		(options.triggerRef?.current
			? toBoundingRect(options.triggerRef.current.getBoundingClientRect())
			: undefined);

	return {
		id: merged.id ?? generateId(),
		instanceId: generateId(),
		state: (merged.state ?? "info"),
		title: merged.title ?? "",
		description: merged.description,
		position: merged.position ?? defaults.position,
		duration,
		icon: merged.icon,
		styles: { ...defaults.globalOptions?.styles, ...options.styles },
		className: merged.className,
		button: merged.button,
		animation: merged.animation ?? "slide",
		triggerRect,
		render: merged.render,
		autopilot: merged.autopilot,
		lifecycle: "entering",
		createdAt: Date.now(),
		exiting: false,
		autoExpandDelayMs: auto.expandDelayMs,
		autoCollapseDelayMs: auto.collapseDelayMs,
		onDismiss: merged.onDismiss,
		onAutoClose: merged.onAutoClose,
	};
}

export function updateToastData(
	existing: ToastData,
	updates: Partial<NotixOptions>,
): ToastData {
	const duration = updates.duration ?? existing.duration;
	const auto = resolveAutopilot(
		{ ...existing, ...updates },
		duration,
	);

	return {
		...existing,
		...updates,
		id: existing.id,
		instanceId: generateId(),
		state: (updates.state ?? existing.state),
		title: updates.title ?? existing.title,
		position: updates.position ?? existing.position,
		duration,
		styles: { ...existing.styles, ...updates.styles },
		animation: updates.animation ?? existing.animation,
		triggerRect: updates.triggerRect ?? existing.triggerRect,
		lifecycle: existing.lifecycle,
		createdAt: existing.createdAt,
		exiting: false,
		autoExpandDelayMs: auto.expandDelayMs,
		autoCollapseDelayMs: auto.collapseDelayMs,
	};
}

function toBoundingRect(rect: DOMRect) {
	return {
		top: rect.top,
		left: rect.left,
		width: rect.width,
		height: rect.height,
		bottom: rect.bottom,
		right: rect.right,
	};
}
