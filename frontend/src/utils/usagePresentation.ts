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
    return 'bg-violet-100 text-violet-800 dark:bg-violet-900 dark:text-violet-200'
  }
  if (requestType === 'stream') {
    return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200'
  }
  if (requestType === 'sync') {
    return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200'
  }

  return 'bg-amber-100 text-amber-800 dark:bg-amber-900 dark:text-amber-200'
}
