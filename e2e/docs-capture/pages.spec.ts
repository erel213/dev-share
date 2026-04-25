import { test } from "@playwright/test";
import path from "path";
import { login, screenshotPage, DOCS_IMAGE_ROOT, ensureDir } from "./helpers";

test.describe("@docs pages", () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test("dashboard", async ({ page }) => {
    await screenshotPage(page, {
      path: "/",
      outputPath: "guides/dashboard.png",
    });
  });

  test("environments page", async ({ page }) => {
    await screenshotPage(page, {
      path: "/environments",
      outputPath: "guides/environments.png",
      waitFor: "Environments",
    });
  });

  test("templates list", async ({ page }) => {
    await screenshotPage(page, {
      path: "/templates",
      outputPath: "guides/template-list.png",
      waitFor: "Templates",
    });
  });

  test("template browser", async ({ page }) => {
    await page.goto("/templates");
    await page.waitForLoadState("networkidle");
    await page.locator("table tbody tr").first().getByRole("link").click();
    await page.waitForLoadState("networkidle");

    const outputFile = path.join(DOCS_IMAGE_ROOT, "guides/template-browser.png");
    ensureDir(path.dirname(outputFile));
    await page.screenshot({ path: outputFile, fullPage: true });
  });

  test("users page", async ({ page }) => {
    await screenshotPage(page, {
      path: "/users",
      outputPath: "guides/user-management.png",
      waitFor: "Users",
    });
  });

  test("groups page", async ({ page }) => {
    await screenshotPage(page, {
      path: "/groups",
      outputPath: "guides/group-management.png",
      waitFor: "Groups",
    });
  });
});

test.describe("@docs pages unauthenticated", () => {
  test("login page", async ({ page }) => {
    await screenshotPage(page, {
      path: "/login",
      outputPath: "guides/login.png",
    });
  });

  test("setup wizard", async ({ page }) => {
    await screenshotPage(page, {
      path: "/setup",
      outputPath: "guides/setup-wizard.png",
    });
  });
});
