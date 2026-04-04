import { beforeEach, describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'
import { useGroupsViewManagement } from '../useGroupsViewManagement'
import type { AdminGroup } from '@/types'

const { createGroupRequest, updateGroupRequest, deleteGroupRequest } = vi.hoisted(() => ({
  createGroupRequest: vi.fn(),
  updateGroupRequest: vi.fn(),
  deleteGroupRequest: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    groups: {
      create: createGroupRequest,
      update: updateGroupRequest,
      delete: deleteGroupRequest
    },
    accounts: {
      list: vi.fn(),
      getById: vi.fn()
    }
  }
}))

function createGroup(overrides: Partial<AdminGroup> = {}): AdminGroup {
  return {
    id: 1,
    name: 'Alpha',
    description: 'desc',
    platform: 'anthropic',
    rate_multiplier: 1,
    is_exclusive: false,
    status: 'active',
    subscription_type: 'standard',
    daily_limit_usd: null,
    weekly_limit_usd: null,
    monthly_limit_usd: null,
    image_price_1k: null,
    image_price_2k: null,
    image_price_4k: null,
    sora_image_price_360: null,
    sora_image_price_540: null,
    sora_video_price_per_request: null,
    sora_video_price_per_request_hd: null,
    sora_storage_quota_bytes: 0,
    claude_code_only: false,
    fallback_group_id: null,
    fallback_group_id_on_invalid_request: null,
    allow_messages_dispatch: false,
    default_mapped_model: '',
    require_oauth_only: false,
    require_privacy_set: false,
    model_routing: null,
    model_routing_enabled: false,
    mcp_xml_inject: true,
    simulate_claude_max_enabled: false,
    sort_order: 10,
    created_at: '2026-04-04T00:00:00Z',
    updated_at: '2026-04-04T00:00:00Z',
    ...overrides
  }
}

describe('useGroupsViewManagement', () => {
  beforeEach(() => {
    createGroupRequest.mockReset()
    updateGroupRequest.mockReset()
    deleteGroupRequest.mockReset()

    createGroupRequest.mockResolvedValue(createGroup())
    updateGroupRequest.mockResolvedValue(createGroup({ name: 'Updated' }))
    deleteGroupRequest.mockResolvedValue({ message: 'ok' })
  })

  it('creates groups and advances onboarding after success', async () => {
    const loadGroups = vi.fn().mockResolvedValue(undefined)
    const nextStep = vi.fn()
    const showError = vi.fn()
    const state = useGroupsViewManagement({
      t: (key: string) => key,
      showError,
      showSuccess: vi.fn(),
      loadGroups,
      isCurrentOnboardingStep: vi.fn(() => true),
      advanceOnboarding: nextStep
    })

    await state.handleCreateGroup()
    expect(showError).toHaveBeenCalledWith('admin.groups.nameRequired')

    state.showCreateModal.value = true
    state.createForm.name = 'New Group'
    state.createForm.subscription_type = 'subscription'
    await nextTick()
    expect(state.createForm.is_exclusive).toBe(true)

    await state.handleCreateGroup()
    expect(createGroupRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        name: 'New Group',
        is_exclusive: true
      })
    )
    expect(loadGroups).toHaveBeenCalledTimes(1)
    expect(nextStep).toHaveBeenCalledWith(500)
    expect(state.showCreateModal.value).toBe(false)
  })

  it('applies platform rules and handles edit/update/delete flows', async () => {
    const loadGroups = vi.fn().mockResolvedValue(undefined)
    const showSuccess = vi.fn()
    const state = useGroupsViewManagement({
      t: (key: string, params?: Record<string, unknown>) =>
        params ? `${key}:${JSON.stringify(params)}` : key,
      showError: vi.fn(),
      showSuccess,
      loadGroups,
      isCurrentOnboardingStep: vi.fn(() => false),
      advanceOnboarding: vi.fn()
    })

    state.createForm.allow_messages_dispatch = true
    state.createForm.default_mapped_model = 'gpt-5.4'
    state.createForm.require_oauth_only = true
    state.createForm.require_privacy_set = true
    state.createForm.fallback_group_id_on_invalid_request = 2
    state.createForm.platform = 'sora'
    await nextTick()
    expect(state.createForm.allow_messages_dispatch).toBe(false)
    expect(state.createForm.default_mapped_model).toBe('')
    expect(state.createForm.require_oauth_only).toBe(false)
    expect(state.createForm.require_privacy_set).toBe(false)
    expect(state.createForm.fallback_group_id_on_invalid_request).toBeNull()

    const group = createGroup()
    await state.handleEdit(group)
    expect(state.showEditModal.value).toBe(true)
    expect(state.editForm.name).toBe('Alpha')

    state.editForm.name = 'Updated'
    await state.handleUpdateGroup()
    expect(updateGroupRequest).toHaveBeenCalledWith(
      1,
      expect.objectContaining({
        name: 'Updated'
      })
    )

    state.handleRateMultipliers(group)
    expect(state.showRateMultipliersModal.value).toBe(true)
    expect(state.rateMultipliersGroup.value?.id).toBe(1)

    state.handleDelete(createGroup({ subscription_type: 'subscription' }))
    expect(state.deleteConfirmMessage.value).toContain('admin.groups.deleteConfirmSubscription')
    await state.confirmDelete()
    expect(deleteGroupRequest).toHaveBeenCalledWith(1)
    expect(showSuccess).toHaveBeenCalledWith('admin.groups.groupDeleted')
  })
})
