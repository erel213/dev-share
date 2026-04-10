# E2E Tests

End-to-end tests for the dev-share application using [Playwright](https://playwright.dev/).

## Prerequisites

- Node.js 22+
- pnpm
- Playwright browsers: `pnpm exec playwright install`

## Running Tests

The app must be running (frontend on `http://localhost:3000`, backend on `http://localhost:8080`) before executing tests.

```bash
# Install dependencies
pnpm install

# Run tests (headless)
pnpm test

# Run with browser visible
pnpm test:headed

# Run in debug mode (step-through)
pnpm test:debug

# View HTML test report
pnpm report
```

## Configuration

See `playwright.config.ts`:

- **Base URL**: `http://localhost:3000`
- **Browser**: Chromium only
- **Workers**: 1 (sequential execution)
- **Traces**: Retained on failure
- **Screenshots**: Captured on failure

## Project Structure

```
e2e/
├── tests/
│   └── e2e.spec.ts          # Main test suite
├── infra/                    # Terraform configs for provisioning test environments
├── sample-template/          # Sample Terraform template used in tests
├── playwright.config.ts      # Playwright configuration
├── package.json              # Dependencies and scripts
└── deploy.sh                 # Deployment helper script
```

- **`infra/`** — Terraform configuration for spinning up cloud infrastructure to run the app (EC2 instance provisioning)
- **`sample-template/`** — A minimal Terraform template used by tests to exercise the template upload and environment provisioning flows
