import { test, expect } from "@playwright/test";
import path from "path";

// Test data
const ADMIN = {
  name: "E2E Admin",
  email: "admin@e2e-test.com",
  password: "TestPassword123!",
};

const WORKSPACE = {
  name: "E2E Workspace",
  description: "Automated E2E testing workspace",
};

const TEMPLATE_NAME = "e2e-sample-template";

// Path to sample Terraform files
const SAMPLE_TEMPLATE_DIR = path.resolve(__dirname, "../sample-template");

test.describe.serial("Dev-Share E2E", () => {
  // ── 1. Setup Wizard ─────────────────────────────────────────────────

  test("setup wizard — complete initial setup", async ({ page }) => {
    await page.goto("/");

    // Should redirect to /setup on first visit
    await expect(page).toHaveURL(/\/setup/);

    // Step 1: Welcome — click "Get Started"
    await page.getByRole("button", { name: "Get Started" }).click();

    // Step 2: Admin Account
    await page.locator("#name").fill(ADMIN.name);
    await page.locator("#email").fill(ADMIN.email);
    await page.locator("#password").fill(ADMIN.password);
    await page.locator("#confirm-password").fill(ADMIN.password);
    await page.getByRole("button", { name: "Next" }).click();

    // Step 3: Workspace
    await page.locator("#workspace-name").fill(WORKSPACE.name);
    await page.locator("#workspace-desc").fill(WORKSPACE.description);
    await page.getByRole("button", { name: "Next" }).click();

    // Step 4: Review — verify summary then submit
    await expect(page.getByText(ADMIN.name)).toBeVisible();
    await expect(page.getByText(ADMIN.email)).toBeVisible();
    await expect(page.getByText(WORKSPACE.name)).toBeVisible();
    await page.getByRole("button", { name: "Initialize" }).click();

    // Step 5: Success — verify completion and navigate to dashboard
    await expect(
      page.getByRole("button", { name: "Go to Dashboard" })
    ).toBeVisible({ timeout: 15000 });
    await page.getByRole("button", { name: "Go to Dashboard" }).click();

    await expect(page).toHaveURL("/");
  });

  // ── 2. Login ────────────────────────────────────────────────────────

  test("login — authenticate as admin", async ({ page }) => {
    await page.goto("/login");

    await page.locator("#email").fill(ADMIN.email);
    await page.locator("#password").fill(ADMIN.password);
    await page.getByRole("button", { name: "Log in" }).click();

    // Should redirect to home after successful login
    await expect(page).toHaveURL("/", { timeout: 10000 });
  });

  // ── 3. Create Template ─────────────────────────────────────────────

  test("create template — upload terraform files and verify", async ({
    page,
  }) => {
    // Login first (no shared auth state between serial tests)
    await page.goto("/login");
    await page.locator("#email").fill(ADMIN.email);
    await page.locator("#password").fill(ADMIN.password);
    await page.getByRole("button", { name: "Log in" }).click();
    await expect(page).toHaveURL("/", { timeout: 10000 });

    // Navigate to templates page
    await page.goto("/templates");
    await expect(
      page.getByRole("heading", { name: "Templates" })
    ).toBeVisible();

    // Open create template dialog
    await page.getByRole("button", { name: "Create Template" }).click();

    // Fill in template name
    await page.locator("#template-name").fill(TEMPLATE_NAME);

    // Upload sample terraform files
    const fileChooserPromise = page.waitForEvent("filechooser");
    await page.getByRole("button", { name: "Browse Folder" }).click();
    const fileChooser = await fileChooserPromise;
    await fileChooser.setFiles(SAMPLE_TEMPLATE_DIR);

    // Verify files appear as badges
    await expect(page.getByText("main.tf")).toBeVisible();
    await expect(page.getByText("variables.tf")).toBeVisible();
    await expect(page.getByText("outputs.tf")).toBeVisible();

    // Submit
    await page.getByRole("button", { name: "Create" }).click();

    // Verify template appears in the list
    await expect(page.getByRole("link", { name: TEMPLATE_NAME })).toBeVisible({
      timeout: 10000,
    });

    // Navigate to template detail page
    await page.getByRole("link", { name: TEMPLATE_NAME }).click();
    await expect(page).toHaveURL(/\/templates\/.+/);

    // Verify files are visible in the file tree
    await expect(page.getByText("main.tf")).toBeVisible();
    await expect(page.getByText("variables.tf")).toBeVisible();
    await expect(page.getByText("outputs.tf")).toBeVisible();

    // Switch to Variables tab and parse
    await page.getByRole("button", { name: "Variables", exact: true }).click();
    await page.getByRole("button", { name: "Parse & Reconcile" }).click();

    // Verify variables were detected from variables.tf
    await expect(page.getByText("region")).toBeVisible({ timeout: 10000 });
    await expect(page.getByText("instance_type")).toBeVisible();
    await expect(page.getByText("ami_id")).toBeVisible();
  });
});
