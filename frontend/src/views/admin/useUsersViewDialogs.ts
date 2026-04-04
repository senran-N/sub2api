import { ref } from 'vue'
import type { AdminUser } from '@/types'

export interface UserActionMenuPosition {
  top: number
  left: number
}

export interface UserActionMenuMetrics {
  rect: Pick<DOMRect, 'left' | 'top' | 'bottom' | 'width'>
  pointerX: number
  pointerY: number
  viewportWidth: number
  viewportHeight: number
  menuWidth?: number
  menuHeight?: number
  padding?: number
  mobileBreakpoint?: number
}

export function calculateUserActionMenuPosition(
  metrics: UserActionMenuMetrics
): UserActionMenuPosition {
  const menuWidth = metrics.menuWidth ?? 200
  const menuHeight = metrics.menuHeight ?? 240
  const padding = metrics.padding ?? 8
  const mobileBreakpoint = metrics.mobileBreakpoint ?? 768

  if (metrics.viewportWidth < mobileBreakpoint) {
    const left = Math.max(
      padding,
      Math.min(
        metrics.rect.left + metrics.rect.width / 2 - menuWidth / 2,
        metrics.viewportWidth - menuWidth - padding
      )
    )

    let top = metrics.rect.bottom + 4
    if (top + menuHeight > metrics.viewportHeight - padding) {
      top = metrics.rect.top - menuHeight - 4
      if (top < padding) {
        top = padding
      }
    }

    return { top, left }
  }

  const left = Math.max(
    padding,
    Math.min(metrics.pointerX - menuWidth, metrics.viewportWidth - menuWidth - padding)
  )
  let top = metrics.pointerY
  if (top + menuHeight > metrics.viewportHeight - padding) {
    top = metrics.viewportHeight - menuHeight - padding
  }

  return { top, left }
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

    menuPosition.value = calculateUserActionMenuPosition({
      rect: target.getBoundingClientRect(),
      pointerX: event.clientX,
      pointerY: event.clientY,
      viewportWidth: window.innerWidth,
      viewportHeight: window.innerHeight
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
