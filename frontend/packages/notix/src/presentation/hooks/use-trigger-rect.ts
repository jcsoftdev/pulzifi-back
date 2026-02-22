"use client";

import { useCallback, useRef } from "react";
import type { TriggerRect } from "../../domain/entities/types";

export function useTriggerRect() {
	const ref = useRef<HTMLElement>(null);

	const getRect = useCallback((): TriggerRect | undefined => {
		const el = ref.current;
		if (!el) return undefined;
		const rect = el.getBoundingClientRect();
		return {
			top: rect.top,
			left: rect.left,
			width: rect.width,
			height: rect.height,
			bottom: rect.bottom,
			right: rect.right,
		};
	}, []);

	return { ref, getRect };
}
