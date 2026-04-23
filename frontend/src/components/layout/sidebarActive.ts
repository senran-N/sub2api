const EXACT_MATCH_SIDEBAR_PATHS = new Set([
  '/admin/orders'
])

export function isSidebarItemActive(currentPath: string, itemPath: string): boolean {
  if (currentPath === itemPath) {
    return true
  }

  if (EXACT_MATCH_SIDEBAR_PATHS.has(itemPath)) {
    return false
  }

  return currentPath.startsWith(itemPath + '/')
}
