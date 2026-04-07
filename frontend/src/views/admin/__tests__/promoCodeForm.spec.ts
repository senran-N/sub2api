import { describe, expect, it } from 'vitest'
import type { PromoCode } from '@/types'
import {
  buildCreatePromoCodeRequest,
  buildPromoCodeListFilters,
  buildPromoRegisterLink,
  buildUpdatePromoCodeRequest,
  createDefaultPromoCodeCreateForm,
  createDefaultPromoCodeEditForm,
  getPromoCodeStatusClass,
  getPromoCodeStatusLabelKey,
  hydratePromoCodeEditForm,
  resetPromoCodeCreateForm
} from '../promocodes/promoCodeForm'

function createPromoCode(overrides: Partial<PromoCode> = {}): PromoCode {
  return {
    id: 1,
    code: 'WELCOME',
    bonus_amount: 10,
    max_uses: 100,
    used_count: 0,
    status: 'active',
    expires_at: null,
    notes: null,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    ...overrides
  }
}

describe('promoCodeForm helpers', () => {
  it('builds normalized list filters and register links', () => {
    expect(
      buildPromoCodeListFilters(
        {
          status: 'disabled'
        },
        '  spring  '
      )
    ).toEqual({
      status: 'disabled',
      search: 'spring'
    })

    expect(buildPromoRegisterLink('https://sub2api.dev', 'A B+C')).toBe(
      'https://sub2api.dev/register?promo=A%20B%2BC'
    )
  })

  it('resets and builds create payloads with optional fields omitted', () => {
    const createForm = createDefaultPromoCodeCreateForm()
    createForm.code = ''
    createForm.bonus_amount = 5
    createForm.max_uses = 0
    createForm.expires_at_str = ''
    createForm.notes = ''

    expect(buildCreatePromoCodeRequest(createForm)).toEqual({
      code: undefined,
      bonus_amount: 5,
      max_uses: 0,
      expires_at: undefined,
      notes: undefined
    })

    createForm.notes = 'dirty'
    resetPromoCodeCreateForm(createForm)
    expect(createForm).toEqual(createDefaultPromoCodeCreateForm())
  })

  it('hydrates edit form and builds update payloads with unix timestamp expiry', () => {
    const editForm = createDefaultPromoCodeEditForm()
    hydratePromoCodeEditForm(
      editForm,
      createPromoCode({
        code: 'VIP',
        bonus_amount: 25,
        max_uses: 5,
        status: 'disabled',
        expires_at: '2026-03-01T12:34:00Z',
        notes: 'internal'
      })
    )

    expect(editForm).toEqual({
      code: 'VIP',
      bonus_amount: 25,
      max_uses: 5,
      status: 'disabled',
      expires_at_str: '2026-03-01T12:34',
      notes: 'internal'
    })

    expect(buildUpdatePromoCodeRequest(editForm)).toEqual({
      code: 'VIP',
      bonus_amount: 25,
      max_uses: 5,
      status: 'disabled',
      expires_at: Math.floor(new Date('2026-03-01T12:34').getTime() / 1000),
      notes: 'internal'
    })
  })

  it('derives status classes and label keys from expiration and usage state', () => {
    const now = new Date('2026-04-01T00:00:00Z')

    expect(
      getPromoCodeStatusLabelKey(
        createPromoCode({ expires_at: '2026-03-01T00:00:00Z' }),
        now
      )
    ).toBe('admin.promo.statusExpired')
    expect(
      getPromoCodeStatusLabelKey(
        createPromoCode({ max_uses: 5, used_count: 5 }),
        now
      )
    ).toBe('admin.promo.statusMaxUsed')
    expect(
      getPromoCodeStatusLabelKey(
        createPromoCode({ status: 'disabled' }),
        now
      )
    ).toBe('admin.promo.statusDisabled')
    expect(getPromoCodeStatusClass(createPromoCode(), now)).toBe('badge-success')
  })
})
