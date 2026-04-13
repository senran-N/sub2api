import { computed, ref, type Ref } from 'vue'
import type { Account, AccountPlatform, AccountType, SelectOption } from '@/types'
import {
  patchAccountList,
  updateSchedulableAccounts,
  type AccountListPagination,
  type AccountListQuery
} from './accountsList'

interface AccountsViewMenuState {
  show: boolean
  acc: Account | null
}

interface AccountsViewStateOptions {
  accounts: Ref<Account[]>
  isSelected: (accountId: number) => boolean
  toggleVisible: (selected: boolean) => void
  clearSelection: () => void
  reload: () => void | Promise<void>
  params: Pick<
    AccountListQuery,
    'platform' | 'type' | 'status' | 'privacy_mode' | 'group' | 'search'
  >
  pagination: AccountListPagination
  getHasPendingListSync: () => boolean
  setHasPendingListSync: (value: boolean) => void
  removeSelectedAccounts: (accountIds: number[]) => void
  menu: AccountsViewMenuState
  syncMenuAccount: (account: Account) => void
  showCreate: Ref<boolean>
  showSync: Ref<boolean>
  showImportData: Ref<boolean>
  showExportDataDialog: Ref<boolean>
  showBulkEdit: Ref<boolean>
  showErrorPassthrough: Ref<boolean>
}

export function useAccountsViewState(options: AccountsViewStateOptions) {
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

  const scheduleModelOptions = ref<SelectOption[]>([])
  const togglingSchedulable = ref<number | null>(null)

  const selPlatforms = computed<AccountPlatform[]>(() => {
    const platforms = new Set(
      options.accounts.value
        .filter((account) => options.isSelected(account.id))
        .map((account) => account.platform)
    )
    return [...platforms]
  })

  const selTypes = computed<AccountType[]>(() => {
    const types = new Set(
      options.accounts.value
        .filter((account) => options.isSelected(account.id))
        .map((account) => account.type)
    )
    return [...types]
  })

  const isAnyModalOpen = computed(() => {
    return (
      options.showCreate.value ||
      showEdit.value ||
      options.showSync.value ||
      options.showImportData.value ||
      options.showExportDataDialog.value ||
      options.showBulkEdit.value ||
      showTempUnsched.value ||
      showDeleteDialog.value ||
      showReAuth.value ||
      showTest.value ||
      showStats.value ||
      showSchedulePanel.value ||
      options.showErrorPassthrough.value
    )
  })

  const syncAccountRefs = (nextAccount: Account) => {
    if (edAcc.value?.id === nextAccount.id) edAcc.value = nextAccount
    if (reAuthAcc.value?.id === nextAccount.id) reAuthAcc.value = nextAccount
    if (tempUnschedAcc.value?.id === nextAccount.id) tempUnschedAcc.value = nextAccount
    if (deletingAcc.value?.id === nextAccount.id) deletingAcc.value = nextAccount
    options.syncMenuAccount(nextAccount)
  }

  const toggleSelectAllVisible = (event: Event) => {
    const target = event.target as HTMLInputElement | null
    options.toggleVisible(Boolean(target?.checked))
  }

  const updateSchedulableInList = (accountIds: number[], schedulable: boolean) => {
    options.accounts.value = updateSchedulableAccounts(
      options.accounts.value,
      accountIds,
      schedulable
    )
  }

  const handleBulkUpdated = () => {
    options.showBulkEdit.value = false
    options.clearSelection()
    void options.reload()
  }

  const handleDataImported = () => {
    options.showImportData.value = false
    void options.reload()
  }

  const patchAccountInList = (updatedAccount: Account) => {
    const result = patchAccountList(
      options.accounts.value,
      updatedAccount,
      {
        platform: options.params.platform,
        type: options.params.type,
        status: options.params.status,
        privacy_mode: options.params.privacy_mode,
        group: options.params.group,
        search: options.params.search
      },
      options.pagination,
      options.getHasPendingListSync(),
      options.menu.acc?.id ?? null
    )

    if (result.patchedAccount === null && result.removedAccountId === null) {
      return
    }

    options.accounts.value = result.accounts
    options.pagination.page = result.pagination.page
    options.pagination.total = result.pagination.total
    options.pagination.pages = result.pagination.pages
    options.setHasPendingListSync(result.hasPendingListSync)

    if (result.removedAccountId !== null) {
      options.removeSelectedAccounts([result.removedAccountId])
    }

    if (result.shouldCloseMenu) {
      options.menu.show = false
      options.menu.acc = null
    }

    if (result.patchedAccount) {
      syncAccountRefs(result.patchedAccount)
    }
  }

  return {
    selPlatforms,
    selTypes,
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
    isAnyModalOpen,
    syncAccountRefs,
    toggleSelectAllVisible,
    updateSchedulableInList,
    handleBulkUpdated,
    handleDataImported,
    patchAccountInList
  }
}
