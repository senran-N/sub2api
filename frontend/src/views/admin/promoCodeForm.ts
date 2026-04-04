import type {
  CreatePromoCodeRequest,
  PromoCode,
  UpdatePromoCodeRequest
} from '@/types'

export type PromoCodeStatusState = 'expired' | 'max_used' | 'active' | 'disabled'

export interface PromoCodeFiltersState {
  status: '' | 'active' | 'disabled'
}

export interface PromoCodeCreateForm {
  code: string
  bonus_amount: number
  max_uses: number
  expires_at_str: string
  notes: string
}

export interface PromoCodeEditForm {
  code: string
  bonus_amount: number
  max_uses: number
  status: 'active' | 'disabled'
  expires_at_str: string
  notes: string
}

export function createDefaultPromoCodeCreateForm(): PromoCodeCreateForm {
  return {
    code: '',
    bonus_amount: 1,
    max_uses: 0,
    expires_at_str: '',
    notes: ''
  }
}

export function createDefaultPromoCodeEditForm(): PromoCodeEditForm {
  return {
    code: '',
    bonus_amount: 0,
    max_uses: 0,
    status: 'active',
    expires_at_str: '',
    notes: ''
  }
}

export function resetPromoCodeCreateForm(form: PromoCodeCreateForm): void {
  Object.assign(form, createDefaultPromoCodeCreateForm())
}

export function resetPromoCodeEditForm(form: PromoCodeEditForm): void {
  Object.assign(form, createDefaultPromoCodeEditForm())
}

export function hydratePromoCodeEditForm(form: PromoCodeEditForm, code: PromoCode): void {
  Object.assign(form, createDefaultPromoCodeEditForm(), {
    code: code.code,
    bonus_amount: code.bonus_amount,
    max_uses: code.max_uses,
    status: code.status,
    expires_at_str: formatPromoCodeExpiryForInput(code.expires_at),
    notes: code.notes || ''
  })
}

export function buildPromoCodeListFilters(
  filters: PromoCodeFiltersState,
  searchQuery: string
): { status?: 'active' | 'disabled'; search?: string } {
  return {
    status: filters.status || undefined,
    search: searchQuery.trim() || undefined
  }
}

export function buildCreatePromoCodeRequest(form: PromoCodeCreateForm): CreatePromoCodeRequest {
  return {
    code: form.code || undefined,
    bonus_amount: form.bonus_amount,
    max_uses: form.max_uses,
    expires_at: parsePromoCodeExpiryInput(form.expires_at_str, undefined),
    notes: form.notes || undefined
  }
}

export function buildUpdatePromoCodeRequest(form: PromoCodeEditForm): UpdatePromoCodeRequest {
  return {
    code: form.code,
    bonus_amount: form.bonus_amount,
    max_uses: form.max_uses,
    status: form.status,
    expires_at: parsePromoCodeExpiryInput(form.expires_at_str, 0),
    notes: form.notes
  }
}

export function getPromoCodeStatusState(code: PromoCode, now: Date = new Date()): PromoCodeStatusState {
  if (code.expires_at && new Date(code.expires_at) < now) {
    return 'expired'
  }
  if (code.max_uses > 0 && code.used_count >= code.max_uses) {
    return 'max_used'
  }
  return code.status === 'active' ? 'active' : 'disabled'
}

export function getPromoCodeStatusClass(code: PromoCode, now?: Date): string {
  const state = getPromoCodeStatusState(code, now)
  if (state === 'expired') {
    return 'badge-danger'
  }
  if (state === 'active') {
    return 'badge-success'
  }
  return 'badge-gray'
}

export function getPromoCodeStatusLabelKey(code: PromoCode, now?: Date): string {
  const state = getPromoCodeStatusState(code, now)
  if (state === 'expired') {
    return 'admin.promo.statusExpired'
  }
  if (state === 'max_used') {
    return 'admin.promo.statusMaxUsed'
  }
  return state === 'active' ? 'admin.promo.statusActive' : 'admin.promo.statusDisabled'
}

export function buildPromoRegisterLink(origin: string, code: string): string {
  return `${origin}/register?promo=${encodeURIComponent(code)}`
}

function parsePromoCodeExpiryInput(
  value: string,
  emptyValue: number | undefined
): number | undefined {
  if (!value) {
    return emptyValue
  }
  return Math.floor(new Date(value).getTime() / 1000)
}

function formatPromoCodeExpiryForInput(value: string | null): string {
  if (!value) {
    return ''
  }
  return new Date(value).toISOString().slice(0, 16)
}
