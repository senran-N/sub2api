import { ref } from 'vue'
import type { AdminUser } from '@/types'
import {
  calculateContextMenuPosition,
  type FloatingPanelPosition,
  readThemePixelValue
} from '@/utils/floatingPanel'

const USER_ACTION_MENU_WIDTH_FALLBACK = 200
const USER_ACTION_MENU_HEIGHT_FALLBACK = 240
const FLOATING_PANEL_PADDING_FALLBACK = 8
const FLOATING_PANEL_GAP_FALLBACK = 4

export interface UserActionMenuPosition extends FloatingPanelPosition {}

export interface UserActionMenuMetrics {
  rect: Pick<DOMRect, 'left' | 'top' | 'bottom' | 'width'>
  pointerX: number
  pointerY: number
  viewportWidth: number
  viewportHeight: number
  menuWidth?: number
  menuHeight?: number
  padding?: number
  gap?: number
  mobileBreakpoint?: number
}

export function calculateUserActionMenuPosition(
  metrics: UserActionMenuMetrics
): UserActionMenuPosition {
  return calculateContextMenuPosition({
    rect: metrics.rect,
    pointerX: metrics.pointerX,
    pointerY: metrics.pointerY,
    viewportWidth: metrics.viewportWidth,
    viewportHeight: metrics.viewportHeight,
    panelWidth: metrics.menuWidth ?? USER_ACTION_MENU_WIDTH_FALLBACK,
    panelHeight: metrics.menuHeight ?? USER_ACTION_MENU_HEIGHT_FALLBACK,
    padding: metrics.padding ?? FLOATING_PANEL_PADDING_FALLBACK,
    gap: metrics.gap ?? FLOATING_PANEL_GAP_FALLBACK,
    mobileBreakpoint: metrics.mobileBreakpoint
  })
}

export function useUsersViewDialogs() {
  const activeMenuId = ref<number | null>(null)
  const menuPosition = ref<UserActionMenuPosition | null>(null)
  const showApiKeysModal = ref(false)
  const viewingUser = ref<AdminUser | null>(null)
  const showAllowedGroupsModal = ref(false)
  const allowedGroupsUser = ref<AdminUser | null>(null)
  const expandedGroupUserId = ref<number | null>(null)
  const showGroupReplaceModal = ref(false)
  const groupReplaceUser = ref<AdminUser | null>(null)
  const groupReplaceOldGroup = ref<{ id: number; name: string } | null>(null)
  const showBalanceModal = ref(false)
  const balanceUser = ref<AdminUser | null>(null)
  const balanceOperation = ref<'add' | 'subtract'>('add')
  const showBalanceHistoryModal = ref(false)
  const balanceHistoryUser = ref<AdminUser | null>(null)

  const closeActionMenu = () => {
    activeMenuId.value = null
    menuPosition.value = null
  }

  const openActionMenu = (user: AdminUser, event: MouseEvent) => {
    if (activeMenuId.value === user.id) {
      closeActionMenu()
      return
    }

    const target = event.currentTarget as HTMLElement | null
    if (!target) {
      closeActionMenu()
      return
    }

    const menuWidth = readThemePixelValue('--theme-user-action-menu-width', USER_ACTION_MENU_WIDTH_FALLBACK)
    const menuHeight = readThemePixelValue('--theme-user-action-menu-estimated-height', USER_ACTION_MENU_HEIGHT_FALLBACK)
    const padding = readThemePixelValue('--theme-floating-panel-viewport-padding', FLOATING_PANEL_PADDING_FALLBACK)
    const gap = readThemePixelValue('--theme-floating-panel-gap', FLOATING_PANEL_GAP_FALLBACK)

    menuPosition.value = calculateUserActionMenuPosition({
      rect: target.getBoundingClientRect(),
      pointerX: event.clientX,
      pointerY: event.clientY,
      viewportWidth: window.innerWidth,
      viewportHeight: window.innerHeight,
      menuWidth,
      menuHeight,
      padding,
      gap
    })
    activeMenuId.value = user.id
  }

  const openViewApiKeys = (user: AdminUser) => {
    viewingUser.value = user
    showApiKeysModal.value = true
  }

  const closeApiKeysModal = () => {
    showApiKeysModal.value = false
    viewingUser.value = null
  }

  const openAllowedGroups = (user: AdminUser) => {
    allowedGroupsUser.value = user
    showAllowedGroupsModal.value = true
  }

  const closeAllowedGroupsModal = () => {
    showAllowedGroupsModal.value = false
    allowedGroupsUser.value = null
  }

  const toggleExpandedGroup = (userId: number) => {
    expandedGroupUserId.value = expandedGroupUserId.value === userId ? null : userId
  }

  const closeExpandedGroup = () => {
    expandedGroupUserId.value = null
  }

  const openGroupReplace = (user: AdminUser, group: { id: number; name: string }) => {
    closeExpandedGroup()
    groupReplaceUser.value = user
    groupReplaceOldGroup.value = group
    showGroupReplaceModal.value = true
  }

  const closeGroupReplaceModal = () => {
    showGroupReplaceModal.value = false
    groupReplaceUser.value = null
    groupReplaceOldGroup.value = null
  }

  const openBalanceModal = (user: AdminUser, operation: 'add' | 'subtract') => {
    balanceUser.value = user
    balanceOperation.value = operation
    showBalanceModal.value = true
  }

  const closeBalanceModal = () => {
    showBalanceModal.value = false
    balanceUser.value = null
  }

  const openBalanceHistory = (user: AdminUser) => {
    balanceHistoryUser.value = user
    showBalanceHistoryModal.value = true
  }

  const closeBalanceHistoryModal = () => {
    showBalanceHistoryModal.value = false
    balanceHistoryUser.value = null
  }

  const reopenBalanceFromHistory = (operation: 'add' | 'subtract') => {
    if (!balanceHistoryUser.value) {
      return
    }
    openBalanceModal(balanceHistoryUser.value, operation)
  }

  return {
    activeMenuId,
    menuPosition,
    showApiKeysModal,
    viewingUser,
    showAllowedGroupsModal,
    allowedGroupsUser,
    expandedGroupUserId,
    showGroupReplaceModal,
    groupReplaceUser,
    groupReplaceOldGroup,
    showBalanceModal,
    balanceUser,
    balanceOperation,
    showBalanceHistoryModal,
    balanceHistoryUser,
    openActionMenu,
    closeActionMenu,
    openViewApiKeys,
    closeApiKeysModal,
    openAllowedGroups,
    closeAllowedGroupsModal,
    toggleExpandedGroup,
    closeExpandedGroup,
    openGroupReplace,
    closeGroupReplaceModal,
    openBalanceModal,
    closeBalanceModal,
    openBalanceHistory,
    closeBalanceHistoryModal,
    reopenBalanceFromHistory
  }
}
