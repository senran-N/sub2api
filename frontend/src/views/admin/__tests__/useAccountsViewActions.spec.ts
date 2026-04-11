import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import type { Account } from '@/types'
import { useAccountsViewActions } from '../accounts/useAccountsViewActions'

const {
  deleteAccount,
  batchClearError,
  batchRefresh,
  bulkUpdate,
  getAvailableModels,
  refreshCredentials,
  recoverState,
  resetAccountQuota,
  setPrivacy,
  setSchedulable
} = vi.hoisted(() => ({
  deleteAccount: vi.fn(),
  batchClearError: vi.fn(),
  batchRefresh: vi.fn(),
  bulkUpdate: vi.fn(),
  getAvailableModels: vi.fn(),
  refreshCredentials: vi.fn(),
  recoverState: vi.fn(),
  resetAccountQuota: vi.fn(),
  setPrivacy: vi.fn(),
  setSchedulable: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      delete: deleteAccount,
      batchClearError,
      batchRefresh,
      bulkUpdate,
      getAvailableModels,
      refreshCredentials,
      recoverState,
      resetAccountQuota,
      setPrivacy,
      setSchedulable
    }
  }
}))

function createAccount(overrides: Partial<Account> = {}): Account {
  return {
    id: 1,
    name: 'Account',
    platform: 'openai',
    type: 'oauth',
    credentials: {},
    extra: {},
    proxy_id: null,
    concurrency: 1,
    current_concurrency: 0,
    priority: 0,
    status: 'active',
    error_message: null,
    last_used_at: null,
    expires_at: null,
    auto_pause_on_expired: false,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    schedulable: true,
    rate_limited_at: null,
    rate_limit_reset_at: null,
    overload_until: null,
    temp_unschedulable_until: null,
    temp_unschedulable_reason: null,
    session_window_start: null,
    session_window_end: null,
    session_window_status: null,
    current_window_cost: 0,
    active_sessions: 0,
    ...overrides
  }
}

function createComposable(selectedIds: number[] = [1, 2]) {
  const showEdit = ref(false)
  const showTempUnsched = ref(false)
  const showDeleteDialog = ref(false)
  const showReAuth = ref(false)
  const showTest = ref(false)
  const showStats = ref(false)
  const showSchedulePanel = ref(false)
  const edAcc = ref<Account | null>(null)
  const tempUnschedAcc = ref<Account | null>(null)
  const deletingAcc = ref<Account | null>(null)
  const reAuthAcc = ref<Account | null>(null)
  const testingAcc = ref<Account | null>(null)
  const statsAcc = ref<Account | null>(null)
  const scheduleAcc = ref<Account | null>(null)
  const scheduleModelOptions = ref<{ value: string; label: string }[]>([])
  const togglingSchedulable = ref<number | null>(null)
  const clearSelection = vi.fn()
  const setSelectedIds = vi.fn()
  const load = vi.fn(async () => {})
  const reload = vi.fn(async () => {})
  const patchAccountInList = vi.fn()
  const updateSchedulableInList = vi.fn()
  const enterAutoRefreshSilentWindow = vi.fn()
  const showSuccess = vi.fn()
  const showError = vi.fn()

  const composable = useAccountsViewActions({
    showEdit,
    showTempUnsched,
    showDeleteDialog,
    showReAuth,
    showTest,
    showStats,
    showSchedulePanel,
    edAcc,
    tempUnschedAcc,
    deletingAcc,
    reAuthAcc,
    testingAcc,
    statsAcc,
    scheduleAcc,
    scheduleModelOptions,
    togglingSchedulable,
    getSelectedIds: () => selectedIds,
    confirmAction: () => true,
    clearSelection,
    setSelectedIds,
    load,
    reload,
    patchAccountInList,
    updateSchedulableInList,
    enterAutoRefreshSilentWindow,
    t: (key: string) => key,
    showSuccess,
    showError
  })

  return {
    composable,
    showEdit,
    showDeleteDialog,
    showSchedulePanel,
    deletingAcc,
    scheduleModelOptions,
    togglingSchedulable,
    clearSelection,
    setSelectedIds,
    load,
    reload,
    patchAccountInList,
    updateSchedulableInList,
    enterAutoRefreshSilentWindow,
    showSuccess,
    showError
  }
}

