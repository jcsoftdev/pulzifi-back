"use client";

import {
	type PointerEvent as ReactPointerEvent,
	useCallback,
	useEffect,
	useRef,
	useState,
} from "react";
import {
	AUTO_COLLAPSE_DELAY,
	AUTO_EXPAND_DELAY,
	SWIPE_DISMISS_THRESHOLD,
} from "../../constants";
import type { ToastData } from "../../domain/entities/types";
import { notix } from "../api";
import { DefaultToast } from "./default-toast";

interface ToastItemProps {
	toast: ToastData;
	canExpand: boolean;
	onMouseEnter: () => void;
	onMouseLeave: () => void;
}

export function ToastItem({
	toast,
	canExpand,
	onMouseEnter,
	onMouseLeave,
}: Readonly<ToastItemProps>) {
	const ref = useRef<HTMLDivElement>(null);
	const [isExpanded, setIsExpanded] = useState(false);
	const pointerStartX = useRef(0);
	const autoExpandTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
	const autoCollapseTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

	const toggle = useCallback(() => {
		setIsExpanded((prev) => !prev);
	}, []);

	const dismiss = useCallback(() => {
		notix.dismiss(toast.id);
	}, [toast.id]);

	// Autopilot expand/collapse
	useEffect(() => {
		if (!toast.description || !canExpand) return;

		const expandDelay = toast.autoExpandDelayMs ?? AUTO_EXPAND_DELAY;
		const collapseDelay = toast.autoCollapseDelayMs ?? AUTO_COLLAPSE_DELAY;

		autoExpandTimer.current = globalThis.setTimeout(() => {
			setIsExpanded(true);
		}, expandDelay);

		autoCollapseTimer.current = globalThis.setTimeout(() => {
			setIsExpanded(false);
		}, collapseDelay);

		return () => {
			clearTimeout(autoExpandTimer.current);
			clearTimeout(autoCollapseTimer.current);
		};
	}, [
		toast.description,
		toast.autoExpandDelayMs,
		toast.autoCollapseDelayMs,
		canExpand,
	]);

	// Swipe-to-dismiss
	const handlePointerDown = useCallback(
		(e: ReactPointerEvent<HTMLDivElement>) => {
			pointerStartX.current = e.clientX;
		},
		[],
	);

	const handlePointerUp = useCallback(
		(e: ReactPointerEvent<HTMLDivElement>) => {
			const dx = e.clientX - pointerStartX.current;
			if (Math.abs(dx) > SWIPE_DISMISS_THRESHOLD) {
				dismiss();
			}
		},
		[dismiss],
	);

	// Hover expand for toasts with description
	const handleMouseEnter = useCallback(() => {
		if (toast.description && canExpand) {
			setIsExpanded(true);
		}
		onMouseEnter();
	}, [toast.description, canExpand, onMouseEnter]);

	const handleMouseLeave = useCallback(() => {
		setIsExpanded(false);
		onMouseLeave();
	}, [onMouseLeave]);

	const handleKeyDown = useCallback(
		(e: React.KeyboardEvent<HTMLDivElement>) => {
			if (e.key === "Enter" || e.key === " ") {
				e.preventDefault();
				dismiss();
			}
		},
		[dismiss],
	);

	// Derive alignment from position
	const getAlign = () => {
		if (toast.position?.endsWith("right")) return "right";
		if (toast.position?.endsWith("left")) return "left";
		return "center";
	};

	const align = getAlign();

	// Headless render
	if (toast.render) {
		return (
			<div
				ref={ref}
				role="button"
				tabIndex={0}
				data-notix-toast
				data-state={toast.state}
				data-exiting={toast.exiting || undefined}
				onPointerDown={handlePointerDown}
				onPointerUp={handlePointerUp}
				onMouseEnter={handleMouseEnter}
				onMouseLeave={handleMouseLeave}
				onKeyDown={handleKeyDown}
			>
				{toast.render({
					toast,
					dismiss,
					isExpanded,
					toggle,
					lifecycle: toast.exiting ? "exiting" : "visible",
				})}
			</div>
		);
	}

	return (
		<div
			ref={ref}
			role="button"
			tabIndex={0}
			data-notix-toast
			data-state={toast.state}
			data-expanded={isExpanded || undefined}
			data-exiting={toast.exiting || undefined}
			className={toast.className}
			onPointerDown={handlePointerDown}
			onPointerUp={handlePointerUp}
			onMouseEnter={handleMouseEnter}
			onMouseLeave={handleMouseLeave}
			onKeyDown={handleKeyDown}
		>
			<DefaultToast
				toast={toast}
				isExpanded={isExpanded}
				align={align}
				exiting={toast.exiting}
				canExpand={canExpand}
			/>
		</div>
	);
}
