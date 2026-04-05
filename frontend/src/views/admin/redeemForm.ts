import type { Group, GroupPlatform, RedeemCode, RedeemCodeType, SelectOption, SubscriptionType } from '@/types'

export type RedeemStatusFilter = '' | 'unused' | 'used' | 'expired'
export type AppliedRedeemStatusFilter = Exclude<RedeemStatusFilter, ''>

export interface RedeemFiltersState {
  type: '' | RedeemCodeType
  status: RedeemStatusFilter
}

export interface RedeemGenerationForm {
  type: RedeemCodeType
  value: number
  count: number
  group_id: number | null
  validity_days: number
}

export interface RedeemGroupOption extends SelectOption {
  value: number
  label: string
  description: string | null
  platform: GroupPlatform
  subscriptionType: SubscriptionType
  rate: number
}

export function createDefaultRedeemFilters(): RedeemFiltersState {
  return {
    type: '',
    status: ''
  }
}

export function createDefaultRedeemGenerationForm(): RedeemGenerationForm {
  return {
    type: 'balance',
    value: 10,
    count: 1,
    group_id: null,
    validity_days: 30
  }
}

export function syncRedeemGenerationFormValue(form: RedeemGenerationForm): void {
  if (form.type === 'invitation') {
    form.value = 0
    return
  }

  if (form.value === 0) {
    form.value = 10
  }
}

export function resetRedeemGenerationSubscriptionFields(form: RedeemGenerationForm): void {
  form.group_id = null
  form.validity_days = 30
}

export function buildRedeemListFilters(
  filters: RedeemFiltersState,
  searchQuery: string
): {
  type?: RedeemCodeType
  status?: AppliedRedeemStatusFilter
  search?: string
} {
  return {
    type: filters.type || undefined,
    status: filters.status || undefined,
    search: searchQuery || undefined
  }
}

export function buildRedeemExportFilters(filters: RedeemFiltersState): {
  type?: RedeemCodeType
  status?: AppliedRedeemStatusFilter
} {
  return {
    type: filters.type || undefined,
    status: filters.status || undefined
  }
}

export function buildRedeemSubscriptionGroupOptions(groups: Group[]): RedeemGroupOption[] {
  return groups
    .filter((group) => group.subscription_type === 'subscription')
    .map((group) => ({
      value: group.id,
      label: group.name,
      description: group.description,
      platform: group.platform,
      subscriptionType: group.subscription_type,
      rate: group.rate_multiplier
    }))
}

export function buildGeneratedRedeemCodesText(codes: RedeemCode[]): string {
  return codes.map((code) => code.code).join('\n')
}

export function getGeneratedRedeemTextareaHeight(codeCount: number): number {
  const lineHeight = 24
  const padding = 24
  const minHeight = 60
  const maxHeight = 240
  return Math.min(Math.max(codeCount * lineHeight + padding, minHeight), maxHeight)
}
