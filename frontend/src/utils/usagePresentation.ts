import type { UsageRequestTypeLike } from '@/utils/usageRequestType'
import { resolveUsageRequestType } from '@/utils/usageRequestType'

export type UsageLabelTranslator = (key: string) => string

export function getUsageRequestTypeLabelKey(value: UsageRequestTypeLike): string {
  const requestType = resolveUsageRequestType(value)

  if (requestType === 'ws_v2') {
    return 'usage.ws'
  }
  if (requestType === 'stream') {
    return 'usage.stream'
  }
  if (requestType === 'sync') {
    return 'usage.sync'
  }

  return 'usage.unknown'
}

export function getUsageRequestTypeLabel(
  value: UsageRequestTypeLike,
  t: UsageLabelTranslator
): string {
  return t(getUsageRequestTypeLabelKey(value))
}

export function getUsageRequestTypeExportText(value: UsageRequestTypeLike): string {
  const requestType = resolveUsageRequestType(value)

  if (requestType === 'ws_v2') {
    return 'WS'
  }
  if (requestType === 'stream') {
    return 'Stream'
  }
  if (requestType === 'sync') {
    return 'Sync'
  }

  return 'Unknown'
}

export function getUsageRequestTypeBadgeClass(value: UsageRequestTypeLike): string {
  const requestType = resolveUsageRequestType(value)

  if (requestType === 'ws_v2') {
    return 'theme-chip theme-chip--regular theme-chip--brand-purple'
  }
  if (requestType === 'stream') {
    return 'theme-chip theme-chip--regular theme-chip--info'
  }
  if (requestType === 'sync') {
    return 'theme-chip theme-chip--regular theme-chip--neutral'
  }

  return 'theme-chip theme-chip--regular theme-chip--warning'
}
