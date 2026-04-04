import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import GroupActionToolbar from '../groups/GroupActionToolbar.vue'
import GroupDialogFooter from '../groups/GroupDialogFooter.vue'
import GroupEditStatusField from '../groups/GroupEditStatusField.vue'
import GroupFilterFields from '../groups/GroupFilterFields.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) =>
      params ? `${key}:${JSON.stringify(params)}` : key
  })
}))

describe('group toolbar components', () => {
  it('renders filter fields and emits search/filter updates', async () => {
    const wrapper = mount(GroupFilterFields, {
      props: {
        searchQuery: 'anthropic',
        platform: '',
        status: '',
        isExclusive: '',
        platformOptions: [
          { value: '', label: 'All' },
          { value: 'openai', label: 'OpenAI' }
        ],
        statusOptions: [
          { value: '', label: 'All' },
          { value: 'active', label: 'Active' }
        ],
        exclusiveOptions: [
          { value: '', label: 'All' },
          { value: 'true', label: 'Exclusive' }
        ]
      },
      global: {
        stubs: {
          Icon: true,
          Select: {
            props: ['modelValue', 'options'],
            template:
              '<button class="select-stub" @click="$emit(\'update:modelValue\', options[1].value); $emit(\'change\')">{{ modelValue }}</button>'
          }
        }
      }
    })

    await wrapper.find('input').setValue('openai')
    expect(wrapper.emitted('update:searchQuery')?.[0]).toEqual(['openai'])
    expect(wrapper.emitted('search-input')?.length).toBe(1)

    const buttons = wrapper.findAll('.select-stub')
    await buttons[0].trigger('click')
    await buttons[1].trigger('click')
    await buttons[2].trigger('click')

    expect(wrapper.emitted('update:platform')?.[0]).toEqual(['openai'])
    expect(wrapper.emitted('platform-change')?.length).toBe(1)
    expect(wrapper.emitted('update:status')?.[0]).toEqual(['active'])
    expect(wrapper.emitted('status-change')?.length).toBe(1)
    expect(wrapper.emitted('update:isExclusive')?.[0]).toEqual(['true'])
    expect(wrapper.emitted('exclusive-change')?.length).toBe(1)
  })

  it('renders action toolbar and emits toolbar actions', async () => {
    const wrapper = mount(GroupActionToolbar, {
      props: {
        loading: false
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    const buttons = wrapper.findAll('button')
    await buttons[0].trigger('click')
    await buttons[1].trigger('click')
    await buttons[2].trigger('click')

    expect(wrapper.emitted('refresh')?.length).toBe(1)
    expect(wrapper.emitted('sort-order')?.length).toBe(1)
    expect(wrapper.emitted('create')?.length).toBe(1)
  })

  it('renders dialog footer and edit status field bindings', async () => {
    const footer = mount(GroupDialogFooter, {
      props: {
        formId: 'edit-group-form',
        submitting: true,
        submittingText: 'saving',
        submitText: 'save'
      }
    })

    expect(footer.text()).toContain('saving')
    await footer.find('button').trigger('click')
    expect(footer.emitted('close')?.length).toBe(1)

    const statusField = mount(GroupEditStatusField, {
      props: {
        modelValue: 'inactive',
        statusOptions: [
          { value: 'active', label: 'Active' },
          { value: 'inactive', label: 'Inactive' }
        ]
      },
      global: {
        stubs: {
          Select: {
            props: ['modelValue', 'options'],
            template:
              '<button class="status-select" @click="$emit(\'update:modelValue\', options[0].value)">{{ modelValue }}</button>'
          }
        }
      }
    })

    expect(statusField.text()).toContain('admin.groups.form.status')
    await statusField.find('.status-select').trigger('click')
    expect(statusField.emitted('update:modelValue')?.[0]).toEqual(['active'])
  })
})
