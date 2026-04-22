export interface ThemeChartPalette {
  colors: string[]
}

export interface ThemeChartAlphaPalette {
  alpha: number
  colors: string[]
}

export interface ThemeChartTooltipColors {
  background: string
  text: string
}

export interface ThemeDoughnutChartConfig {
  cutout: string
  borderRadius: number
  spacing: number
  hoverOffset: number
}

export interface ThemeLineChartConfig {
  pointRadius: number
  pointHoverRadius: number
}

export interface ThemeChartTokens {
  palette: ThemeChartPalette
  alphaPalette: ThemeChartAlphaPalette
  tooltip: ThemeChartTooltipColors
  doughnut: ThemeDoughnutChartConfig
  line: ThemeLineChartConfig
}

const THEME_DEFAULTS: Record<string, string> = {
  '--theme-page-text': '#111827',
  '--theme-page-muted': '#4B5563',
  '--theme-page-border': 'rgba(17, 24, 39, 0.1)',
  '--theme-card-border': '#111827',
  '--theme-surface': '#ffffff',
  '--theme-surface-soft': '#F4F4F5',
  '--theme-surface-contrast': '#171717',
  '--theme-surface-contrast-text': '#ffffff',
  '--theme-accent': '#C43C00',
  '--theme-accent-rgb': '196 60 0',
  '--theme-success-rgb': '58 115 82',
  '--theme-warning-rgb': '138 82 0',
  '--theme-danger-rgb': '178 59 59',
  '--theme-info-rgb': '37 99 235',
  '--theme-brand-orange-rgb': '217 83 30',
  '--theme-brand-purple-rgb': '147 51 234',
  '--theme-brand-rose-rgb': '225 29 72',
  '--theme-chart-seq-1-rgb': '37 99 235',
  '--theme-chart-seq-2-rgb': '58 115 82',
  '--theme-chart-seq-3-rgb': '204 122 0',
  '--theme-chart-seq-4-rgb': '147 51 234',
  '--theme-chart-seq-5-rgb': '225 29 72',
  '--theme-chart-seq-6-rgb': '217 83 30',
  '--theme-chart-seq-7-rgb': '178 59 59',
  '--theme-chart-seq-8-rgb': '14 116 144',
  '--theme-chart-seq-9-rgb': '100 116 139',
  '--theme-chart-seq-10-rgb': '77 124 15',
  '--theme-chart-seq-11-rgb': '180 83 9',
  '--theme-chart-seq-12-rgb': '109 40 217',
  '--theme-chart-donut-cutout': '65',
  '--theme-chart-donut-border-radius': '3',
  '--theme-chart-donut-spacing': '3',
  '--theme-chart-point-radius': '0',
  '--theme-chart-point-hover-radius': '5'
}

const THEME_CHART_SEQUENCE = [
  '--theme-chart-seq-1-rgb',
  '--theme-chart-seq-2-rgb',
  '--theme-chart-seq-3-rgb',
  '--theme-chart-seq-4-rgb',
  '--theme-chart-seq-5-rgb',
  '--theme-chart-seq-6-rgb',
  '--theme-chart-seq-7-rgb',
  '--theme-chart-seq-8-rgb',
  '--theme-chart-seq-9-rgb',
  '--theme-chart-seq-10-rgb',
  '--theme-chart-seq-11-rgb',
  '--theme-chart-seq-12-rgb'
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
  return getThemeChartPalette().colors
}

export function getThemeChartSequenceAlpha(alpha: number): string[] {
  return getThemeChartAlphaPalette(alpha).colors
}

export function getThemeChartPalette(): ThemeChartPalette {
  return {
    colors: THEME_CHART_SEQUENCE.map(variableName => readThemeRgb(variableName))
  }
}

export function getThemeChartAlphaPalette(alpha: number): ThemeChartAlphaPalette {
  return {
    alpha,
    colors: THEME_CHART_SEQUENCE.map(variableName => readThemeRgbAlpha(variableName, alpha))
  }
}

export function getThemeChartTooltipColors(): ThemeChartTooltipColors {
  return {
    background: readThemeCssVariable('--theme-surface-contrast'),
    text: readThemeCssVariable('--theme-surface-contrast-text')
  }
}

export function readThemeCssVariableNumber(variableName: string, fallback = 0): number {
  const value = readThemeCssVariable(variableName)
  const parsed = parseFloat(value)
  return isNaN(parsed) ? fallback : parsed
}

export function getThemeDoughnutChartConfig(): ThemeDoughnutChartConfig {
  return {
    cutout: readThemeCssVariableNumber('--theme-chart-donut-cutout', 65) + '%',
    borderRadius: readThemeCssVariableNumber('--theme-chart-donut-border-radius', 3),
    spacing: readThemeCssVariableNumber('--theme-chart-donut-spacing', 3),
    hoverOffset: 8
  }
}

export function getThemeLineChartConfig(): ThemeLineChartConfig {
  return {
    pointRadius: readThemeCssVariableNumber('--theme-chart-point-radius', 0),
    pointHoverRadius: readThemeCssVariableNumber('--theme-chart-point-hover-radius', 5)
  }
}

export function getThemeChartTokens(alpha = 0.14): ThemeChartTokens {
  return {
    palette: getThemeChartPalette(),
    alphaPalette: getThemeChartAlphaPalette(alpha),
    tooltip: getThemeChartTooltipColors(),
    doughnut: getThemeDoughnutChartConfig(),
    line: getThemeLineChartConfig()
  }
}
