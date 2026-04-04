export interface ProxyListFiltersState {
  protocol: string
  status: string
}

export interface ProxyPaginationState {
  page: number
  page_size: number
  total?: number
  pages?: number
}

export function buildProxyListFilters(
  filters: ProxyListFiltersState,
  searchQuery: string
): {
  protocol?: string
  status?: 'active' | 'inactive'
  search?: string
} {
  return {
    protocol: filters.protocol || undefined,
    status: (filters.status || undefined) as 'active' | 'inactive' | undefined,
    search: searchQuery.trim() || undefined
  }
}

export function applyProxyPageChange(pagination: ProxyPaginationState, page: number) {
  pagination.page = page
}

export function applyProxyPageSizeChange(pagination: ProxyPaginationState, pageSize: number) {
  pagination.page_size = pageSize
  pagination.page = 1
}

export function resetProxyListPage(pagination: ProxyPaginationState) {
  pagination.page = 1
}
