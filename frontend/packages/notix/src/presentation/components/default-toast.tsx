"use client";

import { motion } from "motion/react";
import {
	type CSSProperties,
	type ReactNode,
	memo,
	useCallback,
	useEffect,
	useLayoutEffect,
	useMemo,
	useRef,
	useState,
} from "react";
import {
	BLUR_RATIO,
	DEFAULT_ROUNDNESS,
	HEADER_EXIT_MS,
	MIN_EXPAND_RATIO,
	PILL_PADDING,
	SPRING,
	SWAP_COLLAPSE_MS,
	TOAST_HEIGHT as HEIGHT,
	TOAST_WIDTH as WIDTH,
} from "../../constants";
import type { ToastData, ToastState } from "../../domain/entities/types";
import {
	ArrowRightIcon,
	CheckIcon,
	CircleAlertIcon,
	InfoIcon,
	LifeBuoyIcon,
	LoaderCircleIcon,
	XIcon,
} from "./icons";
import { GooeyDefs } from "./gooey-defs";

/* ---------------------------------- Icons --------------------------------- */

const STATE_ICONS: Record<ToastState, ReactNode> = {
	success: <CheckIcon />,
	error: <XIcon />,
	warning: <CircleAlertIcon />,
	info: <InfoIcon />,
	loading: <LoaderCircleIcon data-notix-icon="spin" />,
	action: <LifeBuoyIcon />,
};

/* ---------------------------------- View ---------------------------------- */

interface View {
	title: string;
	description?: ReactNode | string;
	state: ToastState;
	icon?: ReactNode | null;
	styles?: ToastData["styles"];
	button?: ToastData["button"];
	fill: string;
}

/* ------------------------------- Component -------------------------------- */

interface DefaultToastProps {
	toast: ToastData;
	isExpanded: boolean;
	align?: "left" | "center" | "right";
	ready?: boolean;
	exiting?: boolean;
	canExpand?: boolean;
}

