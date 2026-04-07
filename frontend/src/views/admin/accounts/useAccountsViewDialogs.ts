import { reactive, ref } from 'vue'
import type { Account } from '@/types'
import {
  calculateContextMenuPosition,
  clampFloatingPanelPosition,
  readThemePixelValue
} from '@/utils/floatingPanel'

interface MenuPosition {
  top: number
  left: number
}

const ACCOUNT_ACTION_MENU_WIDTH_FALLBACK = 208
const ACCOUNT_ACTION_MENU_HEIGHT_FALLBACK = 240
const FLOATING_PANEL_PADDING_FALLBACK = 8
const FLOATING_PANEL_GAP_FALLBACK = 4

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
    const menuWidth = readThemePixelValue('--theme-account-action-menu-width', ACCOUNT_ACTION_MENU_WIDTH_FALLBACK)
    const menuHeight = readThemePixelValue('--theme-account-action-menu-estimated-height', ACCOUNT_ACTION_MENU_HEIGHT_FALLBACK)
    const padding = readThemePixelValue('--theme-floating-panel-viewport-padding', FLOATING_PANEL_PADDING_FALLBACK)
    const gap = readThemePixelValue('--theme-floating-panel-gap', FLOATING_PANEL_GAP_FALLBACK)

    const target = event.currentTarget as HTMLElement | null
    if (!target) {
      menu.pos = clampFloatingPanelPosition(
        {
          top: event.clientY,
          left: event.clientX - menuWidth
        },
        {
          panelWidth: menuWidth,
          panelHeight: menuHeight,
          padding
        }
      )
      menu.show = true
      return
    }

    menu.pos = calculateContextMenuPosition({
      rect: target.getBoundingClientRect(),
      pointerX: event.clientX,
      pointerY: event.clientY,
      viewportWidth: window.innerWidth,
      viewportHeight: window.innerHeight,
      panelWidth: menuWidth,
      panelHeight: menuHeight,
      padding,
      gap
    })
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
