import type { AnimationMode, TriggerRect } from "../entities/types";

export interface AnimationKeyframes {
	keyframes: Keyframe[];
	options: KeyframeAnimationOptions;
}

export interface IAnimationStrategy {
	readonly mode: AnimationMode;
	enter(
		element: HTMLElement,
		triggerRect?: TriggerRect,
		targetRect?: DOMRect,
	): AnimationKeyframes;
	exit(
		element: HTMLElement,
		triggerRect?: TriggerRect,
		targetRect?: DOMRect,
	): AnimationKeyframes;
}
