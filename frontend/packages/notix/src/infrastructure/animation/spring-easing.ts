import { SPRING_EASING } from "../../constants";
import type { AnimationMode } from "../../domain/entities/types";
import type { IAnimationStrategy } from "../../domain/ports/animation";
import { FlyAnimationStrategy } from "./fly";
import { MorphAnimationStrategy } from "./morph";
import { SlideAnimationStrategy } from "./slide";

export { SPRING_EASING };

const strategies: Record<AnimationMode, () => IAnimationStrategy> = {
	slide: () => new SlideAnimationStrategy(),
	morph: () => new MorphAnimationStrategy(),
	fly: () => new FlyAnimationStrategy(),
};

export function resolveStrategy(mode: AnimationMode): IAnimationStrategy {
	return strategies[mode]();
}
