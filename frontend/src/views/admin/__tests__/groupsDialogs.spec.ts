import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import GroupCreateDialog from '../groups/GroupCreateDialog.vue'
import GroupEditDialog from '../groups/GroupEditDialog.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) =>
      params ? `${key}:${JSON.stringify(params)}` : key
  })
}))

const BaseDialogStub = {
  props: ['show', 'title', 'width'],
  emits: ['close'],
  template:
    '<div><button class="dialog-close" @click="$emit(\'close\')">close</button><slot /><slot name="footer" /></div>'
}

const GroupFormSectionsStub = {
  emits: ['add-copy-group', 'remove-copy-group', 'toggle-scope'],
  template:
    '<div><button class="add-group" @click="$emit(\'add-copy-group\', 7)">add</button><button class="remove-group" @click="$emit(\'remove-copy-group\', 7)">remove</button><button class="toggle-scope" @click="$emit(\'toggle-scope\', \'claude\')">scope</button></div>'
}

const GroupDialogFooterStub = {
  props: ['formId', 'submitting', 'submittingText', 'submitText'],
  emits: ['close'],
  template:
    '<div><button class="footer-close" @click="$emit(\'close\')">close</button><button class="footer-submit">{{ submitText }}</button></div>'
}

function createCommonProps() {
  return {
    submitting: false,
    platformOptions: [{ value: 'anthropic', label: 'Anthropic' }],
    copyAccountsGroupOptions: [{ value: 1, label: 'Group A' }],
    subscriptionTypeOptions: [{ value: 'standard', label: 'Standard' }],
    fallbackGroupOptions: [{ value: null, label: 'None' }],
    invalidRequestFallbackOptions: [{ value: null, label: 'None' }],
    rules: [],
    accountSearchKeyword: {},
    accountSearchResults: {},
    showAccountDropdown: {},
    getRuleRenderKey: () => 'rule',
    getRuleSearchKey: () => 'rule',
    searchAccountsByRule: vi.fn(),
    selectAccount: vi.fn(),
    removeSelectedAccount: vi.fn(),
    onAccountSearchFocus: vi.fn(),
    addRoutingRule: vi.fn(),
    removeRoutingRule: vi.fn()
  }
}

function createGlobalStubs() {
  return {
    BaseDialog: BaseDialogStub,
    GroupFormSections: GroupFormSectionsStub,
    GroupDialogFooter: GroupDialogFooterStub,
    GroupEditStatusField: true
  }
}

describe('group dialogs', () => {
  it('renders create dialog shell and re-emits create interactions', async () => {
    const wrapper = mount(GroupCreateDialog, {
      props: {
        ...createCommonProps(),
        show: true,
        form: {
          name: '',
          platform: 'anthropic',
          supported_model_scopes: []
        } as any
      },
      global: {
        stubs: createGlobalStubs()
      }
    })

    await wrapper.find('.add-group').trigger('click')
    await wrapper.find('.remove-group').trigger('click')
    await wrapper.find('.toggle-scope').trigger('click')
    await wrapper.find('form').trigger('submit')
    await wrapper.find('.dialog-close').trigger('click')
    await wrapper.find('.footer-close').trigger('click')

    expect(wrapper.emitted('add-copy-group')?.[0]).toEqual([7])
    expect(wrapper.emitted('remove-copy-group')?.[0]).toEqual([7])
    expect(wrapper.emitted('toggle-scope')?.[0]).toEqual(['claude'])
    expect(wrapper.emitted('submit')?.length).toBe(1)
    expect(wrapper.emitted('close')?.length).toBe(2)
  })

  it('renders edit dialog shell only when editing group exists and re-emits events', async () => {
    const wrapper = mount(GroupEditDialog, {
      props: {
        ...createCommonProps(),
        show: true,
        editingGroup: { id: 9, name: 'Alpha' } as any,
        editStatusOptions: [{ value: 'active', label: 'Active' }],
        form: {
          name: 'Alpha',
          status: 'active',
          platform: 'anthropic',
          supported_model_scopes: []
        } as any
      },
      global: {
        stubs: createGlobalStubs()
      }
    })

    expect(wrapper.find('form').exists()).toBe(true)
    await wrapper.find('.add-group').trigger('click')
    await wrapper.find('.toggle-scope').trigger('click')
    await wrapper.find('form').trigger('submit')

    expect(wrapper.emitted('add-copy-group')?.[0]).toEqual([7])
    expect(wrapper.emitted('toggle-scope')?.[0]).toEqual(['claude'])
    expect(wrapper.emitted('submit')?.length).toBe(1)
  })
})