export const DefaultToast = memo(function DefaultToast({
	toast,
	isExpanded,
	align = "right",
	ready: readyProp,
	exiting = false,
	canExpand = true,
}: DefaultToastProps) {
	const fill = "#FFFFFF";
	const id = toast.instanceId;

	const next: View = useMemo(
		() => ({
			title: toast.title,
			description: toast.description,
			state: toast.state,
			icon: toast.icon,
			styles: toast.styles,
			button: toast.button,
			fill,
		}),
		[toast.title, toast.description, toast.state, toast.icon, toast.styles, toast.button],
	);

	const [view, setView] = useState<View>(next);
	const [localExpanded, setLocalExpanded] = useState(false);
	const [ready, setReady] = useState(false);
	const [pillWidth, setPillWidth] = useState(0);
	const [contentHeight, setContentHeight] = useState(0);

	const hasDesc = Boolean(view.description) || Boolean(view.button);
	const isLoading = view.state === "loading";
	const open = hasDesc && (isExpanded || localExpanded) && !isLoading && canExpand;

	const headerKey = `${view.state}-${view.title}`;
	const filterId = `notix-gooey-${id}`;
	const resolvedRoundness = DEFAULT_ROUNDNESS;
	const blur = resolvedRoundness * BLUR_RATIO;

	const headerRef = useRef<HTMLDivElement>(null);
	const contentRef = useRef<HTMLDivElement>(null);
	const innerRef = useRef<HTMLDivElement>(null);
	const headerExitRef = useRef<number | null>(null);
	const swapTimerRef = useRef<number | null>(null);
	const pendingRef = useRef<{ payload: View } | null>(null);

	const [headerLayer, setHeaderLayer] = useState<{
		current: { key: string; view: View };
		prev: { key: string; view: View } | null;
	}>({ current: { key: headerKey, view }, prev: null });

	/* ------------------------------ Measurements ------------------------------ */

	const headerPadRef = useRef<number | null>(null);
	const pillRoRef = useRef<ResizeObserver | null>(null);
	const pillRafRef = useRef(0);
	const pillObservedRef = useRef<Element | null>(null);

	// biome-ignore lint/correctness/useExhaustiveDependencies: headerLayer.current.key triggers re-measure
	useLayoutEffect(() => {
		const el = innerRef.current;
		const header = headerRef.current;
		if (!el || !header) return;
		if (headerPadRef.current === null) {
			const cs = getComputedStyle(header);
			headerPadRef.current =
				parseFloat(cs.paddingLeft) + parseFloat(cs.paddingRight);
		}
		const px = headerPadRef.current;
		const measure = () => {
			const w = el.scrollWidth + px + PILL_PADDING;
			if (w > PILL_PADDING) {
				setPillWidth((prev) => (prev === w ? prev : w));
			}
		};
		measure();

		if (!pillRoRef.current) {
			pillRoRef.current = new ResizeObserver(() => {
				cancelAnimationFrame(pillRafRef.current);
				pillRafRef.current = requestAnimationFrame(() => {
					const inner = innerRef.current;
					const pad = headerPadRef.current ?? 0;
					if (!inner) return;
					const w = inner.scrollWidth + pad + PILL_PADDING;
					if (w > PILL_PADDING) {
						setPillWidth((prev) => (prev === w ? prev : w));
					}
				});
			});
		}

		if (pillObservedRef.current !== el) {
			if (pillObservedRef.current) {
				pillRoRef.current.unobserve(pillObservedRef.current);
			}
			pillRoRef.current.observe(el);
			pillObservedRef.current = el;
		}
	}, [headerLayer.current.key]);

	useEffect(() => {
		return () => {
			cancelAnimationFrame(pillRafRef.current);
			pillRoRef.current?.disconnect();
		};
	}, []);

	// Content height measurement
	useLayoutEffect(() => {
		if (!hasDesc) {
			setContentHeight(0);
			return;
		}
		const el = contentRef.current;
		if (!el) return;
		const measure = () => {
			const h = el.scrollHeight;
			setContentHeight((prev) => (prev === h ? prev : h));
		};
		measure();
		let rafId = 0;
		const ro = new ResizeObserver(() => {
			cancelAnimationFrame(rafId);
			rafId = requestAnimationFrame(measure);
		});
		ro.observe(el);
		return () => {
			cancelAnimationFrame(rafId);
			ro.disconnect();
		};
	}, [hasDesc]);

	// Ready after first frame
	useEffect(() => {
		if (readyProp !== undefined) {
			setReady(readyProp);
			return;
		}
		const raf = requestAnimationFrame(() => setReady(true));
		return () => cancelAnimationFrame(raf);
	}, [readyProp]);

	/* ----------------------------- Header layers ------------------------------ */

	useLayoutEffect(() => {
		setHeaderLayer((state) => {
			if (state.current.key === headerKey) {
				if (state.current.view === view) return state;
				return { ...state, current: { key: headerKey, view } };
			}
			return {
				prev: state.current,
				current: { key: headerKey, view },
			};
		});
	}, [headerKey, view]);

	useEffect(() => {
		if (!headerLayer.prev) return;
		if (headerExitRef.current) {
			clearTimeout(headerExitRef.current);
		}
		headerExitRef.current = globalThis.setTimeout(() => {
			headerExitRef.current = null;
			setHeaderLayer((state) => ({ ...state, prev: null }));
		}, HEADER_EXIT_MS);
		return () => {
			if (headerExitRef.current) {
				clearTimeout(headerExitRef.current);
				headerExitRef.current = null;
			}
		};
	}, [headerLayer.prev]);

	/* ----------------------------- Refresh logic ------------------------------ */

	useEffect(() => {
		if (swapTimerRef.current) {
			clearTimeout(swapTimerRef.current);
			swapTimerRef.current = null;
		}

		if (open) {
			pendingRef.current = { payload: next };
			setLocalExpanded(false);
			swapTimerRef.current = globalThis.setTimeout(() => {
				swapTimerRef.current = null;
				const pending = pendingRef.current;
				if (!pending) return;
				setView(pending.payload);
				pendingRef.current = null;
			}, SWAP_COLLAPSE_MS);
		} else {
			pendingRef.current = null;
			setView(next);
		}
		// Only refresh when the view payload actually changes
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [next]);

	/* ------------------------------ Derived values ---------------------------- */

	const minExpanded = HEIGHT * MIN_EXPAND_RATIO;
	const rawExpanded = hasDesc
		? Math.max(minExpanded, HEIGHT + contentHeight)
		: minExpanded;

	const frozenExpandedRef = useRef(rawExpanded);
	if (open) {
		frozenExpandedRef.current = rawExpanded;
	}

	const expanded = open ? rawExpanded : frozenExpandedRef.current;
	const svgHeight = hasDesc ? Math.max(expanded, minExpanded) : HEIGHT;
	const expandedContent = Math.max(0, expanded - HEIGHT);
	const resolvedPillWidth = Math.max(pillWidth || HEIGHT, HEIGHT);
	const pillHeight = HEIGHT + blur * 3;

	const pillX =
		align === "right"
			? WIDTH - resolvedPillWidth
			: align === "center"
				? (WIDTH - resolvedPillWidth) / 2
				: 0;

	/* ------------------------------- Animate targets -------------------------- */

	const pillAnimate = useMemo(
		() => ({
			x: pillX,
			width: resolvedPillWidth,
			height: open ? pillHeight : HEIGHT,
		}),
		[pillX, resolvedPillWidth, open, pillHeight],
	);

	const bodyAnimate = useMemo(
		() => ({
			height: open ? expandedContent : 0,
			opacity: open ? 1 : 0,
		}),
		[open, expandedContent],
	);

	const bodyTransition = useMemo(
		() => (open ? SPRING : { ...SPRING, bounce: 0 }),
		[open],
	);

	const pillTransition = useMemo(
		() => (ready ? SPRING : { duration: 0 }),
		[ready],
	);

	const viewBox = `0 0 ${WIDTH} ${svgHeight}`;

	const canvasStyle = useMemo<CSSProperties>(
		() => ({ filter: `url(#${filterId})` }),
		[filterId],
	);

	/* ------------------------------- Inline styles ---------------------------- */

	const rootStyle = useMemo<CSSProperties & Record<string, string>>(
		() => ({
			"--_h": `${open ? expanded : HEIGHT}px`,
			"--_pw": `${resolvedPillWidth}px`,
			"--_px": `${pillX}px`,
			"--_ht": `translateY(${open ? 3 : 0}px) scale(${open ? 0.9 : 1})`,
			"--_co": `${open ? 1 : 0}`,
		}),
		[open, expanded, resolvedPillWidth, pillX],
	);

	/* -------------------------------- Handlers -------------------------------- */

	const handleButtonClick = useCallback(
		(e: React.MouseEvent) => {
			e.preventDefault();
			e.stopPropagation();
			view.button?.onClick();
		},
		[view.button],
	);

	/* --------------------------------- Render --------------------------------- */

	return (
		<div
			data-notix-default
			data-ready={ready}
			data-expanded={open}
			data-exiting={exiting}
			data-state={view.state}
			data-align={align}
			className={toast.styles?.toast}
			style={rootStyle}
		>
			{/* SVG Canvas with gooey filter */}
			<div data-notix-canvas style={canvasStyle}>
				<svg data-notix-svg width={WIDTH} height={svgHeight} viewBox={viewBox}>
					<title>Notification</title>
					<GooeyDefs filterId={filterId} blur={blur} />
					<motion.rect
						data-notix-pill-rect
						rx={resolvedRoundness}
						ry={resolvedRoundness}
						fill={view.fill}
						initial={false}
						animate={pillAnimate}
						transition={pillTransition}
					/>
					<motion.rect
						data-notix-body-rect
						y={HEIGHT}
						width={WIDTH}
						rx={resolvedRoundness}
						ry={resolvedRoundness}
						fill={view.fill}
						initial={false}
						animate={bodyAnimate}
						transition={bodyTransition}
					/>
				</svg>
			</div>

			{/* Header (badge + title) overlaid on pill */}
			<div ref={headerRef} data-notix-header>
				<div data-notix-header-stack>
					<div
						ref={innerRef}
						key={headerLayer.current.key}
						data-notix-header-inner
						data-layer="current"
					>
						<div
							data-notix-badge
							data-state={headerLayer.current.view.state}
							className={headerLayer.current.view.styles?.badge}
							aria-hidden="true"
						>
							{headerLayer.current.view.icon ??
								STATE_ICONS[headerLayer.current.view.state]}
						</div>
						<span
							data-notix-title
							data-state={headerLayer.current.view.state}
							className={headerLayer.current.view.styles?.title}
						>
							{headerLayer.current.view.title}
						</span>
					</div>
					{headerLayer.prev && (
						<div
							key={headerLayer.prev.key}
							data-notix-header-inner
							data-layer="prev"
							data-exiting="true"
						>
							<div
								data-notix-badge
								data-state={headerLayer.prev.view.state}
								className={headerLayer.prev.view.styles?.badge}
								aria-hidden="true"
							>
								{headerLayer.prev.view.icon ??
									STATE_ICONS[headerLayer.prev.view.state]}
							</div>
							<span
								data-notix-title
								data-state={headerLayer.prev.view.state}
								className={headerLayer.prev.view.styles?.title}
							>
								{headerLayer.prev.view.title}
							</span>
						</div>
					)}
				</div>

			</div>

			{/* Expandable content */}
			{hasDesc && (
				<div data-notix-content data-visible={open}>

					<div
						ref={contentRef}
						data-notix-description
						className={view.styles?.description}
					>
						{view.description}
						{view.button && (
							<button
								type="button"
								data-notix-button
								data-state={view.state}
								className={view.styles?.button}
								onClick={handleButtonClick}
							>
								{view.button.title}
								<ArrowRightIcon />
							</button>
						)}
					</div>
				</div>
			)}
		</div>
	);
});
