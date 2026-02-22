"use client";

import { useSyncExternalStore } from "react";
import type { ToastData } from "../../domain/entities/types";
import { getGlobalStore, notix } from "../api";

const EMPTY_TOASTS: ReadonlyArray<ToastData> = [];

export function useToastStore(): ReadonlyArray<ToastData> {
	const store = getGlobalStore();
	return useSyncExternalStore(
		store.subscribe.bind(store),
		store.getSnapshot.bind(store),
		() => EMPTY_TOASTS,
	);
}

export function useToast() {
	const toasts = useToastStore();
	return {
		toasts,
		show: notix.show,
		success: notix.success,
		error: notix.error,
		warning: notix.warning,
		info: notix.info,
		loading: notix.loading,
		action: notix.action,
		promise: notix.promise,
		dismiss: notix.dismiss,
		update: notix.update,
		clear: notix.clear,
	};
}
