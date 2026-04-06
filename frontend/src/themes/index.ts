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
  document.documentElement.setAttribute(BRAND_THEME_ATTRIBUTE, normalized)
  document.body?.setAttribute(BRAND_THEME_ATTRIBUTE, normalized)
  return normalized
}

