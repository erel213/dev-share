---
name: create-marketing-video
description: >
  Create or update a Remotion marketing video composition for Dev Share.
  Triggers when: adding a new marketing video or social clip, adding a Remotion
  composition or scene, capturing screenshots/recordings for video use,
  or rendering a marketing video to MP4.
allowed-tools: Read Write Edit Bash Glob Grep
---

# Create Marketing Video

Marketing videos live in `marketing/videos/` (Remotion project). Assets (screenshots, recordings) are captured with Playwright from the running app and served as static files.

## Directory Map

```
marketing/videos/
  src/
    Root.tsx                  — registers all <Composition> entries
    index.ts                  — Remotion entry point
    compositions/             — one file per video (DevShareDemo, FeatureWalkthrough, SocialClip)
    components/               — reusable Remotion components (BrowserFrame, AnimatedText, CodeWindow, TerminalOutput)
    lib/
      animations.ts           — fadeIn, fadeOut, springEnter, springEnterExit, textReveal*, slideIn*, EASINGS
      theme.ts                — COLORS, FONTS, RADIUS, BRAND constants (mirror frontend dark-mode tokens)
  public/
    screenshots/              — PNG stills → referenced as staticFile('screenshots/name.png')
    recordings/               — WEBM videos → referenced as staticFile('recordings/name.webm')
    fonts/
    music/
  remotion.config.ts          — codec h264, output ./out, concurrency 4, @/* alias → src/
  package.json                — render:demo, render:walkthrough, render:social scripts

e2e/docs-capture/
  marketing.spec.ts           — @marketing Playwright specs (screenshots + recordings)
  helpers.ts                  — login(), screenshotPage(), recordVideo(), DOCS_IMAGE_ROOT
  pages.spec.ts               — @docs pages (non-marketing screenshots)
  workflows.spec.ts           — @docs workflows (non-marketing recordings)

docs/static/images/           — DOCS_IMAGE_ROOT (Playwright saves here)
  marketing/                  — marketing captures land here; must be synced to public/
```

## Step 1: Determine What to Capture

Identify which app screens and workflows the video needs. Check `marketing.spec.ts` for what already exists:

```ts
// Existing marketing captures:
// screenshots/hero-dashboard.png          → screenshotPage at /
// screenshots/environments-hover.png      → screenshotPage at /environments (hover state)
// screenshots/create-environment-dialog.png → dialog open state
// recordings/create-environment-workflow.webm → recordVideo of create-environment flow
```

For each new piece of content, decide:
- **Static screenshot** → use `screenshotPage()` or `page.screenshot()` for complex states
- **Workflow recording** → use `recordVideo()` which produces a WEBM

## Step 2: Add Capture Specs to marketing.spec.ts

Add tests to `e2e/docs-capture/marketing.spec.ts`. All tests must be inside `test.describe("@marketing", ...)`.

### Screenshot of a page

```ts
test("my page — description", async ({ page }) => {
  await screenshotPage(page, {
    path: "/my-page",
    outputPath: "marketing/my-screenshot.png",
    waitFor: "Page Heading",        // optional: wait for visible text
    prepare: async (page) => {      // optional: set up hover/focus/etc.
      await page.locator("button").first().hover();
      await page.waitForTimeout(200);
    },
  });
});
```

### Screenshot of a dialog (must use raw `page.screenshot()`)

```ts
test("my dialog open", async ({ page }) => {
  await page.goto("/my-page");
  await page.waitForLoadState("networkidle");
  await page.getByRole("button", { name: "Open Dialog" }).click();
  await page.locator("#dialog-field").waitFor({ state: "visible" });

  const outputFile = path.join(DOCS_IMAGE_ROOT, "marketing/my-dialog.png");
  ensureDir(path.dirname(outputFile));
  await page.screenshot({ path: outputFile, fullPage: false });
});
```

### Workflow recording

```ts
test("my workflow", async ({ browser }) => {
  await recordVideo(browser, {
    outputPath: "marketing/my-workflow.webm",
    workflow: async (page) => {
      await login(page);
      await page.goto("/my-page");
      await page.waitForLoadState("networkidle");
      // perform steps...
      await page.waitForTimeout(1500); // pause at end so viewer sees result
    },
  });
});
```

## Step 3: Run Captures

**Prerequisite**: The app must be running at `http://localhost:3000` with data (templates, groups, etc. — the global setup in `global-setup.ts` seeds these automatically if absent).

```bash
# All @marketing captures
cd e2e && pnpm docs:capture:marketing

# Single test by name
cd e2e && npx playwright test --config=playwright.docs.config.ts --grep "my page"

# Headed mode to debug
cd e2e && pnpm docs:capture:debug
```

Override defaults:
```bash
DOCS_BASE_URL=http://localhost:5173 DOCS_ADMIN_EMAIL=admin@example.com DOCS_ADMIN_PASSWORD=secret pnpm docs:capture:marketing
```

