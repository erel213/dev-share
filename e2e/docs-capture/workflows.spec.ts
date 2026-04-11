import { test } from "@playwright/test";
import { login, recordVideo } from "./helpers";

test.describe("@docs workflows", () => {
  test("template creation workflow", async ({ browser }) => {
    await recordVideo(browser, {
      outputPath: "guides/create-template-workflow.webm",
      workflow: async (page) => {
        await login(page);
        await page.goto("/templates");
        await page.waitForLoadState("networkidle");

        // Open the create template dialog
        await page.getByRole("button", { name: "Create Template" }).click();
        await page.locator("#template-name").waitFor({ state: "visible" });

        // Fill in template name
        await page.locator("#template-name").fill("demo-template");

        // Pause so viewer can see the filled form
        await page.waitForTimeout(1500);
      },
    });
  });

  test("user invite workflow", async ({ browser }) => {
    await recordVideo(browser, {
      outputPath: "guides/invite-user-workflow.webm",
      workflow: async (page) => {
        await login(page);
        await page.goto("/users");
        await page.waitForLoadState("networkidle");

        // Open the invite user dialog
        await page.getByRole("button", { name: "Invite User" }).click();
        await page.locator("#invite-name").waitFor({ state: "visible" });

        // Fill in user details
        await page.locator("#invite-name").fill("Jane Developer");
        await page.locator("#invite-email").fill("jane@example.com");

        // Select role
        await page.getByRole("combobox", { name: "Role" }).click();
        await page.getByRole("option", { name: "Editor" }).click();

        // Pause so viewer can see the completed form
        await page.waitForTimeout(1500);
      },
    });
  });
});
