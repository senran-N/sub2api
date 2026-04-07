import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import GroupClaudeCodeSection from '../groups/GroupClaudeCodeSection.vue'
import GroupInvalidRequestFallbackSection from '../groups/GroupInvalidRequestFallbackSection.vue'
import GroupMcpXmlSection from '../groups/GroupMcpXmlSection.vue'
import GroupModelRoutingSection from '../groups/GroupModelRoutingSection.vue'
import GroupSupportedScopesSection from '../groups/GroupSupportedScopesSection.vue'

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
  template:
    '<select class="select" :value="modelValue ?? \'\'" @change="$emit(\'update:modelValue\', $event.target.value === \'\' ? null : Number($event.target.value))"><option v-for="option in options" :key="String(option.value)" :value="option.value ?? \'\'">{{ option.label }}</option></select>'
}

describe('group advanced sections', () => {
  it('emits scope toggles through the supported scopes section', async () => {
    const wrapper = mount(GroupSupportedScopesSection, {
      props: {
        form: {
          platform: 'antigravity',
          supported_model_scopes: ['claude']
        } as any
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    const checkboxes = wrapper.findAll('input[type="checkbox"]')
    await checkboxes[0].trigger('change')
    await checkboxes[1].trigger('change')

    expect(wrapper.emitted('toggle-scope')).toEqual([['claude'], ['gemini_text']])
  })

  it('binds the mcp xml toggle through the shared section', async () => {
    const form = {
      platform: 'antigravity',
      mcp_xml_inject: true
    } as any

    const wrapper = mount(GroupMcpXmlSection, {
      props: {
        form
      },
      global: {
        stubs: {
          Toggle: ToggleStub,
          Icon: true
        }
      }
    })

    await wrapper.find('.toggle').setValue(false)

    expect(form.mcp_xml_inject).toBe(false)
  })

  it('binds claude code toggle and fallback group selection through the shared section', async () => {
    const form = {
      platform: 'anthropic',
      claude_code_only: false,
      fallback_group_id: null
    } as any

    const wrapper = mount(GroupClaudeCodeSection, {
      props: {
        form,
        fallbackGroupOptions: [
          { value: null, label: 'No fallback' },
          { value: 12, label: 'Fallback Group' }
        ]
      },
      global: {
        stubs: {
          Toggle: ToggleStub,
          Select: SelectStub,
          Icon: true
        }
      }
    })

    await wrapper.find('.toggle').setValue(true)
    await wrapper.find('.select').setValue('12')

    expect(form.claude_code_only).toBe(true)
    expect(form.fallback_group_id).toBe(12)
  })

  it('binds invalid request fallback selection through the shared section', async () => {
    const form = {
      platform: 'anthropic',
      subscription_type: 'standard',
      fallback_group_id_on_invalid_request: null
    } as any

    const wrapper = mount(GroupInvalidRequestFallbackSection, {
      props: {
        form,
        options: [
          { value: null, label: 'No fallback' },
          { value: 8, label: 'Fallback Group' }
        ]
      },
      global: {
        stubs: {
          Select: SelectStub
        }
      }
    })

    await wrapper.find('.select').setValue('8')

    expect(form.fallback_group_id_on_invalid_request).toBe(8)
  })

  it('binds model routing state and delegates rule actions through the shared section', async () => {
    const form = {
      platform: 'anthropic',
      model_routing_enabled: false
    } as any
    const rule = {
      pattern: '',
      accounts: [{ id: 1, name: 'Account One' }]
    }
    const searchAccountsByRule = vi.fn()
    const selectAccount = vi.fn()
    const removeSelectedAccount = vi.fn()
    const onAccountSearchFocus = vi.fn()
    const addRoutingRule = vi.fn()
    const removeRoutingRule = vi.fn()

    const wrapper = mount(GroupModelRoutingSection, {
      props: {
        form,
        rules: [rule],
        accountSearchKeyword: { createRule: '' },
        accountSearchResults: {
          createRule: [
            { id: 1, name: 'Account One' },
            { id: 2, name: 'Account Two' }
          ]
        },
        showAccountDropdown: { createRule: true },
        getRuleRenderKey: () => 'create-rule-render',
        getRuleSearchKey: () => 'createRule',
        searchAccountsByRule,
        selectAccount,
        removeSelectedAccount,
        onAccountSearchFocus,
        addRoutingRule,
        removeRoutingRule
      },
      global: {
        stubs: {
          Toggle: ToggleStub,
          Icon: true
        }
      }
    })

    await wrapper.find('.toggle').setValue(true)

    const inputs = wrapper.findAll('input[type="text"]')
    await inputs[0].setValue('claude-*')
    await inputs[1].trigger('input')
    await inputs[1].trigger('focus')

    await wrapper.get('.group-model-routing-rule-card__account-chip-remove').trigger('click')
    await wrapper.get('button[title="admin.groups.modelRouting.removeRule"]').trigger('click')
    await wrapper.findAll('button').find((button) => button.text().includes('admin.groups.modelRouting.addRule'))!.trigger('click')
    await wrapper.get('.group-model-routing-rule-card__dropdown-option:not(:disabled)').trigger('click')

    expect(form.model_routing_enabled).toBe(true)
    expect(rule.pattern).toBe('claude-*')
    expect(searchAccountsByRule).toHaveBeenCalledWith(rule)
    expect(onAccountSearchFocus).toHaveBeenCalledWith(rule)
    expect(removeSelectedAccount).toHaveBeenCalledWith(rule, 1)
    expect(removeRoutingRule).toHaveBeenCalledWith(rule)
    expect(addRoutingRule).toHaveBeenCalled()
    expect(selectAccount).toHaveBeenCalledWith(rule, { id: 2, name: 'Account Two' })
  })
})