Captures save to `docs/static/images/marketing/`. The `postdocs:capture:marketing` lifecycle hook in `e2e/package.json` automatically syncs all PNGs and WEBMs to `marketing/videos/public/` immediately after the Playwright run finishes — no manual copy needed.

## Step 4: Sync Assets to Remotion public/

**Automatic**: `pnpm docs:capture:marketing` triggers a post-script that runs:
```bash
cp ../docs/static/images/marketing/*.png ../marketing/videos/public/screenshots/
cp ../docs/static/images/marketing/*.webm ../marketing/videos/public/recordings/
```

Manual sync (if you captured a single file with `npx playwright test --grep ...` or need to re-sync):

Or sync only specific files:
```bash
cp docs/static/images/marketing/my-screenshot.png marketing/videos/public/screenshots/
cp docs/static/images/marketing/my-workflow.webm marketing/videos/public/recordings/
```

Reference in compositions:
```ts
import { Img, OffthreadVideo, staticFile } from "remotion";

// Screenshot
<Img src={staticFile("screenshots/my-screenshot.png")} style={{ width: "100%" }} />

// Recording
<OffthreadVideo src={staticFile("recordings/my-workflow.webm")} style={{ width: "100%" }} />
```

## Step 5: Create a Remotion Composition

### 5a. Create the composition file

Create `marketing/videos/src/compositions/MyVideo.tsx`. Every composition follows this structure:

```tsx
import React from "react";
import { z } from "zod";
import { AbsoluteFill, Series, useCurrentFrame, useVideoConfig, Img, OffthreadVideo, staticFile } from "remotion";
import { springEnter, fadeIn, fadeOut, slideInY } from "../lib/animations";
import { COLORS, FONTS, BRAND, RADIUS } from "../lib/theme";
import { AnimatedText } from "../components/AnimatedText";
import { BrowserFrame } from "../components/BrowserFrame";

export const myVideoSchema = z.object({
  title: z.string().default("My Video Title"),
  // add configurable props here
});

export type MyVideoProps = z.infer<typeof myVideoSchema>;

// ── Scene components ────────────────────────────────────────────────────────

const IntroScene: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  // frame 0..N for this sequence only (Series.Sequence resets frame)
  return (
    <AbsoluteFill style={{ backgroundColor: COLORS.background, alignItems: "center", justifyContent: "center" }}>
      {/* use springEnter, fadeIn from lib/animations — never CSS transitions/keyframes */}
    </AbsoluteFill>
  );
};

// ── Composition root ──────────────────────────────────────────────────────

// Frame budget: 30fps → 30 frames/s. 900 = 30s. Plan scene durations to add up.
export const MyVideo: React.FC<MyVideoProps> = ({ title }) => {
  return (
    <Series>
      <Series.Sequence durationInFrames={90} name="Intro">
        <IntroScene />
      </Series.Sequence>
      {/* add more sequences */}
    </Series>
  );
};
```

**Animation reference** (`lib/animations.ts`):
- `fadeIn(frame, durationInFrames)` → 0→1 opacity
- `fadeOut(frame, durationInFrames)` → 1→0 opacity
- `fadeInHoldFadeOut(frame, inEnd, holdEnd, outEnd)` → 0→1→1→0
- `slideInY(frame, duration, offsetPx?)` → vertical slide distance (pass to `translateY`)
- `slideInX(frame, duration, offsetPx?)` → horizontal slide distance
- `scaleIn(frame, duration, from?)` → scale value 0.85→1
- `springEnter(frame, fps, config?)` → spring 0→1 (use for scale/opacity)
- `springEnterExit({fps, frame, durationInFrames, ...})` → spring 0→1→0 within a sequence
- `textRevealOpacity(frame, index, staggerFrames?, revealDuration?)` → staggered word opacity
- `textRevealY(frame, index, staggerFrames?, revealDuration?)` → staggered word slide

**Component reference**:
- `<AnimatedText text="..." style={{...}} staggerFrames={4} revealDuration={12} mode="words|letters" />`
- `<BrowserFrame url="localhost:3000/path" animate={true|false}>...</BrowserFrame>` — mac-style chrome
- `<CodeWindow lines={[...]} title="file.ts" framesPerLine={6} />` — syntax-highlighted code
- `<TerminalOutput lines={[...]} framesPerChar={4} pausePerLine={15} />` — typewriter terminal

**Theme reference** (`lib/theme.ts`):
```ts
COLORS.background        // oklch(0.145 0 0)    dark bg
COLORS.foreground        // oklch(0.985 0 0)    near-white text
COLORS.card              // oklch(0.205 0 0)    card bg
COLORS.primary           // oklch(0.922 0 0)    primary button/logo
COLORS.primaryForeground // oklch(0.205 0 0)    text on primary
COLORS.muted             // oklch(0.269 0 0)    muted bg
COLORS.mutedForeground   // oklch(0.708 0 0)    muted text
COLORS.border            // oklch(1 0 0 / 10%)  subtle border
COLORS.accentBlue        // oklch(0.488 0.243 264.376)
BRAND.name               // "Dev Share"
BRAND.monogram           // "DS"
BRAND.tagline            // "Manage temporary developer environments with ease."
FONTS.sans / FONTS.mono
RADIUS.sm / .md / .lg / .xl / ["2xl"]
```

