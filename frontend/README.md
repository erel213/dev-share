# Dev-Share Frontend

React-based web interface for dev-share — managing developer environments, templates, groups, and users.

## Tech Stack

- **Framework**: React 19 + TypeScript
- **Build Tool**: Vite 7
- **Styling**: Tailwind CSS v4 + shadcn/ui (New York style, Zinc base)
- **State Management**: Redux Toolkit + React Redux
- **Routing**: React Router v7
- **HTTP Client**: Axios
- **Icons**: Lucide React
- **Package Manager**: pnpm

## Getting Started

### Prerequisites
- Node.js 22+
- pnpm

### Development

```bash
pnpm install
pnpm dev
```

The dev server runs at `http://localhost:5173` and proxies API requests to the backend at the URL specified by `VITE_API_BASE_URL` (default: `http://localhost:8080`).

### Scripts

| Script | Description |
|---|---|
| `pnpm dev` | Start Vite dev server with HMR |
| `pnpm build` | Type-check and build for production |
| `pnpm lint` | Run ESLint |
| `pnpm preview` | Preview production build locally |

## Project Structure

```
src/
├── pages/              # Page-level components
│   ├── HomePage.tsx
│   ├── LoginPage.tsx
│   ├── SetupPage.tsx
│   ├── TemplateBrowserPage.tsx
│   ├── TemplatesPage.tsx
│   ├── EnvironmentsPage.tsx
│   ├── GroupsPage.tsx
│   └── UsersPage.tsx
├── components/         # Reusable UI components (shadcn/ui + custom)
├── store/              # Redux store (authSlice)
├── router/             # React Router route definitions
├── lib/                # Utilities (Axios instance, cn() helper)
├── hooks/              # Custom React hooks
└── types/              # TypeScript type definitions
```

## Configuration

- **Path alias**: `@/*` maps to `src/*` (configured in `vite.config.ts` and `tsconfig.app.json`)
- **API base URL**: Set `VITE_API_BASE_URL` in `.env` (see root `.env.example`)
- **Authentication**: JWT via httpOnly cookies — Axios is configured with `withCredentials: true`

## Docker

The frontend is containerized with Nginx for production. The `Dockerfile` builds the Vite app and serves it behind Nginx, which proxies `/api` and `/admin` requests to the backend container. See `nginx.conf` for proxy configuration.
