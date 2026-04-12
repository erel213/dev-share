---
name: update-docs
description: >
  Update Hugo documentation when significant code changes are made.
  Triggers when: new API endpoints are added or modified, new domain entities
  are created, environment variables or configuration options change,
  new features are implemented, architectural decisions are made,
  breaking changes are introduced, or infrastructure templates are added.
allowed-tools: Read Write Edit Glob Grep Bash
---

# Documentation Update Skill

When significant code changes are made, follow this process to keep client-facing documentation in sync.

## Step 1: Assess the Change

Examine the current conversation context and any recent code modifications. Classify the change:

- **New API endpoint** — a new route or handler was added
- **Modified API endpoint** — request/response schema, behavior, or auth changed
- **New feature** — a new user-facing capability was implemented
- **Configuration change** — new env var, config option, or deployment parameter
- **Breaking change** — existing behavior was removed or altered incompatibly
- **New concept** — a new domain entity or architectural pattern clients interact with
- **Infrastructure change** — new templates, providers, or deployment workflows

## Step 2: Decide If Documentation Is Needed

**YES — update docs:**
- New or modified public API endpoint
- New or changed configuration option or environment variable
- New user-facing feature or workflow
- Breaking change to existing behavior
- New domain concept that clients interact with

**NO — skip docs:**
- Internal refactoring with no behavioral change
- Test-only changes
- CI/CD pipeline changes
- Code style or formatting fixes
- Internal performance optimizations invisible to clients

**UNCERTAIN — ask the user:**
- Dependency updates that may affect client behavior
- Internal architecture changes that could affect API contracts

If documentation is not needed, inform the user briefly and stop.

## Step 3: Create or Update Documentation

Reference the documentation rule at `.claude/rules/documentation.md` for style, structure, and conventions.

### For new API endpoints:
1. Read the handler to extract: route path, HTTP method, request/response types, auth requirements, error codes
2. Create or update the relevant page in `docs/content/docs/api/`
3. Include: method, path, request schema, response schema, curl example, error responses

### For new features:
1. Understand the end-to-end workflow from the user's perspective
2. Create a guide in `docs/content/docs/guides/` with prerequisites, step-by-step instructions, and expected outcomes

### For configuration changes:
1. Read `.env.example` or the relevant config to find new/changed options
2. Update `docs/content/docs/reference/` with the option name, type, default, and description

### For breaking changes:
1. Update all affected documentation pages
2. Add a clear migration note explaining what changed and how to adapt

### For new concepts:
1. Write an explainer in `docs/content/docs/concepts/` covering what it is, why it exists, and how clients interact with it

## Step 4: Capture UI Screenshots and Videos with Playwright

When documentation involves UI features (pages, dialogs, workflows), capture screenshots and videos using the doc capture tooling in `e2e/docs-capture/`.

**Prerequisite**: The app must be running (default: `http://localhost:3000`). The app should have data — either run the e2e suite first or set up manually.

### Running Captures

```bash
# All captures (screenshots + videos)
cd e2e && pnpm docs:capture

# Only page screenshots
cd e2e && pnpm docs:capture:pages

# Single scenario by name
cd e2e && npx playwright test --config=playwright.docs.config.ts --grep "dashboard"

# Headed mode (see the browser)
cd e2e && pnpm docs:capture:debug
```

Override the base URL or credentials with environment variables:

```bash
DOCS_BASE_URL=http://localhost:5173 DOCS_ADMIN_EMAIL=admin@example.com DOCS_ADMIN_PASSWORD=secret pnpm docs:capture
```

### File Structure

```
e2e/docs-capture/
├── helpers.ts          # login(), screenshotPage(), recordVideo() utilities
├── pages.spec.ts       # Page-level screenshots (dashboard, templates, users, etc.)
├── dialogs.spec.ts     # Dialog/modal state captures (create template, invite user, etc.)
└── workflows.spec.ts   # Video recordings of multi-step workflows
```

