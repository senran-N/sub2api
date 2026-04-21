import { describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import type { Account } from '@/types'
import { useAccountsViewState } from '../accounts/useAccountsViewState'

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

function createComposable(accountsList: Account[] = [createAccount()]) {
  const accounts = ref(accountsList)
  const selectedIds = ref<number[]>([])
  const hasPendingListSync = ref(false)
  const showCreate = ref(false)
  const showSync = ref(false)
  const showImportData = ref(false)
  const showExportDataDialog = ref(false)
  const showBulkEdit = ref(false)
  const showErrorPassthrough = ref(false)
  const toggleVisible = vi.fn()
  const clearSelection = vi.fn()
  const reload = vi.fn()
  const removeSelectedAccounts = vi.fn()
  const syncMenuAccount = vi.fn()
  const menu = {
    show: false,
    acc: null as Account | null
  }
  const pagination = {
    page: 1,
    page_size: 20,
    total: accountsList.length,
    pages: accountsList.length > 0 ? 1 : 0
  }
  const params = {
    platform: '',
    type: '',
    status: '',
    privacy_mode: '',
    group: '',
    search: ''
  }

  const state = useAccountsViewState({
    accounts,
    selectedIds,
    isSelected: (accountId) => selectedIds.value.includes(accountId),
    toggleVisible,
    clearSelection,
    reload,
    params,
    pagination,
    getHasPendingListSync: () => hasPendingListSync.value,
    setHasPendingListSync: (value) => {
      hasPendingListSync.value = value
    },
    removeSelectedAccounts,
    menu,
    syncMenuAccount,
    showCreate,
    showSync,
    showImportData,
    showExportDataDialog,
    showBulkEdit,
    showErrorPassthrough
  })

  return {
    state,
    accounts,
    selectedIds,
    hasPendingListSync,
    showImportData,
    showBulkEdit,
    toggleVisible,
    clearSelection,
    reload,
    removeSelectedAccounts,
    syncMenuAccount,
    menu,
    pagination,
    params
  }
}

describe('useAccountsViewState', () => {
  it('derives selected platforms and types and tracks modal visibility', () => {
    const setup = createComposable([
      createAccount({ id: 1, platform: 'openai', type: 'oauth' }),
      createAccount({ id: 2, platform: 'claude', type: 'session' }),
      createAccount({ id: 3, platform: 'openai', type: 'oauth' })
    ])

    setup.selectedIds.value = [1, 2]

    expect(setup.state.selPlatforms.value).toEqual(['openai', 'claude'])
    expect(setup.state.selTypes.value).toEqual(['oauth', 'session'])
    expect(setup.state.isAnyModalOpen.value).toBe(false)

    setup.state.showEdit.value = true
    expect(setup.state.isAnyModalOpen.value).toBe(true)
  })

  it('delegates visible-row selection and reload helpers', () => {
    const setup = createComposable()

    setup.state.toggleSelectAllVisible({
      target: { checked: true }
    } as Event)
    setup.state.handleBulkUpdated()
    setup.state.handleDataImported()

    expect(setup.toggleVisible).toHaveBeenCalledWith(true)
    expect(setup.showBulkEdit.value).toBe(false)
    expect(setup.showImportData.value).toBe(false)
    expect(setup.clearSelection).toHaveBeenCalledTimes(1)
    expect(setup.reload).toHaveBeenCalledTimes(2)
  })

  it('patches list updates, syncs refs, and removes filtered-out accounts', () => {
    const first = createAccount({ id: 1, name: 'Alpha' })
    const second = createAccount({ id: 2, name: 'Beta' })
    const setup = createComposable([first, second])

    setup.selectedIds.value = [2]
    setup.hasPendingListSync.value = true
    setup.menu.show = true
    setup.menu.acc = second
    setup.state.edAcc.value = second
    setup.params.search = 'Alpha'

    setup.state.patchAccountInList(createAccount({ id: 2, name: 'Gamma' }))

    expect(setup.accounts.value).toHaveLength(1)
    expect(setup.accounts.value[0].id).toBe(1)
    expect(setup.pagination.total).toBe(1)
    expect(setup.pagination.pages).toBe(1)
    expect(setup.hasPendingListSync.value).toBe(true)
    expect(setup.removeSelectedAccounts).toHaveBeenCalledWith([2])
    expect(setup.menu.show).toBe(false)
    expect(setup.menu.acc).toBeNull()

    setup.params.search = ''
    setup.state.edAcc.value = first
    setup.state.testingAcc.value = first
    setup.state.statsAcc.value = first
    setup.state.patchAccountInList(createAccount({ id: 1, name: 'Alpha Prime' }))

    expect(setup.accounts.value[0].name).toBe('Alpha Prime')
    expect(setup.state.edAcc.value?.name).toBe('Alpha Prime')
    expect(setup.state.testingAcc.value?.name).toBe('Alpha Prime')
    expect(setup.state.statsAcc.value?.name).toBe('Alpha Prime')
    expect(setup.syncMenuAccount).toHaveBeenCalledWith(
      expect.objectContaining({ id: 1, name: 'Alpha Prime' })
    )
  })
})
