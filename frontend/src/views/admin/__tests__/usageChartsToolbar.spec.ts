import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import UsageChartsToolbar from '../usage/UsageChartsToolbar.vue'

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

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

const SelectStub = {
  props: ['modelValue', 'options'],
  emits: ['update:modelValue', 'change'],
  template: `
    <button
      class="granularity-stub"
      @click="
        $emit('update:modelValue', options[1]?.value ?? modelValue);
        $emit('change', options[1]?.value ?? modelValue)
      "
    >
      {{ modelValue }}
    </button>
  `
}

describe('usage charts toolbar', () => {
  it('emits date and granularity updates', async () => {
    const wrapper = mount(UsageChartsToolbar, {
      props: {
        startDate: '2026-04-04',
        endDate: '2026-04-05',
        granularity: 'hour',
        granularityOptions: [
          { value: 'hour', label: 'Hour' },
          { value: 'day', label: 'Day' }
        ]
      },
      global: {
        stubs: {
          DateRangePicker: DateRangePickerStub,
          Select: SelectStub
        }
      }
    })

    await wrapper.find('.date-range-stub').trigger('click')
    await wrapper.find('.granularity-stub').trigger('click')

    expect(wrapper.emitted('update:startDate')?.[0]).toEqual(['2026-04-01'])
    expect(wrapper.emitted('update:endDate')?.[0]).toEqual(['2026-04-05'])
    expect(wrapper.emitted('date-range-change')?.[0]).toEqual([
      { startDate: '2026-04-01', endDate: '2026-04-05', preset: null }
    ])
    expect(wrapper.emitted('update:granularity')?.[0]).toEqual(['day'])
    expect(wrapper.emitted('granularity-change')?.length).toBe(1)
  })
})
