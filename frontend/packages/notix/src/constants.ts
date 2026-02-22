// ─── Layout ──────────────────────────────────────────────────────────────────

export const TOAST_WIDTH = 350;
export const TOAST_HEIGHT = 40;
export const DEFAULT_ROUNDNESS = 16;

// ─── Timing ──────────────────────────────────────────────────────────────────

export const DURATION_MS = 600;
export const DURATION_S = DURATION_MS / 1000;

export const DEFAULT_TOAST_DURATION = 6000;
export const EXIT_DURATION = 300;
export const ANIMATION_DURATION = 400;
export const AUTO_EXPAND_DELAY = DEFAULT_TOAST_DURATION * 0.025;
export const AUTO_COLLAPSE_DELAY = DEFAULT_TOAST_DURATION - 2000;

// ─── Interaction ─────────────────────────────────────────────────────────────

export const SWIPE_DISMISS_THRESHOLD = 30;

// ─── Gooey / SVG ────────────────────────────────────────────────────────────

export const BLUR_RATIO = 0.5;
export const PILL_PADDING = 10;
export const MIN_EXPAND_RATIO = 2.25;
export const SWAP_COLLAPSE_MS = 200;
export const HEADER_EXIT_MS = DURATION_MS * 0.7;

export const SPRING = {
	type: "spring" as const,
	bounce: 0.25,
	duration: DURATION_S,
};

// ─── Spring Easing ───────────────────────────────────────────────────────────

export const SPRING_EASING = `linear(
	0,
	0.002 0.6%,
	0.007 1.2%,
	0.015 1.8%,
	0.026 2.4%,
	0.041 3.1%,
	0.06 3.8%,
	0.108 5.3%,
	0.157 6.6%,
	0.214 8%,
	0.467 13.7%,
	0.577 16.3%,
	0.631 17.7%,
	0.682 19.1%,
	0.73 20.5%,
	0.771 21.8%,
	0.808 23.1%,
	0.844 24.5%,
	0.874 25.8%,
	0.903 27.2%,
	0.928 28.6%,
	0.952 30.1%,
	0.972 31.6%,
	0.988 33.1%,
	1.01 35.7%,
	1.025 38.5%,
	1.034 41.6%,
	1.038 45%,
	1.035 50.1%,
	1.012 64.2%,
	1.003 73%,
	0.999 83.7%,
	1
)`;
