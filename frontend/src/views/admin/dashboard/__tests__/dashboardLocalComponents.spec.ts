import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import type { DashboardStats } from '@/types'
import DashboardChartControls from '../DashboardChartControls.vue'
import DashboardLoadingSkeleton from '../DashboardLoadingSkeleton.vue'
import DashboardStatsCards from '../DashboardStatsCards.vue'
import DashboardUserTrendCard from '../DashboardUserTrendCard.vue'

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

vi.mock('vue-chartjs', () => ({
  Line: {
    template: '<div class="line-stub">line</div>'
  }
}))

function createDashboardStats(): DashboardStats {
  return {
    total_users: 100,
    today_new_users: 5,
    active_users: 20,
    hourly_active_users: 10,
    stats_updated_at: '',
    stats_stale: false,
    total_api_keys: 30,
    active_api_keys: 25,
    total_accounts: 12,
    normal_accounts: 10,
    error_accounts: 2,
    ratelimit_accounts: 0,
    overload_accounts: 0,
    total_requests: 999,
    total_input_tokens: 0,
    total_output_tokens: 0,
    total_cache_creation_tokens: 0,
    total_cache_read_tokens: 0,
    total_tokens: 2000,
    total_cost: 30,
    total_actual_cost: 20,
    today_requests: 88,
    today_input_tokens: 0,
    today_output_tokens: 0,
    today_cache_creation_tokens: 0,
    today_cache_read_tokens: 0,
    today_tokens: 777,
    today_cost: 3,
    today_actual_cost: 2,
    average_duration_ms: 456,
    uptime: 0,
    rpm: 66,
    tpm: 777
  }
}

const SelectStub = {
  props: ['modelValue', 'options'],
  emits: ['update:modelValue', 'change'],
  template: `
    <button
      class="select-stub"
      @click="
        $emit('update:modelValue', options[1]?.value ?? modelValue);
        $emit('change', options[1]?.value ?? modelValue)
      "
    >
      {{ modelValue }}
    </button>
  `
}

const DateRangePickerStub = {
  props: ['startDate', 'endDate'],
  emits: ['update:startDate', 'update:endDate', 'change'],
  template: `
    <button
      class="range-stub"
      @click="
        $emit('update:startDate', '2026-04-01');
        $emit('update:endDate', '2026-04-04');
        $emit('change', { startDate: '2026-04-01', endDate: '2026-04-04', preset: null })
      "
    >
      range
    </button>
  `
}

describe('dashboard local components', () => {
  it('renders loading skeleton cards', () => {
    const wrapper = mount(DashboardLoadingSkeleton)
    expect(wrapper.findAll('.card')).toHaveLength(9)
  })

  it('renders stats cards content through formatter props', () => {
    const wrapper = mount(DashboardStatsCards, {
      props: {
        stats: createDashboardStats(),
        formatTokens: (value: number | undefined | null) => `tokens:${value ?? 0}`,
        formatNumber: (value: number) => `num:${value}`,
        formatCost: (value: number) => `cost:${value}`,
        formatDuration: (value: number) => `dur:${value}`
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    expect(wrapper.text()).toContain('30')
    expect(wrapper.text()).toContain('num:999')
    expect(wrapper.text()).toContain('tokens:777')
    expect(wrapper.text()).toContain('cost:20')
    expect(wrapper.text()).toContain('dur:456')
  })

  it('renders chart controls and emits filter changes', async () => {
    const wrapper = mount(DashboardChartControls, {
      props: {
        startDate: '2026-04-03',
        endDate: '2026-04-04',
        granularity: 'hour',
        granularityOptions: [
          { value: 'hour', label: 'Hour' },
          { value: 'day', label: 'Day' }
        ],
        loading: false
      },
      global: {
        stubs: {
          DateRangePicker: DateRangePickerStub,
          Select: SelectStub
        }
      }
    })

    const buttons = wrapper.findAll('button')
    await buttons[0].trigger('click')
    await buttons[1].trigger('click')
    await buttons[2].trigger('click')

    expect(wrapper.emitted('update:startDate')?.[0]).toEqual(['2026-04-01'])
    expect(wrapper.emitted('update:endDate')?.[0]).toEqual(['2026-04-04'])
    expect(wrapper.emitted('date-range-change')?.[0]).toEqual([
      { startDate: '2026-04-01', endDate: '2026-04-04', preset: null }
    ])
    expect(wrapper.emitted('refresh')?.length).toBe(1)
    expect(wrapper.emitted('update:granularity')?.[0]).toEqual(['day'])
    expect(wrapper.emitted('granularity-change')?.length).toBe(1)
  })

  it('renders user trend card states and actions', async () => {
    const withData = mount(DashboardUserTrendCard, {
      props: {
        loading: false,
        chartData: {
          labels: ['2026-04-04'],
          datasets: [
            {
              label: 'alice',
              data: [100],
              borderColor: '#000',
              backgroundColor: '#111',
              fill: false,
              tension: 0.3
            }
          ]
        },
        chartOptions: {}
      },
      global: {
        stubs: {
          LoadingSpinner: true
        }
      }
    })
    expect(withData.find('.line-stub').exists()).toBe(true)

    const loading = mount(DashboardUserTrendCard, {
      props: {
        loading: true,
        chartData: null,
        chartOptions: {}
      },
      global: {
        stubs: {
          LoadingSpinner: { template: '<div class="spinner-stub">spin</div>' },
        }
      }
    })
    expect(loading.find('.spinner-stub').exists()).toBe(true)

    const empty = mount(DashboardUserTrendCard, {
      props: {
        loading: false,
        chartData: null,
        chartOptions: {}
      },
      global: {
        stubs: {
          LoadingSpinner: true
        }
      }
    })
    expect(empty.text()).toContain('admin.dashboard.noDataAvailable')
  })
})
