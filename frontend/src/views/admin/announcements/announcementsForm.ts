import { formatDateTimeLocalInput, parseDateTimeLocalInput } from '@/utils/format'
import type {
  AdminGroup,
  Announcement,
  AnnouncementNotifyMode,
  AnnouncementStatus,
  AnnouncementTargeting,
  CreateAnnouncementRequest,
  UpdateAnnouncementRequest
} from '@/types'

export type AnnouncementStatusFilter = '' | AnnouncementStatus

export interface AnnouncementFiltersState {
  status: AnnouncementStatusFilter
}

export interface AnnouncementFormState {
  title: string
  content: string
  status: AnnouncementStatus
  notify_mode: AnnouncementNotifyMode
  starts_at_str: string
  ends_at_str: string
  targeting: AnnouncementTargeting
}

interface TranslateFn {
  (key: string, params?: Record<string, unknown>): string
}

export function createDefaultAnnouncementFilters(): AnnouncementFiltersState {
  return {
    status: ''
  }
}

export function createDefaultAnnouncementForm(): AnnouncementFormState {
  return {
    title: '',
    content: '',
    status: 'draft',
    notify_mode: 'silent',
    starts_at_str: '',
    ends_at_str: '',
    targeting: { any_of: [] }
  }
}

export function resetAnnouncementForm(form: AnnouncementFormState): void {
  Object.assign(form, createDefaultAnnouncementForm())
}

export function fillAnnouncementForm(
  form: AnnouncementFormState,
  announcement: Announcement
): void {
  form.title = announcement.title
  form.content = announcement.content
  form.status = announcement.status
  form.notify_mode = announcement.notify_mode || 'silent'
  form.starts_at_str = announcement.starts_at
    ? formatDateTimeLocalInput(Math.floor(new Date(announcement.starts_at).getTime() / 1000))
    : ''
  form.ends_at_str = announcement.ends_at
    ? formatDateTimeLocalInput(Math.floor(new Date(announcement.ends_at).getTime() / 1000))
    : ''
  form.targeting = announcement.targeting ?? { any_of: [] }
}

export function buildAnnouncementListFilters(
  filters: AnnouncementFiltersState,
  searchQuery: string
): {
  status?: AnnouncementStatus
  search?: string
} {
  return {
    status: filters.status || undefined,
    search: searchQuery || undefined
  }
}

export function getAnnouncementStatusLabel(
  status: AnnouncementStatus | string,
  t: TranslateFn
): string {
  if (status === 'draft') return t('admin.announcements.statusLabels.draft')
  if (status === 'active') return t('admin.announcements.statusLabels.active')
  if (status === 'archived') return t('admin.announcements.statusLabels.archived')
  return status
}

export function getAnnouncementTargetingSummary(
  targeting: AnnouncementTargeting,
  t: TranslateFn
): string {
  const anyOf = targeting?.any_of ?? []
  if (anyOf.length === 0) {
    return t('admin.announcements.targetingSummaryAll')
  }
  return t('admin.announcements.targetingSummaryCustom', { groups: anyOf.length })
}

export function filterAnnouncementSubscriptionGroups(groups: AdminGroup[]): AdminGroup[] {
  return groups.filter((group) => group.subscription_type === 'subscription')
}

export function validateAnnouncementTargeting(targeting: AnnouncementTargeting): boolean {
  const anyOf = targeting?.any_of ?? []
  if (anyOf.length > 50) {
    return false
  }

  return anyOf.every((group) => (group?.all_of ?? []).length <= 50)
}

export function buildCreateAnnouncementRequest(
  form: AnnouncementFormState
): CreateAnnouncementRequest {
  const startsAt = parseDateTimeLocalInput(form.starts_at_str)
  const endsAt = parseDateTimeLocalInput(form.ends_at_str)

  return {
    title: form.title,
    content: form.content,
    status: form.status,
    notify_mode: form.notify_mode,
    targeting: form.targeting,
    starts_at: startsAt ?? undefined,
    ends_at: endsAt ?? undefined
  }
}

export function buildUpdateAnnouncementRequest(
  form: AnnouncementFormState,
  original: Announcement
): UpdateAnnouncementRequest {
  const payload: UpdateAnnouncementRequest = {}

  if (form.title !== original.title) payload.title = form.title
  if (form.content !== original.content) payload.content = form.content
  if (form.status !== original.status) payload.status = form.status
  if (form.notify_mode !== (original.notify_mode || 'silent')) {
    payload.notify_mode = form.notify_mode
  }

  const originalStarts = original.starts_at
    ? Math.floor(new Date(original.starts_at).getTime() / 1000)
    : null
  const originalEnds = original.ends_at ? Math.floor(new Date(original.ends_at).getTime() / 1000) : null
  const newStarts = parseDateTimeLocalInput(form.starts_at_str)
  const newEnds = parseDateTimeLocalInput(form.ends_at_str)

  if (newStarts !== originalStarts) {
    payload.starts_at = newStarts === null ? 0 : newStarts
  }
  if (newEnds !== originalEnds) {
    payload.ends_at = newEnds === null ? 0 : newEnds
  }
  if (JSON.stringify(form.targeting ?? {}) !== JSON.stringify(original.targeting ?? {})) {
    payload.targeting = form.targeting
  }

  return payload
}
