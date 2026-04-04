import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import SettingsClaudeCodeCard from '../settings/SettingsClaudeCodeCard.vue'
import SettingsGatewayForwardingCard from '../settings/SettingsGatewayForwardingCard.vue'
import SettingsSchedulingCard from '../settings/SettingsSchedulingCard.vue'

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

function createForm(overrides: Record<string, unknown> = {}) {
  return {
    min_claude_code_version: '',
    max_claude_code_version: '',
    allow_ungrouped_key_scheduling: false,
    enable_fingerprint_unification: true,
    enable_metadata_passthrough: false,
    ...overrides
  }
}

describe('settings gateway cards', () => {
  it('keeps claude code version inputs bound through the extracted card', async () => {
    const form = createForm() as any

    const wrapper = mount(SettingsClaudeCodeCard, {
      props: {
        form
      }
    })

    const inputs = wrapper.findAll('input')
    await inputs[0].setValue('1.2.3')
    await inputs[1].setValue('2.0.0')

    expect(form.min_claude_code_version).toBe('1.2.3')
    expect(form.max_claude_code_version).toBe('2.0.0')
  })

  it('keeps scheduling toggle bound through the extracted card', async () => {
    const form = createForm() as any

    const wrapper = mount(SettingsSchedulingCard, {
      props: {
        form
      },
      global: {
        stubs: {
          Toggle: ToggleStub
        }
      }
    })

    await wrapper.find('.toggle').setValue(true)

    expect(form.allow_ungrouped_key_scheduling).toBe(true)
  })

  it('keeps gateway forwarding toggles bound through the extracted card', async () => {
    const form = createForm() as any

    const wrapper = mount(SettingsGatewayForwardingCard, {
      props: {
        form
      },
      global: {
        stubs: {
          Toggle: ToggleStub
        }
      }
    })

    const toggles = wrapper.findAll('.toggle')
    await toggles[0].setValue(false)
    await toggles[1].setValue(true)

    expect(form.enable_fingerprint_unification).toBe(false)
    expect(form.enable_metadata_passthrough).toBe(true)
  })
})
