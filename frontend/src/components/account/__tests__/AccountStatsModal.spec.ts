import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { ref } from 'vue'
import AccountStatsModal from '../AccountStatsModal.vue'
import type { Account, AccountUsageStatsResponse } from '@/types'

const { getStats } = vi.hoisted(() => ({
  getStats: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      getStats
    }
  }
}))

vi.mock('@/composables/useDocumentThemeVersion', () => ({
  useDocumentThemeVersion: () => ref(0)
}))

vi.mock('vue-chartjs', () => ({
  Line: {
    name: 'Line',
    template: '<div class="line-chart-stub" />'
  }
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

function makeAccount(overrides: Partial<Account> = {}): Account {
  return {
    id: 1,
    name: 'Account Alpha',
    platform: 'openai',
    type: 'oauth',
    proxy_id: null,
    concurrency: 1,
    priority: 1,
    status: 'active',
    error_message: null,
    last_used_at: null,
    expires_at: null,
    auto_pause_on_expired: true,
    created_at: '2026-04-01T00:00:00Z',
    updated_at: '2026-04-01T00:00:00Z',
    schedulable: true,
    rate_limited_at: null,
    rate_limit_reset_at: null,
    overload_until: null,
    temp_unschedulable_until: null,
    temp_unschedulable_reason: null,
    session_window_start: null,
    session_window_end: null,
    session_window_status: null,
    ...overrides
  }
}

function makeStatsResponse(): AccountUsageStatsResponse {
  return {
    history: [
      {
        date: '2026-04-07',
        label: '04/07',
        requests: 12,
        tokens: 1200,
        cost: 0.8,
        actual_cost: 1.2,
        user_cost: 1.5
      }
    ],
    summary: {
      days: 30,
      actual_days_used: 1,
      total_cost: 1.2,
      total_user_cost: 1.5,
      total_standard_cost: 0.8,
      total_requests: 12,
      total_tokens: 1200,
      avg_daily_cost: 1.2,
      avg_daily_user_cost: 1.5,
      avg_daily_requests: 12,
      avg_daily_tokens: 1200,
      avg_duration_ms: 320,
      today: {
        date: '2026-04-07',
        cost: 1.2,
        user_cost: 1.5,
        requests: 12,
        tokens: 1200
      },
      highest_cost_day: {
        date: '2026-04-07',
        label: '04/07',
        cost: 1.2,
        user_cost: 1.5,
        requests: 12
      },
      highest_request_day: {
        date: '2026-04-07',
        label: '04/07',
        requests: 12,
        cost: 1.2,
        user_cost: 1.5
      }
    },
    models: [],
    endpoints: [],
    upstream_endpoints: []
  }
}

describe('AccountStatsModal', () => {
  beforeEach(() => {
    getStats.mockReset()
  })

  it('首次打开时会立即拉取并渲染账号统计数据', async () => {
    getStats.mockResolvedValue(makeStatsResponse())

    const wrapper = mount(AccountStatsModal, {
      props: {
        show: true,
        account: makeAccount({ id: 42, name: 'Stats Account' })
      },
      global: {
        stubs: {
          BaseDialog: {
            props: ['show', 'title', 'width'],
            template: '<div class="base-dialog"><slot /><slot name="footer" /></div>'
          },
          LoadingSpinner: true,
          ModelDistributionChart: true,
          EndpointDistributionChart: true,
          Icon: true
        }
      }
    })

    await flushPromises()

    expect(getStats).toHaveBeenCalledWith(42, 30)
    expect(wrapper.text()).toContain('Stats Account')
    expect(wrapper.text()).toContain('12')
    expect(wrapper.find('.line-chart-stub').exists()).toBe(true)
  })
})
