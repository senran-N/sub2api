import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import GroupAccountFilterSection from '../groups/GroupAccountFilterSection.vue'
import GroupOpenAIMessagesSection from '../groups/GroupOpenAIMessagesSection.vue'

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

describe('group platform sections', () => {
  it('binds openai messages dispatch and default model through the shared section', async () => {
    const form = {
      platform: 'openai',
      allow_messages_dispatch: false,
      default_mapped_model: ''
    } as any

    const wrapper = mount(GroupOpenAIMessagesSection, {
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
    await wrapper.get('input[type="text"]').setValue('gpt-5.4')

    expect(form.allow_messages_dispatch).toBe(true)
    expect(form.default_mapped_model).toBe('gpt-5.4')
  })

  it('binds oauth and privacy account filters through the shared section', async () => {
    const form = {
      platform: 'anthropic',
      require_oauth_only: false,
      require_privacy_set: false
    } as any

    const wrapper = mount(GroupAccountFilterSection, {
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
    await toggles[0].setValue(true)
    await toggles[1].setValue(true)

    expect(form.require_oauth_only).toBe(true)
    expect(form.require_privacy_set).toBe(true)
  })
})