describe('useAccountsViewActions', () => {
  beforeEach(() => {
    deleteAccount.mockReset()
    batchClearError.mockReset()
    batchRefresh.mockReset()
    bulkUpdate.mockReset()
    getAvailableModels.mockReset()
    refreshCredentials.mockReset()
    recoverState.mockReset()
    resetAccountQuota.mockReset()
    setPrivacy.mockReset()
    setSchedulable.mockReset()
  })

  it('handles bulk reset success and reloads', async () => {
    const setup = createComposable()
    batchClearError.mockResolvedValue({ success: 2, failed: 0 })

    await setup.composable.handleBulkResetStatus()

    expect(batchClearError).toHaveBeenCalledWith([1, 2])
    expect(setup.clearSelection).toHaveBeenCalledTimes(1)
    expect(setup.showSuccess).toHaveBeenCalledWith(
      'admin.accounts.bulkActions.resetStatusSuccess'
    )
    expect(setup.reload).toHaveBeenCalledTimes(1)
  })

  it('surfaces request detail for bulk reset failures instead of stringifying objects', async () => {
    const setup = createComposable()
    batchClearError.mockRejectedValueOnce({
      response: {
        data: {
          detail: 'bulk-reset-failed'
        }
      }
    })

    await setup.composable.handleBulkResetStatus()

    expect(setup.showError).toHaveBeenCalledWith('bulk-reset-failed')
  })

  it('keeps failed accounts selected after bulk refresh partial failure', async () => {
    const setup = createComposable([11, 12, 13])
    batchRefresh.mockResolvedValue({
      total: 3,
      success: 2,
      failed: 1,
      errors: [{ account_id: 12, error: 'refresh failed' }]
    })

    await setup.composable.handleBulkRefreshToken()

    expect(batchRefresh).toHaveBeenCalledWith([11, 12, 13])
    expect(setup.setSelectedIds).toHaveBeenCalledWith([12])
    expect(setup.showError).toHaveBeenCalledWith('admin.accounts.bulkActions.partialSuccess')
    expect(setup.reload).toHaveBeenCalledTimes(1)
  })

  it('falls back to localized error text for bulk refresh failures without detail', async () => {
    const setup = createComposable([11, 12])
    batchRefresh.mockRejectedValueOnce(new Error('network down'))

    await setup.composable.handleBulkRefreshToken()

    expect(setup.showError).toHaveBeenCalledWith('network down')
  })

  it('handles unknown bulk schedulable results by reloading and restoring selection', async () => {
    const setup = createComposable([7, 8])
    bulkUpdate.mockResolvedValue({})

    await setup.composable.handleBulkToggleSchedulable(true)

    expect(bulkUpdate).toHaveBeenCalledWith([7, 8], { schedulable: true })
    expect(setup.showError).toHaveBeenCalledWith('admin.accounts.bulkSchedulableResultUnknown')
    expect(setup.setSelectedIds).toHaveBeenCalledWith([7, 8])
    expect(setup.load).toHaveBeenCalledTimes(1)
  })

  it('opens schedule panel and loads available models', async () => {
    const setup = createComposable()
    getAvailableModels.mockResolvedValue([
      { id: 'claude-1', display_name: 'Claude 1' }
    ])

    await setup.composable.handleSchedule(createAccount({ id: 9 }))

    expect(setup.showSchedulePanel.value).toBe(true)
    expect(setup.scheduleModelOptions.value).toEqual([
      { value: 'claude-1', label: 'Claude 1' }
    ])
  })

  it('confirms delete and reloads', async () => {
    const setup = createComposable()
    const account = createAccount({ id: 5 })
    deleteAccount.mockResolvedValue({ message: 'ok' })

    setup.composable.handleDelete(account)
    await setup.composable.confirmDelete()

    expect(setup.showDeleteDialog.value).toBe(false)
    expect(setup.deletingAcc.value).toBeNull()
    expect(deleteAccount).toHaveBeenCalledWith(5)
    expect(setup.reload).toHaveBeenCalledTimes(1)
  })

  it('toggles schedulable state and clears loading flag', async () => {
    const setup = createComposable([1])
    setSchedulable.mockResolvedValue({ schedulable: false })

    await setup.composable.handleToggleSchedulable(createAccount({ id: 3, schedulable: true }))

    expect(setup.updateSchedulableInList).toHaveBeenCalledWith([3], false)
    expect(setup.enterAutoRefreshSilentWindow).toHaveBeenCalledTimes(1)
    expect(setup.togglingSchedulable.value).toBeNull()
  })

  it('uses resolved request messages for recover state failures', async () => {
    const setup = createComposable()

    recoverState.mockRejectedValueOnce({
      response: {
        data: {
          detail: 'recover-state-blocked'
        }
      }
    })
    await setup.composable.handleRecoverState(createAccount({ id: 14 }))

    recoverState.mockRejectedValueOnce(new Error('recover unavailable'))
    await setup.composable.handleRecoverState(createAccount({ id: 15 }))

    expect(setup.showError).toHaveBeenNthCalledWith(1, 'recover-state-blocked')
    expect(setup.showError).toHaveBeenNthCalledWith(2, 'recover unavailable')
  })

  it('uses resolved request messages for privacy failures', async () => {
    const setup = createComposable()

    setPrivacy.mockRejectedValueOnce({
      response: {
        data: {
          detail: 'privacy-blocked'
        }
      }
    })
    await setup.composable.handleSetPrivacy(createAccount({ id: 16 }))

    setPrivacy.mockRejectedValueOnce({
      response: {
        data: {
          message: 'privacy message fallback'
        }
      }
    })
    await setup.composable.handleSetPrivacy(createAccount({ id: 17 }))

    expect(setup.showError).toHaveBeenNthCalledWith(1, 'privacy-blocked')
    expect(setup.showError).toHaveBeenNthCalledWith(2, 'privacy message fallback')
  })
})
