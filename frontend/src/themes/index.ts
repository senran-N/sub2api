import {
  FRONTEND_THEMES,
  FRONTEND_THEME_DEFAULT,
  getFrontendThemeDefinition,
  normalizeFrontendTheme,
  type FrontendThemeDefinition,
  type FrontendThemeId
} from './registry'

const BRAND_THEME_ATTRIBUTE = 'data-brand-theme'

export {
  FRONTEND_THEMES,
  FRONTEND_THEME_DEFAULT,
  getFrontendThemeDefinition,
  normalizeFrontendTheme,
  BRAND_THEME_ATTRIBUTE
}

export type { FrontendThemeId, FrontendThemeDefinition }

export function applyFrontendTheme(value: string | null | undefined): FrontendThemeId {
  const normalized = normalizeFrontendTheme(value)
  // Set only on <html>. CSS custom properties inherit top-down, so descendants
  // pick up the value without an explicit body attribute. Setting it on <body>
  // would place light-scope [data-brand-theme] variables on a nearer ancestor
  // than <html>, silently masking all .dark[data-brand-theme] overrides.
  document.documentElement.setAttribute(BRAND_THEME_ATTRIBUTE, normalized)
  return normalized
}
