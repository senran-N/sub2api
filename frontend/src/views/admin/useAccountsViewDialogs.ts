import { reactive, ref } from 'vue'
import type { Account } from '@/types'

interface MenuPosition {
  top: number
  left: number
}

const ACCOUNT_ACTION_MENU_WIDTH = 200
const ACCOUNT_ACTION_MENU_HEIGHT = 240
const ACCOUNT_ACTION_MENU_PADDING = 8

export function useAccountsViewDialogs() {
  const showCreate = ref(false)
  const showSync = ref(false)
  const showImportData = ref(false)
  const showExportDataDialog = ref(false)
  const includeProxyOnExport = ref(true)
  const showBulkEdit = ref(false)
  const showErrorPassthrough = ref(false)
  const showTLSFingerprintProfiles = ref(false)
  const menu = reactive<{
    show: boolean
    acc: Account | null
    pos: MenuPosition | null
  }>({
    show: false,
    acc: null,
    pos: null
  })

  const closeActionMenu = () => {
    menu.show = false
  }

  const syncMenuAccount = (nextAccount: Account) => {
    if (menu.acc?.id === nextAccount.id) {
      menu.acc = nextAccount
    }
  }

  const openMenu = (account: Account, event: MouseEvent) => {
    menu.acc = account

    const target = event.currentTarget as HTMLElement | null
    if (!target) {
      menu.pos = {
        top: event.clientY,
        left: event.clientX - ACCOUNT_ACTION_MENU_WIDTH
      }
      menu.show = true
      return
    }

    const rect = target.getBoundingClientRect()
    const viewportWidth = window.innerWidth
    const viewportHeight = window.innerHeight

    let left: number
    let top: number

    if (viewportWidth < 768) {
      left = Math.max(
        ACCOUNT_ACTION_MENU_PADDING,
        Math.min(
          rect.left + rect.width / 2 - ACCOUNT_ACTION_MENU_WIDTH / 2,
          viewportWidth - ACCOUNT_ACTION_MENU_WIDTH - ACCOUNT_ACTION_MENU_PADDING
        )
      )

      top = rect.bottom + 4
      if (top + ACCOUNT_ACTION_MENU_HEIGHT > viewportHeight - ACCOUNT_ACTION_MENU_PADDING) {
        top = rect.top - ACCOUNT_ACTION_MENU_HEIGHT - 4
        if (top < ACCOUNT_ACTION_MENU_PADDING) {
          top = ACCOUNT_ACTION_MENU_PADDING
        }
      }
    } else {
      left = Math.max(
        ACCOUNT_ACTION_MENU_PADDING,
        Math.min(
          event.clientX - ACCOUNT_ACTION_MENU_WIDTH,
          viewportWidth - ACCOUNT_ACTION_MENU_WIDTH - ACCOUNT_ACTION_MENU_PADDING
        )
      )
      top = event.clientY
      if (top + ACCOUNT_ACTION_MENU_HEIGHT > viewportHeight - ACCOUNT_ACTION_MENU_PADDING) {
        top = viewportHeight - ACCOUNT_ACTION_MENU_HEIGHT - ACCOUNT_ACTION_MENU_PADDING
      }
    }

    menu.pos = { top, left }
    menu.show = true
  }

  const openExportDataDialog = () => {
    includeProxyOnExport.value = true
    showExportDataDialog.value = true
  }

  return {
    showCreate,
    showSync,
    showImportData,
    showExportDataDialog,
    includeProxyOnExport,
    showBulkEdit,
    showErrorPassthrough,
    showTLSFingerprintProfiles,
    menu,
    closeActionMenu,
    syncMenuAccount,
    openMenu,
    openExportDataDialog
  }
}
