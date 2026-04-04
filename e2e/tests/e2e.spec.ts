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

const INVITED_USER = {
  name: "E2E Invited User",
  email: "invited@e2e-test.com",
  role: "Editor",
};

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

  // ── 4. User Management ────────────────────────────────────────────

  let invitedUserPassword = "";

  test("user management — admin can see Users link in sidebar", async ({
    page,
  }) => {
    // Login as admin
    await page.goto("/login");
    await page.locator("#email").fill(ADMIN.email);
    await page.locator("#password").fill(ADMIN.password);
    await page.getByRole("button", { name: "Log in" }).click();
    await expect(page).toHaveURL("/", { timeout: 10000 });

    // Verify Users link in sidebar
    await expect(page.getByRole("link", { name: "Users" })).toBeVisible();
  });

  test("user management — admin can navigate to users page", async ({
    page,
  }) => {
    await page.goto("/login");
    await page.locator("#email").fill(ADMIN.email);
    await page.locator("#password").fill(ADMIN.password);
    await page.getByRole("button", { name: "Log in" }).click();
    await expect(page).toHaveURL("/", { timeout: 10000 });

    await page.getByRole("link", { name: "Users" }).click();
    await expect(page).toHaveURL("/users");
    await expect(
      page.getByRole("heading", { name: "Users" })
    ).toBeVisible();
  });

  test("user management — admin can invite a user", async ({ page }) => {
    await page.goto("/login");
    await page.locator("#email").fill(ADMIN.email);
    await page.locator("#password").fill(ADMIN.password);
    await page.getByRole("button", { name: "Log in" }).click();
    await expect(page).toHaveURL("/", { timeout: 10000 });

    await page.goto("/users");
    await expect(
      page.getByRole("heading", { name: "Users" })
    ).toBeVisible();

    // Click Invite User button
    await page.getByRole("button", { name: "Invite User" }).click();

    // Fill in the invite form
    await page.locator("#invite-name").fill(INVITED_USER.name);
    await page.locator("#invite-email").fill(INVITED_USER.email);

    // Select role
    await page.getByRole("combobox", { name: "Role" }).click();
    await page.getByRole("option", { name: INVITED_USER.role }).click();

    // Submit
    await page.getByRole("button", { name: "Invite" }).click();

    // Verify password is shown
    await expect(page.getByText("User Invited")).toBeVisible({
      timeout: 10000,
    });
    await expect(
      page.getByText("This password will not be shown again")
    ).toBeVisible();

    // Capture the generated password
    const passwordInput = page.locator("input[readonly]");
    invitedUserPassword = (await passwordInput.inputValue()) || "";
    expect(invitedUserPassword.length).toBeGreaterThan(0);

    // Close the dialog
    await page.getByRole("button", { name: "Done" }).click();
  });

  test("user management — invited user appears in table", async ({
    page,
  }) => {
    await page.goto("/login");
    await page.locator("#email").fill(ADMIN.email);
    await page.locator("#password").fill(ADMIN.password);
    await page.getByRole("button", { name: "Log in" }).click();
    await expect(page).toHaveURL("/", { timeout: 10000 });

    await page.goto("/users");

    // Verify invited user is in the table
    await expect(page.getByText(INVITED_USER.name)).toBeVisible({
      timeout: 10000,
    });
    await expect(page.getByText(INVITED_USER.email)).toBeVisible();
  });

  test("user management — admin can reset user password", async ({
    page,
  }) => {
    await page.goto("/login");
    await page.locator("#email").fill(ADMIN.email);
    await page.locator("#password").fill(ADMIN.password);
    await page.getByRole("button", { name: "Log in" }).click();
    await expect(page).toHaveURL("/", { timeout: 10000 });

    await page.goto("/users");
    await expect(page.getByText(INVITED_USER.name)).toBeVisible({
      timeout: 10000,
    });

    // Find the row for the invited user and click reset password
    const userRow = page
      .getByRole("row")
      .filter({ hasText: INVITED_USER.email });
    await userRow.getByTitle("Reset password").click();

    // Confirm reset
    await page.getByRole("button", { name: "Reset Password" }).click();

    // Verify new password is shown
    await expect(page.getByText("Password Reset")).toBeVisible({
      timeout: 10000,
    });
    await expect(
      page.getByText("This password will not be shown again")
    ).toBeVisible();

    // Capture new password for next test
    const passwordInput = page.locator("input[readonly]");
    invitedUserPassword = (await passwordInput.inputValue()) || "";
    expect(invitedUserPassword.length).toBeGreaterThan(0);

    await page.getByRole("button", { name: "Done" }).click();
  });

  test("user management — invited user can login with reset password", async ({
    page,
  }) => {
    // Login as the invited user with the reset password
    await page.goto("/login");
    await page.locator("#email").fill(INVITED_USER.email);
    await page.locator("#password").fill(invitedUserPassword);
    await page.getByRole("button", { name: "Log in" }).click();

    // Should reach dashboard
    await expect(page).toHaveURL("/", { timeout: 10000 });
  });

  // ── 5. Group Management ─────────────────────────────────────────────

  const GROUP = {
    name: "Engineering",
    description: "Backend and frontend engineers",
  };

  test("group management — admin can navigate to groups page", async ({
    page,
  }) => {
    await page.goto("/login");
    await page.locator("#email").fill(ADMIN.email);
    await page.locator("#password").fill(ADMIN.password);
    await page.getByRole("button", { name: "Log in" }).click();
    await expect(page).toHaveURL("/", { timeout: 10000 });

    // Verify Groups link in sidebar
    await expect(page.getByRole("link", { name: "Groups" })).toBeVisible();

    await page.getByRole("link", { name: "Groups" }).click();
    await expect(page).toHaveURL("/groups");
    await expect(
      page.getByRole("heading", { name: "Groups" })
    ).toBeVisible();
  });

  test("group management — admin can create a group", async ({ page }) => {
    await page.goto("/login");
    await page.locator("#email").fill(ADMIN.email);
    await page.locator("#password").fill(ADMIN.password);
    await page.getByRole("button", { name: "Log in" }).click();
    await expect(page).toHaveURL("/", { timeout: 10000 });

    await page.goto("/groups");
    await expect(
      page.getByRole("heading", { name: "Groups" })
    ).toBeVisible();

    // Click Create Group button
    await page.getByRole("button", { name: "Create Group" }).click();

    // Fill in the form
    await page.locator("#group-name").fill(GROUP.name);
    await page.locator("#group-description").fill(GROUP.description);

    // Submit
    await page.getByRole("button", { name: "Create" }).click();

    // Verify group appears in the table
    await expect(page.getByText(GROUP.name)).toBeVisible({ timeout: 10000 });
    await expect(page.getByText(GROUP.description)).toBeVisible();
    await expect(page.getByText("Custom")).toBeVisible();
  });

  test("group management — created group appears in table after reload", async ({
    page,
  }) => {
    await page.goto("/login");
    await page.locator("#email").fill(ADMIN.email);
    await page.locator("#password").fill(ADMIN.password);
    await page.getByRole("button", { name: "Log in" }).click();
    await expect(page).toHaveURL("/", { timeout: 10000 });

    await page.goto("/groups");

    // Verify group persisted
    await expect(page.getByText(GROUP.name)).toBeVisible({ timeout: 10000 });
    await expect(page.getByText(GROUP.description)).toBeVisible();
  });

  test("group management — admin can edit a group", async ({ page }) => {
    await page.goto("/login");
    await page.locator("#email").fill(ADMIN.email);
    await page.locator("#password").fill(ADMIN.password);
    await page.getByRole("button", { name: "Log in" }).click();
    await expect(page).toHaveURL("/", { timeout: 10000 });

    await page.goto("/groups");
    await expect(page.getByText(GROUP.name)).toBeVisible({ timeout: 10000 });

    // Click edit on the group row
    const groupRow = page.getByRole("row").filter({ hasText: GROUP.name });
    await groupRow.getByTitle("Edit group").click();

    // Toggle access all templates
    await page.locator("#edit-access-all").click();

    // Save
    await page.getByRole("button", { name: "Save" }).click();

    // Verify badge changed to "All Templates"
    await expect(page.getByText("All Templates")).toBeVisible({
      timeout: 10000,
    });
  });

  test("group management — admin can delete a group", async ({ page }) => {
    await page.goto("/login");
    await page.locator("#email").fill(ADMIN.email);
    await page.locator("#password").fill(ADMIN.password);
    await page.getByRole("button", { name: "Log in" }).click();
    await expect(page).toHaveURL("/", { timeout: 10000 });

    await page.goto("/groups");
    await expect(page.getByText(GROUP.name)).toBeVisible({ timeout: 10000 });

    // Click delete on the group row
    const groupRow = page.getByRole("row").filter({ hasText: GROUP.name });
    await groupRow.getByTitle("Delete group").click();

    // Confirm deletion
    await page.getByRole("button", { name: "Delete" }).click();

    // Wait for dialog to close
    await expect(page.getByRole("alertdialog")).not.toBeVisible({
      timeout: 10000,
    });

    // Verify group is removed
    await expect(
      page.getByRole("row").filter({ hasText: GROUP.name })
    ).not.toBeVisible({ timeout: 10000 });
  });

  // ── 6. User Cleanup ────────────────────────────────────────────────

  test("user management — admin can delete a user", async ({ page }) => {
    // Login as admin
    await page.goto("/login");
    await page.locator("#email").fill(ADMIN.email);
    await page.locator("#password").fill(ADMIN.password);
    await page.getByRole("button", { name: "Log in" }).click();
    await expect(page).toHaveURL("/", { timeout: 10000 });

    await page.goto("/users");
    await expect(page.getByText(INVITED_USER.name)).toBeVisible({
      timeout: 10000,
    });

    // Find the row and click delete
    const userRow = page
      .getByRole("row")
      .filter({ hasText: INVITED_USER.email });
    await userRow.getByTitle("Delete user").click();

    // Confirm deletion
    await page.getByRole("button", { name: "Delete" }).click();

    // Wait for the confirmation dialog to close
    await expect(page.getByRole("alertdialog")).not.toBeVisible({
      timeout: 10000,
    });

    // Verify user is removed from table
    await expect(
      page.getByRole("row").filter({ hasText: INVITED_USER.email })
    ).not.toBeVisible({ timeout: 10000 });
  });
});
