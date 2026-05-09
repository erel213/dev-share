---
paths:
  - "marketing/videos/src/**/*.ts"
  - "marketing/videos/src/**/*.tsx"
  - "marketing/videos/remotion.config.ts"
  - "marketing/videos/package.json"
---

# Remotion Video Rules

All marketing video compositions live in `marketing/videos/src/`. Follow these rules when creating or modifying any file in this directory.

## Project Structure

```
src/
  Root.tsx              — register every <Composition> here; nowhere else
  index.ts              — Remotion entry point; do not modify
  compositions/         — one file per composition, exports component + zod schema
  components/           — reusable Remotion components; shared across compositions
  lib/
    animations.ts       — all animation helpers; extend here, never inline raw interpolate()
    theme.ts            — all design tokens; never hardcode colors, fonts, or radii
```

## Composition Rules

### Every composition file must export
1. A `z.object(...)` schema named `<Name>Schema` (camelCase with lowercase first char) — e.g. `devShareDemoSchema`
2. A `type <Name>Props = z.infer<typeof <name>Schema>` type
3. A `React.FC<<Name>Props>` component with a matching name

### Scene components
- Define each scene as a private `const SceneName: React.FC<...> = () => {}` in the same file
- Keep scene components small — extract sub-parts if a scene exceeds ~60 lines
- Scene components receive only the props they need; pull `frame` and `fps` from hooks

### Sequencing
- Use `<Series>` + `<Series.Sequence durationInFrames={N} name="Label">` for linear timelines
- Use named `<Sequence from={N} durationInFrames={N} name="Label">` when scenes overlap or a sub-element needs a local frame offset
- Always set the `name` prop — it shows in Remotion Studio's timeline
- `durationInFrames` in `Root.tsx` must equal the sum of all top-level sequence durations

### Props and schema
- All user-facing configuration goes through the zod schema (no magic constants in JSX)
- Always provide `.default()` for every schema field so Studio works without explicit props
- Do not use optional (`z.optional`) fields — use defaults instead

## Animation Rules

**Never** use CSS `transition`, `animation`, `@keyframes`, or `keyframes()` from any library. Remotion renders frame-by-frame; CSS animations are non-deterministic across frames.

**Always** use helpers from `src/lib/animations.ts`:

| Need | Use |
|------|-----|
| Fade in | `fadeIn(frame, durationInFrames)` |
| Fade out | `fadeOut(frame, durationInFrames)` |
| Fade in → hold → fade out | `fadeInHoldFadeOut(frame, inEnd, holdEnd, outEnd)` |
| Slide from below | `slideInY(frame, duration, offsetPx?)` → pass result to `translateY(...)` |
| Slide from side | `slideInX(frame, duration, offsetPx?)` → pass result to `translateX(...)` |
| Scale entry | `scaleIn(frame, duration, from?)` → pass result to `scale(...)` |
| Springy entry | `springEnter(frame, fps, config?)` → use for scale or opacity |
| Springy entry + exit | `springEnterExit({ fps, frame, durationInFrames, ... })` |
| Staggered word/letter reveal opacity | `textRevealOpacity(frame, index, staggerFrames?, revealDuration?)` |
| Staggered word/letter reveal slide | `textRevealY(frame, index, staggerFrames?, revealDuration?)` |

`frame` inside a `<Series.Sequence>` is always local (starts at 0 when the sequence starts). This is intentional — use it directly.

If you need a new animation primitive, add it to `animations.ts` as an exported function. Do not inline `interpolate()` or `spring()` in component files.

## Theme Rules

**Never** hardcode colors, font stacks, border radii, or brand strings. Always import from `src/lib/theme.ts`:

```ts
import { COLORS, FONTS, RADIUS, BRAND } from "../lib/theme";
```

All inline styles must reference these tokens. The tokens mirror the frontend's dark-mode OKLCH CSS variables — keep them in sync when the frontend theme changes.

If you need a new token, add it to `theme.ts`. Do not use Tailwind classes (they are not available in Remotion).

## Asset Rules

### Loading assets
- Always use `staticFile("path/relative/to/public/")` — never relative paths or absolute URLs
- Screenshots: `staticFile("screenshots/filename.png")`
- Recordings: `staticFile("recordings/filename.webm")`
- Use `<Img>` (from remotion) for images — never `<img>`
- Use `<OffthreadVideo>` (from remotion) for video files — never `<video>`

### Asset source
Assets in `public/` are captured by `e2e/docs-capture/marketing.spec.ts` using Playwright:
- Screenshots → `docs/static/images/marketing/*.png` → copy to `public/screenshots/`
- Recordings → `docs/static/images/marketing/*.webm` → copy to `public/recordings/`

When a composition references an asset that does not yet exist in `public/`, add the corresponding capture test to `e2e/docs-capture/marketing.spec.ts` before writing the composition code.

## Component Rules

Use the shared components from `src/components/` — do not duplicate their logic inline:

| Component | Use for |
|-----------|---------|
| `<AnimatedText text={...} style={...} staggerFrames={4} revealDuration={12} mode="words\|letters" />` | Any animated headline or body text |
| `<BrowserFrame url="localhost:3000/path" animate={true\|false}>` | Wrapping screenshots or video in a browser chrome |
| `<CodeWindow lines={[...]} title="file.ts" framesPerLine={6} />` | Animated syntax-highlighted code reveal |
| `<TerminalOutput lines={[...]} framesPerChar={4} pausePerLine={15} />` | Typewriter terminal output |

Do not add new generic components to individual composition files. If a component will be reused, add it to `src/components/`.

## Root.tsx Rules

- Every `<Composition>` entry must have: `id`, `component`, `width`, `height`, `fps`, `durationInFrames`, `schema`, `defaultProps`
- `id` is the render target name — use PascalCase, match the component name
- Standard dimensions: `1920×1080` (landscape), `1080×1920` (portrait/social)
- Standard fps: `30`
- `durationInFrames` must match the sum of top-level sequences in the composition

## package.json Rules

Every composition registered in `Root.tsx` must have a corresponding `render:<kebab-name>` script:

```json
"render:my-video": "remotion render src/index.ts MyVideo out/my-video.mp4"
```

## DO

- Keep scene components pure — derive everything from `frame`, `fps`, and props
- Use `<AbsoluteFill>` as the root of every scene; set `backgroundColor` explicitly
- Wrap screenshot assets in `<BrowserFrame>` for visual consistency
- Add `name` props to all `<Sequence>` and `<Series.Sequence>` elements
- Calculate frame budgets in comments: `// Total: 90 + 120 + 180 = 390 frames = 13s`

## DON'T

- Don't use `useEffect`, `useState`, or `useRef` — Remotion renders are pure; side effects break frame determinism
- Don't use `Math.random()` or `Date.now()` — non-deterministic; use frame-derived values only
- Don't use `position: fixed` — use `<AbsoluteFill>` instead
- Don't import from outside `marketing/videos/src/` or `remotion` packages — compositions must be self-contained
- Don't put render-time logic (file I/O, API calls) inside components
- Don't register a composition in `Root.tsx` without a render script in `package.json`
