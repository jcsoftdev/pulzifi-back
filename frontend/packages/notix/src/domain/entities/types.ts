import type { ReactNode, RefObject } from "react";

// ─── Value Objects ───────────────────────────────────────────────────────────

export type ToastId = string;
export type Duration = number | null;

export interface TriggerRect {
	top: number;
	left: number;
	width: number;
	height: number;
	bottom: number;
	right: number;
}

// ─── Enums ───────────────────────────────────────────────────────────────────

export type ToastState =
	| "success"
	| "error"
	| "warning"
	| "info"
	| "loading"
	| "action";

export const NOTIX_POSITIONS = [
	"top-left",
	"top-center",
	"top-right",
	"bottom-left",
	"bottom-center",
	"bottom-right",
] as const;

export type NotixPosition = (typeof NOTIX_POSITIONS)[number];

export type AnimationMode = "slide" | "morph" | "fly";

export type ToastLifecycle = "entering" | "visible" | "exiting" | "removed";

// ─── Styles & Buttons ────────────────────────────────────────────────────────

export interface NotixStyles {
	toast?: string;
	title?: string;
	description?: string;
	badge?: string;
	button?: string;
}

export interface NotixButton {
	title: string;
	onClick: () => void;
}

// ─── Render Props (headless) ─────────────────────────────────────────────────

export interface ToastRenderProps {
	toast: ToastData;
	dismiss: () => void;
	isExpanded: boolean;
	toggle: () => void;
	lifecycle: ToastLifecycle;
}

// ─── Options ─────────────────────────────────────────────────────────────────

export interface NotixOptions {
	id?: ToastId;
	title?: string;
	description?: ReactNode | string;
	state?: ToastState;
	position?: NotixPosition;
	duration?: Duration;
	icon?: ReactNode | null;
	styles?: NotixStyles;
	className?: string;
	button?: NotixButton;
	animation?: AnimationMode;
	triggerRef?: RefObject<HTMLElement | null>;
	triggerRect?: TriggerRect;
	render?: (props: ToastRenderProps) => ReactNode;
	autopilot?: boolean | { expand?: number; collapse?: number };
	onDismiss?: () => void;
	onAutoClose?: () => void;
}

// ─── Internal Toast Data ─────────────────────────────────────────────────────

export interface ToastData {
	readonly id: ToastId;
	readonly instanceId: string;
	readonly state: ToastState;
	readonly title: string;
	readonly description?: ReactNode | string;
	readonly position: NotixPosition;
	readonly duration: Duration;
	readonly icon?: ReactNode | null;
	readonly styles?: NotixStyles;
	readonly className?: string;
	readonly button?: NotixButton;
	readonly animation: AnimationMode;
	readonly triggerRect?: TriggerRect;
	readonly render?: (props: ToastRenderProps) => ReactNode;
	readonly autopilot?: boolean | { expand?: number; collapse?: number };
	readonly lifecycle: ToastLifecycle;
	readonly createdAt: number;
	readonly exiting: boolean;
	readonly autoExpandDelayMs?: number;
	readonly autoCollapseDelayMs?: number;
	readonly onDismiss?: () => void;
	readonly onAutoClose?: () => void;
}

// ─── Promise Options ─────────────────────────────────────────────────────────

export interface NotixPromiseOptions<T = unknown> {
	loading: Pick<NotixOptions, "title" | "icon">;
	success: NotixOptions | ((data: T) => NotixOptions);
	error: NotixOptions | ((err: unknown) => NotixOptions);
	action?: NotixOptions | ((data: T) => NotixOptions);
	position?: NotixPosition;
}