### 5b. Register in Root.tsx

Add a `<Composition>` entry to `marketing/videos/src/Root.tsx`:

```tsx
import { MyVideo, myVideoSchema } from "./compositions/MyVideo";

// Inside RemotionRoot:
<Composition
  id="MyVideo"           // used in render commands
  component={MyVideo}
  width={1920}           // 1920×1080 landscape, or 1080×1920 portrait for social
  height={1080}
  fps={30}
  durationInFrames={900} // must match sum of scene durations
  schema={myVideoSchema}
  defaultProps={{
    title: "Default Title",
  }}
/>
```

### 5c. Add render script to package.json

```json
"render:myname": "remotion render src/index.ts MyVideo out/my-video.mp4"
```

## Step 6: Preview in Remotion Studio

```bash
cd marketing/videos && pnpm dev
```

Opens the Remotion Studio at `http://localhost:3000` (or next available port). Inspect frame-by-frame, tweak props in the GUI, verify timing.

## Step 7: Render

```bash
cd marketing/videos && pnpm render:myname
```

Output lands in `marketing/videos/out/my-video.mp4`.

Render all:
```bash
pnpm render:demo && pnpm render:walkthrough && pnpm render:social
```

## Common Patterns

### Full-bleed screenshot with browser frame + caption

```tsx
const ScreenshotScene: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  return (
    <AbsoluteFill style={{ backgroundColor: COLORS.background, alignItems: "center", justifyContent: "center", padding: 60, flexDirection: "column", gap: 24 }}>
      <div style={{ opacity: fadeIn(frame, 20) }}>
        <AnimatedText text="Section Title" style={{ fontSize: 36, fontWeight: 700, color: COLORS.foreground }} />
      </div>
      <div style={{ transform: `scale(${springEnter(frame, fps, { damping: 12 })})`, opacity: fadeIn(frame, 15), width: "100%" }}>
        <BrowserFrame url="localhost:3000/page" animate={false}>
          <Img src={staticFile("screenshots/my-screenshot.png")} style={{ width: "100%", display: "block" }} />
        </BrowserFrame>
      </div>
    </AbsoluteFill>
  );
};
```

### Workflow video scene

```tsx
const WorkflowScene: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  return (
    <AbsoluteFill style={{ backgroundColor: COLORS.background, alignItems: "center", justifyContent: "center", padding: 60 }}>
      <div style={{ transform: `scale(${springEnter(frame, fps, { damping: 14 })})`, opacity: fadeIn(frame, 20), width: "100%" }}>
        <BrowserFrame url="localhost:3000/environments" animate={false}>
          <OffthreadVideo src={staticFile("recordings/my-workflow.webm")} style={{ width: "100%", display: "block" }} />
        </BrowserFrame>
      </div>
    </AbsoluteFill>
  );
};
```

### Title/outro card

```tsx
const TitleScene: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const scale = springEnter(frame, fps, { damping: 10, stiffness: 80 });
  const opacity = fadeIn(frame, 15);
  return (
    <AbsoluteFill style={{ backgroundColor: COLORS.background, alignItems: "center", justifyContent: "center", flexDirection: "column", gap: 24 }}>
      <div style={{ width: 80, height: 80, backgroundColor: COLORS.primary, borderRadius: RADIUS.lg, display: "flex", alignItems: "center", justifyContent: "center", fontSize: 32, fontWeight: 700, color: COLORS.primaryForeground, fontFamily: FONTS.sans, transform: `scale(${scale})`, opacity }}>
        {BRAND.monogram}
      </div>
      <div style={{ opacity, transform: `scale(${scale})` }}>
        <AnimatedText text={BRAND.name} style={{ fontSize: 64, fontWeight: 700, color: COLORS.foreground }} />
      </div>
    </AbsoluteFill>
  );
};
```

## Checklist

- [ ] New capture tests added to `e2e/docs-capture/marketing.spec.ts` (inside `@marketing` describe)
- [ ] Captures run successfully (`pnpm docs:capture:marketing`)
- [ ] Assets copied to `marketing/videos/public/screenshots/` and/or `recordings/`
- [ ] Composition file created in `src/compositions/`
- [ ] Composition registered in `src/Root.tsx` with correct `durationInFrames`
- [ ] Render script added to `package.json`
- [ ] Previewed in Remotion Studio (`pnpm dev`) — timing and animations look correct
- [ ] Rendered to MP4 (`pnpm render:<name>`) — output in `marketing/videos/out/`
