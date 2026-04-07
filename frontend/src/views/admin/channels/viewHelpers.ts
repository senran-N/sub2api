import type { GroupPlatform } from '@/types'

export const platformOrder: GroupPlatform[] = ['anthropic', 'openai', 'gemini', 'antigravity']

export function joinClassNames(...classNames: Array<string | false | null | undefined>): string {
  return classNames.filter(Boolean).join(' ')
}

export function getPlatformTextClass(platform: GroupPlatform | string): string {
  switch (platform) {
    case 'anthropic': return 'channel-view__tone-text channel-view__tone-text--brand-orange'
    case 'openai': return 'channel-view__tone-text channel-view__tone-text--success'
    case 'gemini': return 'channel-view__tone-text channel-view__tone-text--info'
    case 'antigravity': return 'channel-view__tone-text channel-view__tone-text--brand-purple'
    case 'sora': return 'channel-view__tone-text channel-view__tone-text--brand-rose'
    default: return 'channel-view__text-muted'
  }
}

export function getRateBadgeClass(platform: GroupPlatform | string): string {
  switch (platform) {
    case 'anthropic': return 'theme-chip theme-chip--compact theme-chip--brand-orange'
    case 'openai': return 'theme-chip theme-chip--compact theme-chip--success'
    case 'gemini': return 'theme-chip theme-chip--compact theme-chip--info'
    case 'antigravity': return 'theme-chip theme-chip--compact theme-chip--brand-purple'
    case 'sora': return 'theme-chip theme-chip--compact theme-chip--brand-rose'
    default: return 'theme-chip theme-chip--compact theme-chip--neutral'
  }
}

export function getPlatformToggleClasses(platform: GroupPlatform, active: boolean): string {
  return joinClassNames(
    'channel-view__platform-toggle inline-flex cursor-pointer items-center gap-1.5 border text-sm transition-colors',
    active && 'channel-view__platform-toggle--active',
    getPlatformTextClass(platform)
  )
}

export function getGroupChipClasses(platform: GroupPlatform, selected: boolean, disabled: boolean): string {
  return joinClassNames(
    'channel-view__group-chip inline-flex cursor-pointer items-center gap-1.5 border text-xs transition-colors',
    selected && 'channel-view__group-chip--selected',
    disabled && 'opacity-40',
    getPlatformTextClass(platform)
  )
}

export function getActionButtonClasses(tone: 'info' | 'danger'): string {
  return joinClassNames(
    'channel-view__action-button flex flex-col items-center gap-0.5 transition-colors',
    tone === 'info' ? 'channel-view__action-button--info' : 'channel-view__action-button--danger'
  )
}

export function formatDate(value: string): string {
  if (!value) return '-'
  return new Date(value).toLocaleDateString()
}
