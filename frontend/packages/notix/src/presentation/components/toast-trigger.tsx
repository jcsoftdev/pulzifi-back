"use client";

import {
	type ReactElement,
	cloneElement,
	useCallback,
	useRef,
} from "react";
import type { AnimationMode, NotixOptions } from "../../domain/entities/types";
import { notix } from "../api";

export interface NotixTriggerProps {
	children: ReactElement<{ ref?: React.Ref<HTMLElement>; onClick?: () => void }>;
	animation?: AnimationMode;
	toastOptions: NotixOptions;
}

export function NotixTrigger({
	children,
	animation = "morph",
	toastOptions,
}: NotixTriggerProps) {
	const triggerRef = useRef<HTMLElement>(null);

	const handleTrigger = useCallback(() => {
		const el = triggerRef.current;
		const rect = el?.getBoundingClientRect();

		notix.show({
			...toastOptions,
			animation,
			triggerRect: rect
				? {
						top: rect.top,
						left: rect.left,
						width: rect.width,
						height: rect.height,
						bottom: rect.bottom,
						right: rect.right,
					}
				: undefined,
		});
	}, [toastOptions, animation]);

	return cloneElement(children, {
		ref: triggerRef,
		onClick: handleTrigger,
	});
}
