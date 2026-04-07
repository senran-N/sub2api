import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { PromoCode, PromoCodeUsage } from '@/types'
import { usePromoCodesViewActions } from '../promocodes/usePromoCodesViewActions'

const { createPromoCode, updatePromoCode, deletePromoCode, getPromoUsages } = vi.hoisted(() => ({
  createPromoCode: vi.fn(),
  updatePromoCode: vi.fn(),
  deletePromoCode: vi.fn(),
  getPromoUsages: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    promo: {
      create: createPromoCode,
      update: updatePromoCode,
      delete: deletePromoCode,
      getUsages: getPromoUsages
    }
  }
}))

function createCode(overrides: Partial<PromoCode> = {}): PromoCode {
  return {
    id: 1,
    code: 'WELCOME',
    bonus_amount: 10,
    max_uses: 100,
    used_count: 0,
    status: 'active',
    expires_at: null,
    notes: null,
    created_at: '2026-04-04T00:00:00Z',
    updated_at: '2026-04-04T00:00:00Z',
    ...overrides
  }
}

function createUsage(overrides: Partial<PromoCodeUsage> = {}): PromoCodeUsage {
  return {
    id: 1,
    promo_code_id: 1,
    user_id: 2,
    bonus_amount: 10,
    used_at: '2026-04-04T00:00:00Z',
    ...overrides
  } as PromoCodeUsage
}

describe('usePromoCodesViewActions', () => {
  beforeEach(() => {
    createPromoCode.mockReset()
    updatePromoCode.mockReset()
    deletePromoCode.mockReset()
    getPromoUsages.mockReset()

    createPromoCode.mockResolvedValue(createCode())
    updatePromoCode.mockResolvedValue(createCode({ code: 'VIP' }))
    deletePromoCode.mockResolvedValue({ message: 'ok' })
    getPromoUsages.mockResolvedValue({
      items: [createUsage()],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1
    })
  })

  it('creates promo codes and resets the create dialog state', async () => {
    const reloadCodes = vi.fn().mockResolvedValue(undefined)
    const showSuccess = vi.fn()
    const showError = vi.fn()
    const copyToClipboard = vi.fn().mockResolvedValue(true)
    const actions = usePromoCodesViewActions({
      origin: 'https://sub2api.dev',
      t: (key: string) => key,
      showSuccess,
      showError,
      copyToClipboard,
      reloadCodes
    })

    actions.showCreateDialog.value = true
    actions.createForm.code = 'WELCOME'
    actions.createForm.bonus_amount = 5
    actions.createForm.notes = 'internal'
    await actions.handleCreate()

    expect(createPromoCode).toHaveBeenCalledWith({
      code: 'WELCOME',
      bonus_amount: 5,
      max_uses: 0,
      expires_at: undefined,
      notes: 'internal'
    })
    expect(showSuccess).toHaveBeenCalledWith('admin.promo.codeCreated')
    expect(reloadCodes).toHaveBeenCalledTimes(1)
    expect(actions.showCreateDialog.value).toBe(false)
    expect(actions.createForm.code).toBe('')
  })

  it('edits, updates, copies links, and deletes promo codes', async () => {
    const reloadCodes = vi.fn().mockResolvedValue(undefined)
    const showSuccess = vi.fn()
    const showError = vi.fn()
    const copyToClipboard = vi.fn().mockResolvedValue(true)
    const actions = usePromoCodesViewActions({
      origin: 'https://sub2api.dev',
      t: (key: string) => key,
      showSuccess,
      showError,
      copyToClipboard,
      reloadCodes
    })

    const code = createCode({
      expires_at: '2026-04-05T12:34:00Z',
      notes: 'internal'
    })
    actions.handleEdit(code)
    expect(actions.showEditDialog.value).toBe(true)
    expect(actions.editForm.code).toBe('WELCOME')

    actions.editForm.code = 'VIP'
    const expectedExpiry = Math.floor(new Date(actions.editForm.expires_at_str).getTime() / 1000)
    await actions.handleUpdate()
    expect(updatePromoCode).toHaveBeenCalledWith(1, {
      code: 'VIP',
      bonus_amount: 10,
      max_uses: 100,
      status: 'active',
      expires_at: expectedExpiry,
      notes: 'internal'
    })

    await actions.copyRegisterLink(code)
    expect(copyToClipboard).toHaveBeenCalledWith(
      'https://sub2api.dev/register?promo=WELCOME',
      'admin.promo.registerLinkCopied'
    )

    actions.handleDelete(code)
    await actions.confirmDelete()
    expect(deletePromoCode).toHaveBeenCalledWith(1)
    expect(showSuccess).toHaveBeenCalledWith('admin.promo.codeDeleted')
  })

  it('loads usages and resets usages dialog state', async () => {
    const actions = usePromoCodesViewActions({
      origin: 'https://sub2api.dev',
      t: (key: string) => key,
      showSuccess: vi.fn(),
      showError: vi.fn(),
      copyToClipboard: vi.fn().mockResolvedValue(true),
      reloadCodes: vi.fn().mockResolvedValue(undefined)
    })

    await actions.handleViewUsages(createCode())
    expect(getPromoUsages).toHaveBeenCalledWith(1, 1, 20)
    expect(actions.usages.value).toHaveLength(1)

    actions.handleUsagesPageChange(2)
    expect(getPromoUsages).toHaveBeenLastCalledWith(1, 2, 20)

    actions.handleUsagesPageSizeChange(50)
    expect(getPromoUsages).toHaveBeenLastCalledWith(1, 1, 50)

    actions.closeUsagesDialog()
    expect(actions.showUsagesDialog.value).toBe(false)
    expect(actions.usages.value).toEqual([])
    expect(actions.usagesPageSize.value).toBe(20)
  })
})
