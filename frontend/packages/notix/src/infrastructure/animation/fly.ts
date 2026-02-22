import { ANIMATION_DURATION, SPRING_EASING } from "../../constants";
import type { TriggerRect } from "../../domain/entities/types";
import type {
	AnimationKeyframes,
	IAnimationStrategy,
} from "../../domain/ports/animation";
import { SlideAnimationStrategy } from "./slide";

const fallback = new SlideAnimationStrategy();

export class FlyAnimationStrategy implements IAnimationStrategy {
	readonly mode = "fly" as const;

	enter(
		element: HTMLElement,
		triggerRect?: TriggerRect,
		targetRect?: DOMRect,
	): AnimationKeyframes {
		if (!triggerRect || !targetRect) {
			return fallback.enter(element, triggerRect, targetRect);
		}

		const triggerCenterX = triggerRect.left + triggerRect.width / 2;
		const triggerCenterY = triggerRect.top + triggerRect.height / 2;
		const toastCenterX = targetRect.left + targetRect.width / 2;
		const toastCenterY = targetRect.top + targetRect.height / 2;

		const dx = triggerCenterX - toastCenterX;
		const dy = triggerCenterY - toastCenterY;

		return {
			keyframes: [
				{
					transform: `translate(${dx}px, ${dy}px) scale(0.8)`,
					opacity: 0,
				},
				{
					transform: "translate(0, 0) scale(1)",
					opacity: 1,
				},
			],
			options: {
				duration: ANIMATION_DURATION * 1.1,
				easing: SPRING_EASING,
				fill: "forwards",
			},
		};
	}

	exit(
		element: HTMLElement,
		triggerRect?: TriggerRect,
		targetRect?: DOMRect,
	): AnimationKeyframes {
		if (!triggerRect || !targetRect) {
			return fallback.exit(element, triggerRect, targetRect);
		}

		const triggerCenterX = triggerRect.left + triggerRect.width / 2;
		const triggerCenterY = triggerRect.top + triggerRect.height / 2;
		const toastCenterX = targetRect.left + targetRect.width / 2;
		const toastCenterY = targetRect.top + targetRect.height / 2;

		const dx = triggerCenterX - toastCenterX;
		const dy = triggerCenterY - toastCenterY;

		return {
			keyframes: [
				{
					transform: "translate(0, 0) scale(1)",
					opacity: 1,
				},
				{
					transform: `translate(${dx}px, ${dy}px) scale(0.8)`,
					opacity: 0,
				},
			],
			options: {
				duration: ANIMATION_DURATION * 0.7,
				easing: "ease-in",
				fill: "forwards",
			},
		};
	}
}
