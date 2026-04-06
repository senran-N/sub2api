const THEME_DEFAULTS: Record<string, string> = {
  '--theme-page-text': '#171717',
  '--theme-page-muted': '#6b6b6b',
  '--theme-page-border': '#dbdbd6',
  '--theme-card-border': '#d8d8d3',
  '--theme-surface': '#ffffff',
  '--theme-surface-soft': '#fafaf8',
  '--theme-surface-contrast': '#171717',
  '--theme-surface-contrast-text': '#ffffff',
  '--theme-accent': '#ff5a1f',
  '--theme-accent-rgb': '255 90 31',
  '--theme-success-rgb': '22 163 74',
  '--theme-warning-rgb': '217 119 6',
  '--theme-danger-rgb': '220 38 38',
  '--theme-info-rgb': '37 99 235',
  '--theme-brand-orange-rgb': '234 88 12',
  '--theme-brand-purple-rgb': '147 51 234',
  '--theme-brand-rose-rgb': '225 29 72'
}

const THEME_CHART_SEQUENCE = [
  '--theme-info-rgb',
  '--theme-success-rgb',
  '--theme-warning-rgb',
  '--theme-danger-rgb',
  '--theme-brand-purple-rgb',
  '--theme-brand-rose-rgb',
  '--theme-accent-rgb',
  '--theme-brand-orange-rgb',
  '--theme-info-rgb',
  '--theme-success-rgb',
  '--theme-warning-rgb',
  '--theme-brand-purple-rgb'
] as const

function resolveThemeFallback(variableName: string, fallback?: string): string {
  return fallback ?? THEME_DEFAULTS[variableName] ?? ''
}

function isRgbChannelValue(value: string): boolean {
  return /^\d{1,3}\s+\d{1,3}\s+\d{1,3}$/.test(value.trim())
}

function formatRgbValue(value: string): string {
  return isRgbChannelValue(value) ? `rgb(${value})` : value
}

function formatRgbAlphaValue(value: string, alpha: number): string {
  return isRgbChannelValue(value) ? `rgb(${value} / ${alpha})` : value
}

export function readThemeCssVariable(variableName: string, fallback?: string): string {
  const resolvedFallback = resolveThemeFallback(variableName, fallback)
  if (typeof document === 'undefined') {
    return resolvedFallback
  }

  const value = getComputedStyle(document.documentElement).getPropertyValue(variableName).trim()
  return value || resolvedFallback
}

export function readThemeRgb(variableName: string, fallback?: string): string {
  const rawValue = readThemeCssVariable(variableName, '')
  return rawValue ? `rgb(${rawValue})` : formatRgbValue(resolveThemeFallback(variableName, fallback))
}

export function readThemeRgbAlpha(variableName: string, alpha: number, fallback?: string): string {
  const rawValue = readThemeCssVariable(variableName, '')
  return rawValue
    ? `rgb(${rawValue} / ${alpha})`
    : formatRgbAlphaValue(resolveThemeFallback(variableName, fallback), alpha)
}

export function getThemeChartSequence(): string[] {
  return THEME_CHART_SEQUENCE.map(variableName => readThemeRgb(variableName))
}

export function getThemeChartSequenceAlpha(alpha: number): string[] {
  return THEME_CHART_SEQUENCE.map(variableName => readThemeRgbAlpha(variableName, alpha))
}

export function getThemeChartTooltipColors(): { background: string; text: string } {
  return {
    background: readThemeCssVariable('--theme-surface-contrast'),
    text: readThemeCssVariable('--theme-surface-contrast-text')
  }
}
