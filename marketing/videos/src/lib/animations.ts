import { interpolate, spring, Easing } from "remotion";
import type { SpringConfig } from "remotion";

const CLAMP = {
  extrapolateLeft: "clamp",
  extrapolateRight: "clamp",
} as const;

// ── Fade ─────────────────────────────────────────────────────────────────────

export function fadeIn(frame: number, durationInFrames: number): number {
  return interpolate(frame, [0, durationInFrames], [0, 1], CLAMP);
}

export function fadeOut(frame: number, durationInFrames: number): number {
  return interpolate(frame, [0, durationInFrames], [1, 0], CLAMP);
}

/** Multi-keyframe fade in → hold → fade out */
export function fadeInHoldFadeOut(
  frame: number,
  inEnd: number,
  holdEnd: number,
  outEnd: number
): number {
  return interpolate(frame, [0, inEnd, holdEnd, outEnd], [0, 1, 1, 0], CLAMP);
}

// ── Slide ────────────────────────────────────────────────────────────────────

export function slideInY(
  frame: number,
  durationInFrames: number,
  offsetPx = 40
): number {
  return interpolate(frame, [0, durationInFrames], [offsetPx, 0], CLAMP);
}

export function slideInX(
  frame: number,
  durationInFrames: number,
  offsetPx = 40
): number {
  return interpolate(frame, [0, durationInFrames], [offsetPx, 0], CLAMP);
}

// ── Scale ────────────────────────────────────────────────────────────────────

export function scaleIn(
  frame: number,
  durationInFrames: number,
  from = 0.85
): number {
  return interpolate(frame, [0, durationInFrames], [from, 1], CLAMP);
}

// ── Spring ───────────────────────────────────────────────────────────────────

export function springEnter(
  frame: number,
  fps: number,
  config: Partial<SpringConfig> = { damping: 12 }
): number {
  return spring({ fps, frame, config });
}

export interface SpringEnterExitOptions {
  fps: number;
  frame: number;
  durationInFrames: number;
  /** Frames before exit spring starts. Default: 20 */
  exitBuffer?: number;
  enterConfig?: Partial<SpringConfig>;
  exitConfig?: Partial<SpringConfig>;
}

/**
 * Returns a 0→1→0 value using enter spring minus exit spring.
 * Use for elements that should enter and exit within a Sequence.
 */
export function springEnterExit({
  fps,
  frame,
  durationInFrames,
  exitBuffer = 20,
  enterConfig = { damping: 12 },
  exitConfig = { damping: 12 },
}: SpringEnterExitOptions): number {
  const enter = spring({ fps, frame, config: enterConfig });
  const exit = spring({
    fps,
    frame: frame - (durationInFrames - exitBuffer),
    config: exitConfig,
  });
  return Math.max(0, enter - exit);
}

// ── Text reveal ───────────────────────────────────────────────────────────────

export function textRevealOpacity(
  frame: number,
  index: number,
  staggerFrames = 4,
  revealDuration = 10
): number {
  const start = index * staggerFrames;
  return interpolate(frame, [start, start + revealDuration], [0, 1], CLAMP);
}

export function textRevealY(
  frame: number,
  index: number,
  staggerFrames = 4,
  revealDuration = 10,
  offsetPx = 20
): number {
  const start = index * staggerFrames;
  return interpolate(
    frame,
    [start, start + revealDuration],
    [offsetPx, 0],
    CLAMP
  );
}

// ── Easing presets ────────────────────────────────────────────────────────────

export const EASINGS = {
  easeOut: Easing.out(Easing.cubic),
  easeIn: Easing.in(Easing.cubic),
  easeInOut: Easing.inOut(Easing.cubic),
  bounce: Easing.bounce,
} as const;
