import { type Page, type Browser } from "@playwright/test";
import path from "path";
import fs from "fs";

/** Root directory for documentation images */
export const DOCS_IMAGE_ROOT = path.resolve(
  __dirname,
  "../../docs/static/images"
);

/** Default credentials (match e2e test data) */
const DEFAULT_EMAIL = process.env.DOCS_ADMIN_EMAIL || "admin@e2e-test.com";
const DEFAULT_PASSWORD =
  process.env.DOCS_ADMIN_PASSWORD || "TestPassword123!";

/** Ensure a directory exists, creating intermediate dirs as needed */
export function ensureDir(dirPath: string): void {
  fs.mkdirSync(dirPath, { recursive: true });
}

/**
 * Log in to the app using the standard auth flow.
 * Navigates to /login, fills credentials, and waits for redirect to /.
 */
export async function login(
  page: Page,
  email: string = DEFAULT_EMAIL,
  password: string = DEFAULT_PASSWORD
): Promise<void> {
  await page.goto("/login");
  await page.locator("#email").fill(email);
  await page.locator("#password").fill(password);
  await page.getByRole("button", { name: "Log in" }).click();
  await page.waitForURL("/", { timeout: 10000 });
}

export interface ScreenshotPageOptions {
  /** URL path to navigate to (e.g., "/templates") */
  path: string;
  /** Output path relative to docs/static/images/ (e.g., "guides/dashboard.png") */
  outputPath: string;
  /** Capture the full scrollable page instead of just the viewport */
  fullPage?: boolean;
  /** Text content to wait for before capturing (waits for it to be visible) */
  waitFor?: string;
  /** Optional callback to prepare UI state before capturing */
  prepare?: (page: Page) => Promise<void>;
}

/**
 * Navigate to a page, wait for it to load, and take a screenshot.
 * Saves to docs/static/images/{outputPath}.
 */
export async function screenshotPage(
  page: Page,
  options: ScreenshotPageOptions
): Promise<string> {
  await page.goto(options.path);
  await page.waitForLoadState("networkidle");

  if (options.waitFor) {
    await page.getByText(options.waitFor).first().waitFor({ state: "visible" });
  }

  if (options.prepare) {
    await options.prepare(page);
  }

  const outputFile = path.join(DOCS_IMAGE_ROOT, options.outputPath);
  ensureDir(path.dirname(outputFile));

  await page.screenshot({
    path: outputFile,
    fullPage: options.fullPage ?? false,
  });

  return outputFile;
}

export interface RecordVideoOptions {
  /** Output path relative to docs/static/images/ (e.g., "guides/create-template.webm") */
  outputPath: string;
  /** Async workflow to execute while recording */
  workflow: (page: Page) => Promise<void>;
}

/**
 * Record a video of a workflow.
 * Creates a new browser context with video recording enabled,
 * runs the workflow, then moves the video to the target path.
 */
export async function recordVideo(
  browser: Browser,
  options: RecordVideoOptions
): Promise<string> {
  const tempDir = path.join(__dirname, "../test-results/docs-videos-tmp");
  ensureDir(tempDir);

  const context = await browser.newContext({
    viewport: { width: 1280, height: 720 },
    recordVideo: {
      dir: tempDir,
      size: { width: 1280, height: 720 },
    },
  });

  const page = await context.newPage();

  try {
    await options.workflow(page);
  } finally {
    await page.close();
    const video = page.video();
    if (video) {
      const outputFile = path.join(DOCS_IMAGE_ROOT, options.outputPath);
      ensureDir(path.dirname(outputFile));
      await video.saveAs(outputFile);
    }
    await context.close();
  }

  const outputFile = path.join(DOCS_IMAGE_ROOT, options.outputPath);
  return outputFile;
}
