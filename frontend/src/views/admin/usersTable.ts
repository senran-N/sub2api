import type { UserListFilters } from '@/api/admin/users'
import type { SelectOption } from '@/components/common/Select.vue'
import type { Column } from '@/components/common/types'
import type { AdminGroup, AdminUser, UserAttributeDefinition } from '@/types'

export type UserRoleFilter = '' | 'admin' | 'user'
export type UserStatusFilter = '' | 'active' | 'disabled'
export type BuiltInUserFilterKey = 'role' | 'status' | 'group'

export interface UsersFilterState {
  role: UserRoleFilter
  status: UserStatusFilter
  group: string
}

export interface UsersPaginationState {
  page: number
  page_size: number
  pages?: number
}

export interface UserColumnLabels {
  user: string
  id: string
  username: string
  notes: string
  role: string
  groups: string
  subscriptions: string
  balance: string
  usage: string
  concurrency: string
  status: string
  created: string
  actions: string
}

export interface UserGroupsSummary {
  exclusive: AdminGroup[]
  publicGroups: AdminGroup[]
}

export interface UserColumnToggleEffects {
  refreshSecondaryData: boolean
  reloadUsers: boolean
  loadGroups: boolean
}

export const DEFAULT_USER_HIDDEN_COLUMNS = [
  'notes',
  'groups',
  'subscriptions',
  'usage',
  'concurrency'
] as const

export function createDefaultUsersFilters(): UsersFilterState {
  return {
    role: '',
    status: '',
    group: ''
  }
}

export function buildUserAttributeColumns(
  definitions: UserAttributeDefinition[]
): Column[] {
  return definitions
    .filter((definition) => definition.enabled)
    .map((definition) => ({
      key: `attr_${definition.id}`,
      label: definition.name,
      sortable: false
    }))
}

export function buildUserTableColumns(
  attributeColumns: Column[],
  labels: UserColumnLabels
): Column[] {
  return [
    { key: 'email', label: labels.user, sortable: true },
    { key: 'id', label: labels.id, sortable: true },
    { key: 'username', label: labels.username, sortable: true },
    { key: 'notes', label: labels.notes, sortable: false },
    ...attributeColumns,
    { key: 'role', label: labels.role, sortable: true },
    { key: 'groups', label: labels.groups, sortable: false },
    { key: 'subscriptions', label: labels.subscriptions, sortable: false },
    { key: 'balance', label: labels.balance, sortable: true },
    { key: 'usage', label: labels.usage, sortable: false },
    { key: 'concurrency', label: labels.concurrency, sortable: true },
    { key: 'status', label: labels.status, sortable: true },
    { key: 'created_at', label: labels.created, sortable: true },
    { key: 'actions', label: labels.actions, sortable: false }
  ]
}

export function filterVisibleUserColumns(
  allColumns: Column[],
  hiddenColumns: Set<string>
): Column[] {
  return allColumns.filter(
    (column) => column.key === 'email' || column.key === 'actions' || !hiddenColumns.has(column.key)
  )
}

export function buildUserGroupFilterOptions(
  groups: AdminGroup[],
  allGroupsLabel: string
): SelectOption[] {
  return [
    { value: '', label: allGroupsLabel },
    ...groups
      .filter(
        (group) =>
          group.status === 'active' &&
          group.is_exclusive &&
          group.subscription_type === 'standard'
      )
      .map((group) => ({
        value: group.name,
        label: group.name
      }))
  ]
}

export function getUserGroupsSummary(
  groups: AdminGroup[],
  user: Pick<AdminUser, 'allowed_groups'>
): UserGroupsSummary {
  const exclusive: AdminGroup[] = []
  const publicGroups: AdminGroup[] = []

  for (const group of groups) {
    if (group.status !== 'active' || group.subscription_type !== 'standard') {
      continue
    }

    if (group.is_exclusive) {
      if (user.allowed_groups?.includes(group.id)) {
        exclusive.push(group)
      }
      continue
    }

    publicGroups.push(group)
  }

  return { exclusive, publicGroups }
}

