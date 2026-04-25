import { test } from "@playwright/test";
import path from "path";
import { login, DOCS_IMAGE_ROOT, ensureDir } from "./helpers";

test.describe("@docs dialogs", () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test("create template dialog", async ({ page }) => {
    await page.goto("/templates");
    await page.waitForLoadState("networkidle");
    await page.getByRole("button", { name: "Create Template" }).click();
    await page.locator("#template-name").waitFor({ state: "visible" });

    const outputFile = path.join(
      DOCS_IMAGE_ROOT,
      "guides/create-template-dialog.png"
    );
    ensureDir(path.dirname(outputFile));
    await page.screenshot({ path: outputFile });
  });

  test("invite user dialog", async ({ page }) => {
    await page.goto("/users");
    await page.waitForLoadState("networkidle");
    await page.getByRole("button", { name: "Invite User" }).click();
    await page.locator("#invite-name").waitFor({ state: "visible" });

    const outputFile = path.join(
      DOCS_IMAGE_ROOT,
      "guides/invite-user-dialog.png"
    );
    ensureDir(path.dirname(outputFile));
    await page.screenshot({ path: outputFile });
  });

  test("create group dialog", async ({ page }) => {
    await page.goto("/groups");
    await page.waitForLoadState("networkidle");
    await page.getByRole("button", { name: "Create Group" }).click();
    await page.locator("#group-name").waitFor({ state: "visible" });

    const outputFile = path.join(
      DOCS_IMAGE_ROOT,
      "guides/create-group-dialog.png"
    );
    ensureDir(path.dirname(outputFile));
    await page.screenshot({ path: outputFile });
  });

  test("create environment dialog", async ({ page }) => {
    await page.goto("/environments");
    await page.waitForLoadState("networkidle");
    await page.getByRole("button", { name: "New Environment" }).click();
    await page
      .locator("[role='combobox'], select")
      .first()
      .waitFor({ state: "visible" });

    const outputFile = path.join(
      DOCS_IMAGE_ROOT,
      "guides/create-environment-dialog.png"
    );
    ensureDir(path.dirname(outputFile));
    await page.screenshot({ path: outputFile });
  });

  test("manage members dialog", async ({ page }) => {
    await page.goto("/groups");
    await page.waitForLoadState("networkidle");
    await page
      .locator("table tbody tr")
      .first()
      .getByRole("button", { name: /members/i })
      .click();
    await page.waitForLoadState("networkidle");

    const outputFile = path.join(
      DOCS_IMAGE_ROOT,
      "guides/manage-members-dialog.png"
    );
    ensureDir(path.dirname(outputFile));
    await page.screenshot({ path: outputFile });
  });
});
