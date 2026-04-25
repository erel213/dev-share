import { request } from "@playwright/test";

/**
 * Global setup for docs-capture tests.
 *
 * Ensures the app is in a known state before any screenshot test runs:
 *   1. System is initialized (admin account + workspace exist)
 *   2. At least one template exists (needed for template-browser screenshot)
 *   3. At least one group exists (needed for manage-members-dialog screenshot)
 *
 * Uses DOCS_BASE_URL (default: http://localhost:3000) — the same URL as the
 * browser tests, so Nginx/Vite proxy handles routing to the backend.
 */

const BASE_URL = process.env.DOCS_BASE_URL || "http://localhost:3000";
const ADMIN_EMAIL = process.env.DOCS_ADMIN_EMAIL || "admin@e2e-test.com";
const ADMIN_PASSWORD = process.env.DOCS_ADMIN_PASSWORD || "TestPassword123!";

const SAMPLE_TF = `
variable "environment_name" {
  description = "Name for this environment"
  type        = string
}

output "environment_name" {
  value = var.environment_name
}
`.trim();

export default async function globalSetup() {
  const api = await request.newContext({ baseURL: BASE_URL });

  try {
    // ── 1. Initialize system if needed ──────────────────────────────────
    const statusRes = await api.get("/admin/status");
    const status = await statusRes.json();

    if (!status.initialized) {
      console.log("[docs-setup] System not initialized — running setup...");

      const initRes = await api.post("/admin/init", {
        data: {
          admin_name: "Docs Admin",
          admin_email: ADMIN_EMAIL,
          admin_password: ADMIN_PASSWORD,
          workspace_name: "Docs Workspace",
          workspace_description: "Workspace created by docs-capture setup",
        },
      });

      if (!initRes.ok()) {
        throw new Error(
          `[docs-setup] /admin/init failed (${initRes.status()}): ${await initRes.text()}`
        );
      }

      console.log("[docs-setup] System initialized successfully.");
    } else {
      console.log("[docs-setup] System already initialized, skipping init.");
    }

    // ── 2. Log in to get auth cookie + workspace ID ──────────────────────
    const loginRes = await api.post("/api/v1/login", {
      data: { email: ADMIN_EMAIL, password: ADMIN_PASSWORD },
    });

    if (!loginRes.ok()) {
      throw new Error(
        `[docs-setup] Login failed (${loginRes.status()}): ${await loginRes.text()}`
      );
    }

    const { workspace_id: workspaceId } = await loginRes.json();

    // ── 3. Seed a template if none exist ────────────────────────────────
    const templatesRes = await api.get("/api/v1/templates");
    const templates = await templatesRes.json();

    if (!Array.isArray(templates) || templates.length === 0) {
      console.log("[docs-setup] No templates found — creating sample template...");

      const createTemplateRes = await api.post("/api/v1/templates", {
        multipart: {
          name: "Sample Infrastructure",
          workspace_id: workspaceId,
          paths: "main.tf",
          files: {
            name: "main.tf",
            mimeType: "text/plain",
            buffer: Buffer.from(SAMPLE_TF),
          },
        },
      });

      if (!createTemplateRes.ok()) {
        throw new Error(
          `[docs-setup] Template creation failed (${createTemplateRes.status()}): ${await createTemplateRes.text()}`
        );
      }

      console.log("[docs-setup] Sample template created.");
    } else {
      console.log(`[docs-setup] ${templates.length} template(s) already exist, skipping.`);
    }

    // ── 4. Seed a group if none exist ────────────────────────────────────
    const groupsRes = await api.get("/api/v1/groups");
    const groups = await groupsRes.json();

    if (!Array.isArray(groups) || groups.length === 0) {
      console.log("[docs-setup] No groups found — creating sample group...");

      const createGroupRes = await api.post("/api/v1/groups", {
        data: {
          name: "Frontend Team",
          description: "Developers working on the frontend stack",
          access_all_templates: true,
        },
      });

      if (!createGroupRes.ok()) {
        throw new Error(
          `[docs-setup] Group creation failed (${createGroupRes.status()}): ${await createGroupRes.text()}`
        );
      }

      console.log("[docs-setup] Sample group created.");
    } else {
      console.log(`[docs-setup] ${groups.length} group(s) already exist, skipping.`);
    }
  } finally {
    await api.dispose();
  }
}
