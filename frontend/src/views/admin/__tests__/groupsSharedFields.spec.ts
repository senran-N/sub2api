import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import GroupBaseFieldsSection from '../groups/GroupBaseFieldsSection.vue'
import GroupCopyAccountsField from '../groups/GroupCopyAccountsField.vue'
import GroupExclusiveSection from '../groups/GroupExclusiveSection.vue'
import GroupSubscriptionSection from '../groups/GroupSubscriptionSection.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
  })
}))

const SelectStub = {
  props: ['modelValue', 'options', 'disabled'],
  emits: ['update:modelValue'],
  template:
    '<select class="select" :value="modelValue" :disabled="disabled" @change="$emit(\'update:modelValue\', $event.target.value)"><option v-for="option in options" :key="String(option.value)" :value="option.value">{{ option.label }}</option></select>'
}

const ToggleStub = {
  props: ['modelValue'],
  emits: ['update:modelValue'],
  template:
    '<input class="toggle" type="checkbox" :checked="modelValue" @change="$emit(\'update:modelValue\', $event.target.checked)" />'
}

describe('group shared fields', () => {
  it('binds base group fields and resets copied accounts on platform change', async () => {
    const form = {
      name: '',
      description: '',
      platform: 'anthropic',
      rate_multiplier: 1,
      copy_accounts_from_group_ids: [1]
    } as any

    const wrapper = mount(GroupBaseFieldsSection, {
      props: {
        form,
        platformOptions: [
          { value: 'anthropic', label: 'Anthropic' },
          { value: 'openai', label: 'OpenAI' }
        ],
        copyAccountsOptions: [
          { value: 1, label: 'Alpha (3 个账号)' },
          { value: 2, label: 'Beta (2 个账号)' }
        ],
        copyAccountsTooltipText: 'copy tooltip',
        copyAccountsHintText: 'copy hint',
        platformHint: 'platform hint',
        namePlaceholder: 'enter name',
        descriptionPlaceholder: 'optional description',
        rateMultiplierHint: 'rate hint',
        resetCopyAccountsOnPlatformChange: true,
        nameTourTarget: 'group-form-name',
        platformTourTarget: 'group-form-platform',
        rateMultiplierTourTarget: 'group-form-multiplier'
      },
      global: {
        stubs: {
          Select: SelectStub,
          Icon: true
        }
      }
    })

    const textInputs = wrapper.findAll('input')
    await textInputs[0].setValue('New Group')
    await wrapper.get('textarea').setValue('Description')
    await wrapper.find('.select').setValue('openai')
    await textInputs[1].setValue('1.5')

    expect(form.name).toBe('New Group')
    expect(form.description).toBe('Description')
    expect(form.platform).toBe('openai')
    expect(form.rate_multiplier).toBe(1.5)
    expect(form.copy_accounts_from_group_ids).toEqual([])
    expect(wrapper.text()).toContain('platform hint')
    expect(wrapper.text()).toContain('rate hint')
    expect(textInputs[0].attributes('data-tour')).toBe('group-form-name')
    expect(textInputs[1].attributes('data-tour')).toBe('group-form-multiplier')
  })

  it('renders selected copy-account groups and emits add/remove actions', async () => {
    const wrapper = mount(GroupCopyAccountsField, {
      props: {
        options: [
          { value: 1, label: 'Alpha (3 个账号)' },
          { value: 2, label: 'Beta (2 个账号)' }
        ],
        selectedGroupIds: [1],
        tooltipText: 'copy tooltip',
        hintText: 'copy hint'
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    expect(wrapper.text()).toContain('Alpha (3 个账号)')
    expect(wrapper.text()).toContain('copy tooltip')
    expect(wrapper.text()).toContain('copy hint')

    await wrapper.get('select').setValue('2')
    await wrapper.get('button').trigger('click')

    expect(wrapper.emitted('add-group')?.[0]).toEqual([2])
    expect(wrapper.emitted('remove-group')?.[0]).toEqual([1])
  })

  it('binds subscription type and limits through the shared section', async () => {
    const form = {
      subscription_type: 'standard',
      daily_limit_usd: null,
      weekly_limit_usd: null,
      monthly_limit_usd: null
    } as any

    const wrapper = mount(GroupSubscriptionSection, {
      props: {
        form,
        subscriptionTypeOptions: [
          { value: 'standard', label: 'Standard' },
          { value: 'subscription', label: 'Subscription' }
        ],
        subscriptionTypeHint: 'subscription hint',
        subscriptionTypeDisabled: false
      },
      global: {
        stubs: {
          Select: SelectStub
        }
      }
    })

    expect(wrapper.text()).toContain('subscription hint')

    await wrapper.get('.select').setValue('subscription')
    const inputs = wrapper.findAll('input')
    await inputs[0].setValue('10')
    await inputs[1].setValue('20')
    await inputs[2].setValue('30')

    expect(form.subscription_type).toBe('subscription')
    expect(form.daily_limit_usd).toBe(10)
    expect(form.weekly_limit_usd).toBe(20)
    expect(form.monthly_limit_usd).toBe(30)
  })

  it('binds exclusive state through the shared section', async () => {
    const form = {
      subscription_type: 'standard',
      is_exclusive: false
    } as any

    const wrapper = mount(GroupExclusiveSection, {
      props: {
        form,
        tourTarget: 'group-form-exclusive'
      },
      global: {
        stubs: {
          Toggle: ToggleStub,
          Icon: true
        }
      }
    })

    expect(wrapper.attributes('data-tour')).toBe('group-form-exclusive')

    await wrapper.find('.toggle').setValue(true)

    expect(form.is_exclusive).toBe(true)
    expect(wrapper.text()).toContain('admin.groups.exclusive')
  })
})
