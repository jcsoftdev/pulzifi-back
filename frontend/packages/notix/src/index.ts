"use client";

// ─── Imperative API ──────────────────────────────────────────────────────────
export { notix } from "./presentation/api";

// ─── Components ──────────────────────────────────────────────────────────────
export { Toaster } from "./presentation/components/toaster";
export type { ToasterProps } from "./presentation/components/toaster";
export { NotixTrigger } from "./presentation/components/toast-trigger";
export type { NotixTriggerProps } from "./presentation/components/toast-trigger";
export { NotixAnchor } from "./presentation/components/anchor";
export type { NotixAnchorProps, NotixAnchorClassNames } from "./presentation/components/anchor";

// ─── Hooks ───────────────────────────────────────────────────────────────────
export { useToast, useToastStore } from "./presentation/hooks/use-toast";
export { useTriggerRect } from "./presentation/hooks/use-trigger-rect";

// ─── Types ───────────────────────────────────────────────────────────────────
export type {
	NotixOptions,
	NotixPosition,
	NotixStyles,
	NotixButton,
	NotixPromiseOptions,
	ToastState,
	ToastId,
	ToastData,
	ToastRenderProps,
	ToastLifecycle,
	AnimationMode,
	Duration,
	TriggerRect,
} from "./domain/entities/types";
