import { describe, expect, it } from 'vitest'
import type { AdminGroup, Announcement } from '@/types'
import {
  buildAnnouncementListFilters,
  buildCreateAnnouncementRequest,
  buildUpdateAnnouncementRequest,
  createDefaultAnnouncementFilters,
  createDefaultAnnouncementForm,
  fillAnnouncementForm,
  filterAnnouncementSubscriptionGroups,
  getAnnouncementStatusLabel,
  getAnnouncementTargetingSummary,
  resetAnnouncementForm,
  validateAnnouncementTargeting
} from '../announcements/announcementsForm'

function createGroup(overrides: Partial<AdminGroup> = {}): AdminGroup {
  return {
    id: 1,
    name: 'Pro',
    description: 'subscription plan',
    platform: 'openai',
    rate_multiplier: 1.5,
    status: 'active',
    subscription_type: 'subscription',
    user_count: 0,
    account_count: 0,
    created_at: '2026-04-04T00:00:00Z',
    updated_at: '2026-04-04T00:00:00Z',
    ...overrides
  } as AdminGroup
}

function createAnnouncement(overrides: Partial<Announcement> = {}): Announcement {
  return {
    id: 1,
    title: 'Maintenance',
    content: 'Window',
    status: 'active',
    notify_mode: 'popup',
    targeting: { any_of: [{ all_of: [{ type: 'subscription', operator: 'in', group_ids: [1] }] }] },
    starts_at: '2026-04-04T08:30:00Z',
    ends_at: '2026-04-05T09:45:00Z',
    created_at: '2026-04-04T00:00:00Z',
    updated_at: '2026-04-04T00:00:00Z',
    ...overrides
  }
}

describe('announcementsForm helpers', () => {
  it('creates and resets filter/form state', () => {
    const filters = createDefaultAnnouncementFilters()
    filters.status = 'active'
    expect(buildAnnouncementListFilters(filters, 'maintenance')).toEqual({
      status: 'active',
      search: 'maintenance'
    })

    const form = createDefaultAnnouncementForm()
    form.title = 'Changed'
    resetAnnouncementForm(form)
    expect(form).toEqual(createDefaultAnnouncementForm())
  })

  it('fills form state and builds create/update payloads', () => {
    const form = createDefaultAnnouncementForm()
    const announcement = createAnnouncement()

    fillAnnouncementForm(form, announcement)
    expect(form.title).toBe('Maintenance')
    expect(form.notify_mode).toBe('popup')
    expect(form.starts_at_str).toBeTruthy()
    expect(form.ends_at_str).toBeTruthy()

    form.title = 'Planned Maintenance'
    form.ends_at_str = ''

    expect(buildCreateAnnouncementRequest(form)).toEqual({
      title: 'Planned Maintenance',
      content: 'Window',
      status: 'active',
      notify_mode: 'popup',
      targeting: announcement.targeting,
      starts_at: expect.any(Number),
      ends_at: undefined
    })

    expect(buildUpdateAnnouncementRequest(form, announcement)).toEqual({
      title: 'Planned Maintenance',
      ends_at: 0
    })
  })

  it('summarizes targeting, labels status, validates groups, and filters subscription groups', () => {
    const t = (key: string, params?: Record<string, unknown>) =>
      params ? `${key}:${params.groups}` : key

    expect(getAnnouncementStatusLabel('draft', t)).toBe('admin.announcements.statusLabels.draft')
    expect(getAnnouncementStatusLabel('active', t)).toBe('admin.announcements.statusLabels.active')
    expect(getAnnouncementTargetingSummary({ any_of: [] }, t)).toBe(
      'admin.announcements.targetingSummaryAll'
    )
    expect(
      getAnnouncementTargetingSummary(
        { any_of: [{ all_of: [] }, { all_of: [] }] },
        t
      )
    ).toBe('admin.announcements.targetingSummaryCustom:2')

    expect(
      validateAnnouncementTargeting({
        any_of: Array.from({ length: 51 }, () => ({ all_of: [] }))
      })
    ).toBe(false)
    expect(
      validateAnnouncementTargeting({
        any_of: [{ all_of: Array.from({ length: 51 }, () => ({ type: 'subscription', operator: 'in' as const })) }]
      })
    ).toBe(false)
    expect(validateAnnouncementTargeting({ any_of: [{ all_of: [] }] })).toBe(true)

    expect(
      filterAnnouncementSubscriptionGroups([
        createGroup(),
        createGroup({ id: 2, subscription_type: 'standard', name: 'Standard' }),
        createGroup({ id: 3, name: 'Anthropic', platform: 'anthropic' })
      ])
    ).toEqual([
      createGroup(),
      createGroup({ id: 3, name: 'Anthropic', platform: 'anthropic' })
    ])
  })
})
