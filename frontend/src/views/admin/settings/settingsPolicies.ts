export function sanitizeRectifierPatterns(
  patterns: string[] | null | undefined
): string[] {
  if (!Array.isArray(patterns)) {
    return []
  }

  return patterns
    .map((pattern) => pattern.trim())
    .filter((pattern) => pattern.length > 0)
}

export function maskSettingsApiKey(key: string): string {
  return `${key.substring(0, 10)}...${key.slice(-4)}`
}

const SETTINGS_BETA_DISPLAY_NAMES: Record<string, string> = {
  'fast-mode-2026-02-01': 'Fast Mode',
  'context-1m-2025-08-07': 'Context 1M'
}

export function getSettingsBetaDisplayName(token: string): string {
  return SETTINGS_BETA_DISPLAY_NAMES[token] || token
}