export function getAttributeDefinitionName(
  definitions: UserAttributeDefinition[],
  attributeId: number
): string {
  return definitions.find((definition) => definition.id === attributeId)?.name || String(attributeId)
}

export function formatUserAttributeValue(
  userAttributeValues: Record<number, Record<number, string>>,
  definitions: UserAttributeDefinition[],
  userId: number,
  attributeId: number
): string {
  const value = userAttributeValues[userId]?.[attributeId]
  if (!value) {
    return '-'
  }

  const definition = definitions.find((item) => item.id === attributeId)
  if (!definition) {
    return value
  }

  if (definition.type === 'multi_select') {
    try {
      const parsed = JSON.parse(value)
      if (Array.isArray(parsed)) {
        return parsed
          .map((entry) => definition.options?.find((option) => option.value === entry)?.label || entry)
          .join(', ')
      }
    } catch {
      return value
    }
  }

  if (definition.type === 'select' && definition.options) {
    return definition.options.find((option) => option.value === value)?.label || value
  }

  return value
}

export function buildUserListFilters(
  filters: UsersFilterState,
  searchQuery: string,
  activeAttributeFilters: Record<number, string>,
  includeSubscriptions: boolean
): UserListFilters {
  const normalizedAttributes: Record<number, string> = {}
  for (const [attributeId, value] of Object.entries(activeAttributeFilters)) {
    const normalizedValue = value.trim()
    if (normalizedValue) {
      normalizedAttributes[Number(attributeId)] = normalizedValue
    }
  }

  return {
    role: filters.role || undefined,
    status: filters.status || undefined,
    search: searchQuery.trim() || undefined,
    group_name: filters.group || undefined,
    attributes: Object.keys(normalizedAttributes).length > 0 ? normalizedAttributes : undefined,
    include_subscriptions: includeSubscriptions
  }
}

export function applyUsersPageChange(
  pagination: UsersPaginationState,
  page: number
): void {
  const pageLimit = pagination.pages || 1
  pagination.page = Math.max(1, Math.min(page, pageLimit))
}

export function applyUsersPageSizeChange(
  pagination: UsersPaginationState,
  pageSize: number
): void {
  pagination.page_size = pageSize
  pagination.page = 1
}

export function toggleBuiltInUserFilter(
  visibleFilters: Set<string>,
  filters: UsersFilterState,
  key: BuiltInUserFilterKey
): { shouldLoadGroups: boolean } {
  if (visibleFilters.has(key)) {
    visibleFilters.delete(key)
    if (key === 'role') {
      filters.role = ''
    }
    if (key === 'status') {
      filters.status = ''
    }
    if (key === 'group') {
      filters.group = ''
    }
    return { shouldLoadGroups: false }
  }

  visibleFilters.add(key)
  return { shouldLoadGroups: key === 'group' }
}

export function toggleUserAttributeFilter(
  visibleFilters: Set<string>,
  activeAttributeFilters: Record<number, string>,
  attributeId: number
): void {
  const key = `attr_${attributeId}`
  if (visibleFilters.has(key)) {
    visibleFilters.delete(key)
    delete activeAttributeFilters[attributeId]
    return
  }

  visibleFilters.add(key)
  activeAttributeFilters[attributeId] = ''
}

export function getUserColumnToggleEffects(
  key: string,
  wasHidden: boolean
): UserColumnToggleEffects {
  return {
    refreshSecondaryData: wasHidden && (key === 'usage' || key.startsWith('attr_')),
    reloadUsers: key === 'subscriptions',
    loadGroups: wasHidden && key === 'groups'
  }
}

export function getUserSubscriptionDaysRemaining(
  expiresAt: string,
  now: Date = new Date()
): number {
  return Math.ceil((new Date(expiresAt).getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
}
