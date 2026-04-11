import { test } from "@playwright/test";
import { login, screenshotPage } from "./helpers";

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

  test("templates list", async ({ page }) => {
    await screenshotPage(page, {
      path: "/templates",
      outputPath: "guides/template-list.png",
      waitFor: "Templates",
    });
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
