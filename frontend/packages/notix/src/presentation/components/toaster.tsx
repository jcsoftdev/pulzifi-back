"use client";

import { type CSSProperties, useCallback, useEffect, useMemo, useState } from "react";
import type { NotixOptions, NotixPosition, ToastData } from "../../domain/entities/types";
import { getGlobalManager } from "../api";
import { useToastStore } from "../hooks/use-toast";
import { ToastItem } from "./toast";

type OffsetValue = number | string;
type OffsetConfig = Partial<Record<"top" | "right" | "bottom" | "left", OffsetValue>>;

export interface ToasterProps {
	position?: NotixPosition;
	className?: string;
	offset?: OffsetValue | OffsetConfig;
	options?: Partial<NotixOptions>;
}

export function Toaster({
	position = "top-right",
	className,
	offset,
	options,
}: Readonly<ToasterProps>) {
	const toasts = useToastStore();
	const [activeId, setActiveId] = useState<string>();
	const manager = getGlobalManager();

	// Sync defaults
	useEffect(() => {
		manager.setDefaultPosition(position);
		if (options) manager.setDefaultOptions(options);
	}, [position, options, manager]);

	// Find the latest non-exiting toast
	const latest = useMemo(() => {
		for (let i = toasts.length - 1; i >= 0; i--) {
			const t = toasts[i];
			if (t && !t.exiting) return t.id;
		}
		return undefined;
	}, [toasts]);

	useEffect(() => {
		setActiveId(latest);
	}, [latest]);

	// Hover handlers: pause/resume timers
	const handleMouseEnter = useCallback(
		(toastId: string) => {
			setActiveId(toastId);
			manager.pauseTimers();
		},
		[manager],
	);

	const handleMouseLeave = useCallback(() => {
		setActiveId(latest);
		manager.resumeTimers();
	}, [latest, manager]);

	// Group by position
	const positionGroups = useMemo(() => {
		const map = new Map<NotixPosition, ToastData[]>();
		for (const t of toasts) {
			const pos = t.position ?? position;
			const arr = map.get(pos);
			if (arr) {
				arr.push(t);
			} else {
				map.set(pos, [t]);
			}
		}
		return map;
	}, [toasts, position]);

	const getViewportStyle = useCallback(
		(pos: NotixPosition): CSSProperties | undefined => {
			if (offset === undefined) return undefined;

			const o =
				typeof offset === "object"
					? offset
					: { top: offset, right: offset, bottom: offset, left: offset };

			const s: CSSProperties = {};
			const px = (v: OffsetValue) => (typeof v === "number" ? `${v}px` : v);

			if (pos.startsWith("top") && o.top != null) s.top = px(o.top);
			if (pos.startsWith("bottom") && o.bottom != null) s.bottom = px(o.bottom);
			if (pos.endsWith("left") && o.left != null) s.left = px(o.left);
			if (pos.endsWith("right") && o.right != null) s.right = px(o.right);

			return s;
		},
		[offset],
	);

	return (
		<>
			{Array.from(positionGroups, ([pos, items]) => (
				<section
					key={pos}
					data-notix-viewport
					data-position={pos}
					aria-live="polite"
					aria-atomic="false"
					aria-relevant="additions removals"
					role="region"
					aria-label="Notifications"
					className={className}
					style={getViewportStyle(pos)}
				>
					{items.map((item) => (
						<ToastItem
							key={item.id}
							toast={item}
							canExpand={activeId === undefined || activeId === item.id}
							onMouseEnter={() => handleMouseEnter(item.id)}
							onMouseLeave={handleMouseLeave}
						/>
					))}
				</section>
			))}
		</>
	);
}
