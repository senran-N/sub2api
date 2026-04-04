import type {
  GroupCapacitySummary,
  GroupListFilters,
  GroupSortOrderUpdate,
  GroupUsageSummary
} from '@/api/admin/groups'
import type { AdminGroup, GroupPlatform } from '@/types'

export interface GroupFiltersState {
  platform: '' | GroupPlatform
  status: '' | 'active' | 'inactive'
  is_exclusive: '' | 'true' | 'false'
}

export interface GroupPaginationState {
  page: number
  page_size: number
}

export interface GroupCapacitySnapshot {
  concurrencyUsed: number
  concurrencyMax: number
  sessionsUsed: number
  sessionsMax: number
  rpmUsed: number
  rpmMax: number
}

export function buildGroupListFilters(
  filters: GroupFiltersState,
  searchQuery: string
): GroupListFilters {
  return {
    platform: filters.platform || undefined,
    status: filters.status || undefined,
    is_exclusive: filters.is_exclusive ? filters.is_exclusive === 'true' : undefined,
    search: searchQuery.trim() || undefined
  }
}

export function applyGroupPageReset(pagination: Pick<GroupPaginationState, 'page'>): void {
  pagination.page = 1
}

export function applyGroupPageChange(
  pagination: Pick<GroupPaginationState, 'page'>,
  page: number
): void {
  pagination.page = page
}

export function applyGroupPageSizeChange(
  pagination: GroupPaginationState,
  pageSize: number
): void {
  pagination.page_size = pageSize
  pagination.page = 1
}

export function formatGroupCost(cost: number): string {
  if (cost >= 1000) {
    return cost.toFixed(0)
  }
  if (cost >= 100) {
    return cost.toFixed(1)
  }
  return cost.toFixed(2)
}

export function getGroupPlatformBadgeClass(platform: GroupPlatform): string {
  if (platform === 'anthropic') {
    return 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400'
  }
  if (platform === 'openai') {
    return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
  }
  if (platform === 'antigravity') {
    return 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400'
  }
  return 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
}

export function mapGroupUsageSummary(
  data: GroupUsageSummary[]
): Map<number, { today_cost: number; total_cost: number }> {
  const usageMap = new Map<number, { today_cost: number; total_cost: number }>()
  for (const item of data) {
    usageMap.set(item.group_id, {
      today_cost: item.today_cost,
      total_cost: item.total_cost
    })
  }
  return usageMap
}

export function mapGroupCapacitySummary(
  data: GroupCapacitySummary[]
): Map<number, GroupCapacitySnapshot> {
  const capacityMap = new Map<number, GroupCapacitySnapshot>()
  for (const item of data) {
    capacityMap.set(item.group_id, {
      concurrencyUsed: item.concurrency_used,
      concurrencyMax: item.concurrency_max,
      sessionsUsed: item.sessions_used,
      sessionsMax: item.sessions_max,
      rpmUsed: item.rpm_used,
      rpmMax: item.rpm_max
    })
  }
  return capacityMap
}

export function sortGroupsBySortOrder(groups: AdminGroup[]): AdminGroup[] {
  return [...groups].sort((left, right) => left.sort_order - right.sort_order)
}

export function buildGroupSortOrderUpdates(
  groups: Pick<AdminGroup, 'id'>[],
  step: number = 10
): GroupSortOrderUpdate[] {
  return groups.map((group, index) => ({
    id: group.id,
    sort_order: index * step
  }))
}
