import { ANIMATION_DURATION, SPRING_EASING } from "../../constants";
import type { TriggerRect } from "../../domain/entities/types";
import type {
	AnimationKeyframes,
	IAnimationStrategy,
} from "../../domain/ports/animation";

export class SlideAnimationStrategy implements IAnimationStrategy {
	readonly mode = "slide" as const;

	enter(
		_element: HTMLElement,
		_triggerRect?: TriggerRect,
		_targetRect?: DOMRect,
	): AnimationKeyframes {
		return {
			keyframes: [
				{ transform: "translateY(-8px) scale(0.95)", opacity: 0 },
				{ transform: "translateY(0) scale(1)", opacity: 1 },
			],
			options: {
				duration: ANIMATION_DURATION,
				easing: SPRING_EASING,
				fill: "forwards",
			},
		};
	}

	exit(
		_element: HTMLElement,
		_triggerRect?: TriggerRect,
		_targetRect?: DOMRect,
	): AnimationKeyframes {
		return {
			keyframes: [
				{ transform: "translateY(0) scale(1)", opacity: 1 },
				{ transform: "translateY(-8px) scale(0.95)", opacity: 0 },
			],
			options: {
				duration: ANIMATION_DURATION * 0.6,
				easing: "ease-out",
				fill: "forwards",
			},
		};
	}
}
