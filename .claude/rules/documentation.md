---
paths:
  - "docs/**/*"
  - "backend/internal/infra/http/handlers/**/*.go"
  - "backend/internal/domain/**/*.go"
  - "backend/internal/application/**/*.go"
  - "backend/pkg/**/*.go"
  - "frontend/src/**/*.ts"
  - "frontend/src/**/*.tsx"
  - ".env.example"
  - "docker-compose.yml"
---

# Documentation Rules

This project maintains client-facing documentation as a Hugo site using the Doks theme, located in `docs/` at the project root.

## Documentation Location and Structure

All documentation lives under `docs/content/docs/` with these sections:

```
docs/content/docs/
├── _index.md                  # Docs landing page
├── getting-started/           # Installation, quickstart, initial configuration
│   └── _index.md
├── guides/                    # Task-oriented how-to guides
│   └── _index.md
├── api/                       # API reference (endpoints, request/response schemas)
│   └── _index.md
├── concepts/                  # Architecture, domain concepts, design decisions
│   └── _index.md
└── reference/                 # Environment variables, config options, CLI flags
    └── _index.md
```

- Static assets (images, diagrams) go in `docs/static/`
- Use leaf bundles (`page-name/index.md`) when a page has associated assets

## Hugo / Doks Conventions

### Frontmatter

Every content file must have YAML frontmatter:

```yaml
---
title: "Page Title"
description: "Brief description for SEO and navigation."
weight: 10
draft: false
---
```

- `weight` controls ordering within a section (lower = first)
- Set `draft: true` only for work-in-progress pages
- Section index files (`_index.md`) define the section title and description

### Links

- Use Hugo ref shortcodes for internal links: `[link text]({{< ref "docs/guides/some-page" >}})`
- Use relative paths within the same section when possible

### Shortcodes

- Use Doks built-in shortcodes (alerts, tabs, details) instead of raw HTML
- Use fenced code blocks with language identifiers for all code examples

## Writing Style

- **Audience**: Client developers and platform administrators who use dev-share
- **Voice**: Second person ("you"), present tense, active voice
- **Structure**: One concept or task per page — keep pages focused
- **Code examples**: Include runnable examples for every API endpoint and configuration option
- **Prerequisites**: Start how-to guides with a "Prerequisites" section
- **Headings**: Use sentence case for all headings

### API Reference Pages Must Include

1. HTTP method and path
2. Request body schema (with types and required/optional markers)
3. Response schema (success and error)
4. Example `curl` command
5. Error codes and their meanings

## When to Update Documentation

| Change Type | Action |
|-------------|--------|
| New API endpoint | Create/update page in `docs/content/docs/api/` |
| Modified API endpoint | Update existing API page with new schema/behavior |
| New user-facing feature | Add a guide in `docs/content/docs/guides/` |
| New environment variable or config option | Update `docs/content/docs/reference/` |
| Breaking change | Update affected pages AND add a migration note |
| New domain concept clients interact with | Add explainer in `docs/content/docs/concepts/` |

## DO

- Keep API reference pages in sync with handler route definitions
- Update `lastmod` in frontmatter when modifying existing pages
- Include `curl` examples that work against a local dev server (`localhost:8080`)
- Use admonition shortcodes for warnings, tips, and important notes
- Test that code examples are syntactically correct

## DON'T

- Document internal implementation details (domain layer internals, repository patterns) in client-facing docs
- Include real credentials, secrets, or tokens in examples — use placeholders like `YOUR_API_KEY`
- Use screenshots for CLI output — use code blocks instead
- Add pages without proper frontmatter
- Create deeply nested sections (max 2 levels deep under `docs/`)
