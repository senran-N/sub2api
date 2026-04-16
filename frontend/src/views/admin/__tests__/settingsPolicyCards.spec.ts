import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import SettingsAdminApiKeyCard from '../settings/SettingsAdminApiKeyCard.vue'
import SettingsBetaPolicyCard from '../settings/SettingsBetaPolicyCard.vue'
import SettingsStreamTimeoutCard from '../settings/SettingsStreamTimeoutCard.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
  })
}))

const ToggleStub = {
  props: ['modelValue'],
  emits: ['update:modelValue'],
  template:
    '<input class="toggle" type="checkbox" :checked="modelValue" @change="$emit(\'update:modelValue\', $event.target.checked)" />'
}

const SelectStub = {
  props: ['modelValue', 'options'],
  emits: ['update:modelValue'],
  template: `
    <select
      class="select"
      :value="modelValue"
      @change="$emit('update:modelValue', $event.target.value)"
    >
      <option
        v-for="option in options"
        :key="option.value"
        :value="option.value"
      >
        {{ option.label }}
      </option>
    </select>
  `
}

describe('settings policy cards', () => {
  it('emits admin api key actions and shows generated key', async () => {
    const wrapper = mount(SettingsAdminApiKeyCard, {
      props: {
        loading: false,
        exists: true,
        maskedKey: 'abcdefghij...wxyz',
        operating: false,
        newKey: 'new-secret-key'
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    await wrapper.find('button.btn-secondary').trigger('click')
    await wrapper.find('button.btn-primary').trigger('click')

    expect(wrapper.text()).toContain('new-secret-key')
    expect(wrapper.emitted('regenerate')).toHaveLength(1)
    expect(wrapper.emitted('copy')).toHaveLength(1)
  })

  it('updates stream timeout form through the extracted card', async () => {
    const form = {
      enabled: true,
      action: 'temp_unsched',
      temp_unsched_minutes: 5,
      threshold_count: 3,
      threshold_window_minutes: 10
    } as const

    const wrapper = mount(SettingsStreamTimeoutCard, {
      props: {
        loading: false,
        saving: false,
        form: { ...form }
      },
      global: {
        stubs: {
          Toggle: ToggleStub
        }
      }
    })

    expect(wrapper.text()).toContain('admin.settings.streamTimeout.tempUnschedMinutes')

    await wrapper.find('select').setValue('error')
    await wrapper.find('button.btn-primary').trigger('click')

    expect(wrapper.props('form').action).toBe('error')
    expect(wrapper.text()).not.toContain('admin.settings.streamTimeout.tempUnschedMinutes')
    expect(wrapper.emitted('save')).toHaveLength(1)
  })

  it('renders block error message field for beta policy rules', async () => {
    const rules = [
      {
        beta_token: 'fast-mode-2026-02-01',
        action: 'block',
        scope: 'all',
        error_message: 'blocked',
        model_whitelist: ['claude-opus-*'],
        fallback_action: 'filter'
      }
    ]

    const wrapper = mount(SettingsBetaPolicyCard, {
      props: {
        loading: false,
        saving: false,
        rules,
        actionOptions: [
          { value: 'pass', label: 'Pass' },
          { value: 'block', label: 'Block' }
        ],
        scopeOptions: [{ value: 'all', label: 'All' }],
        getDisplayName: (token: string) => token.toUpperCase()
      },
      global: {
        stubs: {
          Select: SelectStub
        }
      }
    })

    expect(wrapper.text()).toContain('FAST-MODE-2026-02-01')
    expect(wrapper.find('input').element.value).toBe('blocked')
    await wrapper.find('textarea').setValue('claude-opus-*\nclaude-opus-4-1')
    expect(rules[0].model_whitelist).toEqual(['claude-opus-*', 'claude-opus-4-1'])

    await wrapper.find('button.btn-primary').trigger('click')

    expect(wrapper.emitted('save')).toHaveLength(1)
  })
})
