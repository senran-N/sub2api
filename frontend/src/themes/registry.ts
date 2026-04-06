export const FRONTEND_THEME_DEFAULT = 'factory' as const

export const FRONTEND_THEMES = [
  {
    id: 'factory',
    label: 'Factory',
    description: 'Light industrial grid, mono navigation, compact modular surfaces.'
  },
  {
    id: 'claude',
    label: 'Claude Editorial',
    description: 'Editorial serif headings, brutalist borders, warmer retro-geek surfaces.'
  }
] as const

export type FrontendThemeId = (typeof FRONTEND_THEMES)[number]['id']

export interface FrontendThemeDefinition {
  id: FrontendThemeId
  label: string
  description: string
}

export function normalizeFrontendTheme(value: string | null | undefined): FrontendThemeId {
  if (!value) return FRONTEND_THEME_DEFAULT
  const candidate = value.trim().toLowerCase()
  const matched = FRONTEND_THEMES.find((theme) => theme.id === candidate)
  return matched?.id ?? FRONTEND_THEME_DEFAULT
}

export function getFrontendThemeDefinition(
  value: string | null | undefined
): FrontendThemeDefinition {
  const id = normalizeFrontendTheme(value)
  return FRONTEND_THEMES.find((theme) => theme.id === id) ?? FRONTEND_THEMES[0]
}