### Adding a New Page Screenshot

Add a `test()` block to `e2e/docs-capture/pages.spec.ts`:

```ts
test("my new page", async ({ page }) => {
  await screenshotPage(page, {
    path: "/my-page",
    outputPath: "guides/my-page.png",
    waitFor: "My Page Heading",
  });
});
```

### Adding a New Dialog Screenshot

Add a `test()` block to `e2e/docs-capture/dialogs.spec.ts`:

```ts
test("my dialog", async ({ page }) => {
  await page.goto("/my-page");
  await page.waitForLoadState("networkidle");
  await page.getByRole("button", { name: "Open Dialog" }).click();
  await page.locator("#dialog-field").waitFor({ state: "visible" });

  const outputFile = path.join(DOCS_IMAGE_ROOT, "guides/my-dialog.png");
  ensureDir(path.dirname(outputFile));
  await page.screenshot({ path: outputFile });
});
```

### Adding a New Video Workflow

Add a `test()` block to `e2e/docs-capture/workflows.spec.ts`:

```ts
test("my workflow", async ({ browser }) => {
  await recordVideo(browser, {
    outputPath: "guides/my-workflow.webm",
    workflow: async (page) => {
      await login(page);
      await page.goto("/my-page");
      await page.waitForLoadState("networkidle");
      // ... perform the workflow steps ...
      await page.waitForTimeout(1500); // pause at end for viewer
    },
  });
});
```

### Helper Functions (`e2e/docs-capture/helpers.ts`)

- **`login(page, email?, password?)`** — Authenticate via `/login` form. Uses `DOCS_ADMIN_EMAIL`/`DOCS_ADMIN_PASSWORD` env vars with e2e test defaults as fallback.
- **`screenshotPage(page, opts)`** — Navigate to a path, wait for networkidle + optional text, run optional `prepare` callback, save screenshot to `docs/static/images/{outputPath}`.
- **`recordVideo(browser, opts)`** — Create a context with video recording, run a workflow callback, save video to `docs/static/images/{outputPath}`.
- **`DOCS_IMAGE_ROOT`** — Resolved path to `docs/static/images/`.
- **`ensureDir(path)`** — Create directories recursively.

### Screenshot and Video Guidelines

- **Viewport**: 1280x720 (configured in `playwright.docs.config.ts`)
- **Save location**: `docs/static/images/` organized by section (`guides/`, `api/`, `concepts/`)
- **File naming**: kebab-case matching the doc page (e.g., `template-creation.png`, `user-management.png`)
- **Full page vs viewport**: Use viewport screenshots for specific UI elements, `fullPage: true` for complete page captures
- **Wait for load**: The `screenshotPage` helper handles `networkidle` automatically
- **Auth**: The `login` helper handles authentication. Place it in `beforeEach` for page/dialog specs or call it inside `recordVideo` workflow callbacks.
- **Light mode**: Capture in light mode by default for documentation readability
- **Video format**: Playwright records `.webm` by default. Convert to `.mp4` or `.gif` if needed for broader compatibility.

### Referencing in Docs

```markdown
![Template creation dialog](/images/guides/template-creation.png)
```

### When to Capture

- New UI page or major layout change → add to `pages.spec.ts`
- New dialog or modal workflow → add to `dialogs.spec.ts`
- Multi-step workflow → add to `workflows.spec.ts`
- Significant visual redesign → re-run `pnpm docs:capture` to refresh all

### When NOT to Capture

- The app is not running locally (inform the user to start it)
- API-only changes with no UI impact
- Minor styling tweaks

## Step 5: Verify

After writing or updating documentation:

1. Ensure frontmatter is valid YAML with `title`, `description`, and `weight`
2. Ensure internal links use Hugo ref shortcodes
3. Ensure code examples are syntactically correct
4. Ensure screenshots exist at the referenced paths
5. Ensure the page fits logically in the content hierarchy
6. Inform the user what documentation was created or updated
