import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import UserUsageCostTooltip from '../usage/UserUsageCostTooltip.vue'
import UserUsageFiltersBar from '../usage/UserUsageFiltersBar.vue'
import UserUsageStatsCards from '../usage/UserUsageStatsCards.vue'
import UserUsageTokenCell from '../usage/UserUsageTokenCell.vue'
import UserUsageTokenTooltip from '../usage/UserUsageTokenTooltip.vue'

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

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
      class="date-range-stub"
      @click="
        $emit('update:startDate', '2026-04-01');
        $emit('update:endDate', '2026-04-05');
        $emit('change', { startDate: '2026-04-01', endDate: '2026-04-05', preset: null })
      "
    >
      {{ startDate }}-{{ endDate }}
    </button>
  `
}

describe('user usage local components', () => {
  it('renders stats cards', () => {
    const wrapper = mount(UserUsageStatsCards, {
      props: {
        stats: {
          total_requests: 12,
          total_tokens: 2500,
          total_input_tokens: 1000,
          total_output_tokens: 1500,
          total_cost: 1.2345,
          total_actual_cost: 0.9876,
          average_duration_ms: 250
        }
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    expect(wrapper.text()).toContain('12')
    expect(wrapper.text()).toContain('2.50K')
    expect(wrapper.text()).toContain('$0.9876')
    expect(wrapper.text()).toContain('250ms')
  })

  it('renders filters bar and emits filter actions', async () => {
    const wrapper = mount(UserUsageFiltersBar, {
      props: {
        apiKeyId: undefined,
        apiKeyOptions: [
          { value: null, label: 'All' },
          { value: 3, label: 'Demo' }
        ],
        startDate: '2026-04-04',
        endDate: '2026-04-05',
        loading: false,
        exporting: false
      },
      global: {
        stubs: {
          Select: SelectStub,
          DateRangePicker: DateRangePickerStub
        }
      }
    })

    await wrapper.find('.select-stub').trigger('click')
    await wrapper.find('.date-range-stub').trigger('click')
    const buttons = wrapper.findAll('button')
    await buttons[2].trigger('click')
    await buttons[3].trigger('click')
    await buttons[4].trigger('click')

    expect(wrapper.emitted('update:apiKeyId')?.[0]).toEqual([3])
    expect(wrapper.emitted('update:startDate')?.[0]).toEqual(['2026-04-01'])
    expect(wrapper.emitted('update:endDate')?.[0]).toEqual(['2026-04-05'])
    expect(wrapper.emitted('date-range-change')?.[0]).toEqual([
      { startDate: '2026-04-01', endDate: '2026-04-05', preset: null }
    ])
    expect(wrapper.emitted('apply-filters')?.length).toBe(2)
    expect(wrapper.emitted('reset')?.length).toBe(1)
    expect(wrapper.emitted('export')?.length).toBe(1)
  })

  it('renders token cell and emits tooltip events', async () => {
    const row = {
      input_tokens: 1200,
      output_tokens: 3400,
      cache_read_tokens: 1500,
      cache_creation_tokens: 800,
      cache_creation_5m_tokens: 800,
      cache_creation_1h_tokens: 0,
      cache_ttl_overridden: true,
      image_count: 0,
      image_size: null
    } as any

    const wrapper = mount(UserUsageTokenCell, {
      props: { row },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    expect(wrapper.text()).toContain('1,200')
    expect(wrapper.text()).toContain('3,400')
    expect(wrapper.text()).toContain('1.5K')
    expect(wrapper.text()).toContain('800')

    await wrapper.find('button').trigger('mouseenter')
    await wrapper.find('button').trigger('mouseleave')

    expect(wrapper.emitted('show-details')?.[0]?.[1]).toStrictEqual(row)
    expect(wrapper.emitted('hide-details')?.length).toBe(1)
  })

  it('renders token and cost tooltips', () => {
    const log = {
      input_tokens: 4057,
      output_tokens: 101,
      cache_creation_tokens: 500,
      cache_read_tokens: 278272,
      cache_creation_5m_tokens: 200,
      cache_creation_1h_tokens: 300,
      cache_ttl_overridden: true,
      input_cost: 0.020285,
      output_cost: 0.00303,
      cache_creation_cost: 0.0005,
      cache_read_cost: 0.069568,
      total_cost: 0.092883,
      actual_cost: 0.092883,
      rate_multiplier: 1,
      service_tier: 'priority'
    } as any

    const tokenTooltip = mount(UserUsageTokenTooltip, {
      props: {
        visible: true,
        position: { x: 120, y: 80 },
        log
      },
      global: {
        stubs: {
          Teleport: true
        }
      }
    })

    expect(tokenTooltip.text()).toContain('admin.usage.inputTokens')
    expect(tokenTooltip.text()).toContain('4,057')
    expect(tokenTooltip.text()).toContain('admin.usage.cacheCreation5mTokens')
    expect(tokenTooltip.text()).toContain('R-1h')
    expect(tokenTooltip.text()).toContain('usage.cacheTtlOverridden1h')
    expect(tokenTooltip.text()).toContain('282,930')

    const costTooltip = mount(UserUsageCostTooltip, {
      props: {
        visible: true,
        position: { x: 120, y: 80 },
        log
      },
      global: {
        stubs: {
          Teleport: true
        }
      }
    })

    expect(costTooltip.text()).toContain('usage.costDetails')
    expect(costTooltip.text()).toContain('$0.020285')
    expect(costTooltip.text()).toContain('$5.0000')
    expect(costTooltip.text()).toContain('$30.0000')
    expect(costTooltip.text()).toContain('usage.serviceTierPriority')
  })
})
