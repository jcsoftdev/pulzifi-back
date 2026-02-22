import { ANIMATION_DURATION, SPRING_EASING } from "../../constants";
import type { TriggerRect } from "../../domain/entities/types";
import type {
	AnimationKeyframes,
	IAnimationStrategy,
} from "../../domain/ports/animation";
import { SlideAnimationStrategy } from "./slide";

const fallback = new SlideAnimationStrategy();

export class MorphAnimationStrategy implements IAnimationStrategy {
	readonly mode = "morph" as const;

	enter(
		element: HTMLElement,
		triggerRect?: TriggerRect,
		targetRect?: DOMRect,
	): AnimationKeyframes {
		if (!triggerRect || !targetRect) {
			return fallback.enter(element, triggerRect, targetRect);
		}

		const scaleX = triggerRect.width / targetRect.width;
		const scaleY = triggerRect.height / targetRect.height;
		const translateX =
			triggerRect.left -
			targetRect.left +
			(triggerRect.width - targetRect.width) / 2;
		const translateY =
			triggerRect.top -
			targetRect.top +
			(triggerRect.height - targetRect.height) / 2;

		return {
			keyframes: [
				{
					transform: `translate(${translateX}px, ${translateY}px) scale(${scaleX}, ${scaleY})`,
					opacity: 0.4,
					borderRadius: "9999px",
				},
				{
					transform: "translate(0, 0) scale(1, 1)",
					opacity: 1,
					borderRadius: "",
				},
			],
			options: {
				duration: ANIMATION_DURATION * 1.25,
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

		const scaleX = triggerRect.width / targetRect.width;
		const scaleY = triggerRect.height / targetRect.height;
		const translateX =
			triggerRect.left -
			targetRect.left +
			(triggerRect.width - targetRect.width) / 2;
		const translateY =
			triggerRect.top -
			targetRect.top +
			(triggerRect.height - targetRect.height) / 2;

		return {
			keyframes: [
				{
					transform: "translate(0, 0) scale(1, 1)",
					opacity: 1,
				},
				{
					transform: `translate(${translateX}px, ${translateY}px) scale(${scaleX}, ${scaleY})`,
					opacity: 0,
				},
			],
			options: {
				duration: ANIMATION_DURATION * 0.8,
				easing: "ease-in",
				fill: "forwards",
			},
		};
	}
}
