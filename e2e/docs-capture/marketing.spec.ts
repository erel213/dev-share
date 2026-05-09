import { test } from "@playwright/test";
import path from "path";
import { login, screenshotPage, recordVideo, DOCS_IMAGE_ROOT, ensureDir } from "./helpers";

test.describe("@marketing", () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test("hero — dashboard with data", async ({ page }) => {
    await screenshotPage(page, {
      path: "/",
      outputPath: "marketing/hero-dashboard.png",
      fullPage: false,
    });
  });

  test("environments table — hover state", async ({ page }) => {
    await screenshotPage(page, {
      path: "/environments",
      outputPath: "marketing/environments-hover.png",
      waitFor: "Environments",
      prepare: async (page) => {
        await page.locator("table tbody tr").first().hover();
        await page.waitForTimeout(200);
      },
    });
  });

  // Uses raw page.screenshot() — screenshotPage() navigates away and would close the dialog
  test("environments — create environment dialog open", async ({ page }) => {
    await page.goto("/environments");
    await page.waitForLoadState("networkidle");
    await page.getByRole("button", { name: "New Environment" }).click();
    await page
      .locator("[role='combobox'], select")
      .first()
      .waitFor({ state: "visible" });

    const outputFile = path.join(
      DOCS_IMAGE_ROOT,
      "marketing/create-environment-dialog.png"
    );
    ensureDir(path.dirname(outputFile));
    await page.screenshot({ path: outputFile, fullPage: false });
  });

  test("users — user management table", async ({ page }) => {
    await screenshotPage(page, {
      path: "/users",
      outputPath: "marketing/users-management.png",
      waitFor: "Users",
    });
  });

  // Uses raw page.screenshot() — screenshotPage() navigates away and would close the dialog
  test("groups — manage templates dialog", async ({ page }) => {
    await page.goto("/groups");
    await page.waitForLoadState("networkidle");
    await page.locator("table tbody tr").first().hover();
    await page.waitForTimeout(200);
    await page
      .locator("table tbody tr")
      .first()
      .getByRole("button", { name: /manage templates/i })
      .click();
    await page
      .locator("[role='dialog']")
      .waitFor({ state: "visible" });

    const outputFile = path.join(
      DOCS_IMAGE_ROOT,
      "marketing/groups-templates-dialog.png"
    );
    ensureDir(path.dirname(outputFile));
    await page.screenshot({ path: outputFile, fullPage: false });
  });

  test("workflow — admin provision user and assign template access", async ({ browser }) => {
    await recordVideo(browser, {
      outputPath: "marketing/admin-provision-workflow.webm",
      workflow: async (page) => {
        await login(page);
        await page.goto("/users");
        await page.waitForLoadState("networkidle");
        await page.waitForTimeout(800);
        await page.goto("/groups");
        await page.waitForLoadState("networkidle");
        await page.waitForTimeout(800);
        await page.locator("table tbody tr").first().hover();
        await page.waitForTimeout(300);
        await page
          .locator("table tbody tr")
          .first()
          .getByRole("button", { name: /manage templates/i })
          .click();
        await page.locator("[role='dialog']").waitFor({ state: "visible" });
        await page.waitForTimeout(1500);
      },
    });
  });

  test("workflow — create environment", async ({ browser }) => {
    await recordVideo(browser, {
      outputPath: "marketing/create-environment-workflow.webm",
      workflow: async (page) => {
        await login(page);
        await page.goto("/environments");
        await page.waitForLoadState("networkidle");
        await page.getByRole("button", { name: "New Environment" }).click();
        await page
          .locator("[role='combobox'], select")
          .first()
          .waitFor({ state: "visible" });
        await page.waitForTimeout(1500);
      },
    });
  });
});
