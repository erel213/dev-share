// Color tokens mirroring frontend/src/index.css dark-mode OKLCH values.
// Used as inline styles — required for deterministic per-frame Remotion rendering.

export const COLORS = {
  background: "oklch(0.145 0 0)",
  foreground: "oklch(0.985 0 0)",
  card: "oklch(0.205 0 0)",
  cardForeground: "oklch(0.985 0 0)",
  primary: "oklch(0.922 0 0)",
  primaryForeground: "oklch(0.205 0 0)",
  secondary: "oklch(0.269 0 0)",
  muted: "oklch(0.269 0 0)",
  mutedForeground: "oklch(0.708 0 0)",
  accent: "oklch(0.269 0 0)",
  border: "oklch(1 0 0 / 10%)",
  destructive: "oklch(0.704 0.191 22.216)",
  accentBlue: "oklch(0.488 0.243 264.376)",
} as const;

export const FONTS = {
  sans: "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif",
  mono: "'SF Mono', 'Fira Code', 'Fira Mono', 'Roboto Mono', Menlo, Monaco, 'Cascadia Code', monospace",
} as const;

export const RADIUS = {
  base: "10px",
  sm: "6px",
  md: "8px",
  lg: "10px",
  xl: "14px",
  "2xl": "18px",
} as const;

export const BRAND = {
  monogram: "DS",
  name: "Dev Share",
  tagline: "Manage temporary developer environments with ease.",
} as const;

export const THEME = { COLORS, FONTS, RADIUS, BRAND } as const;
