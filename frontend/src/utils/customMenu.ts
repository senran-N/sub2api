import type { CustomMenuItem } from '@/types'

export function resolveCustomPageMenuItem(
  menuItemId: string,
  publicItems: CustomMenuItem[],
  adminItems: CustomMenuItem[],
  isAdmin: boolean
): CustomMenuItem | null {
  const publicItem = publicItems.find((item) => item.id === menuItemId) ?? null
  if (publicItem) {
    return publicItem
  }

  if (!isAdmin) {
    return null
  }

  return adminItems.find((item) => item.id === menuItemId) ?? null
}
