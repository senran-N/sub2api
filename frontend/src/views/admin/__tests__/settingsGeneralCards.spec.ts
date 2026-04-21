import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import SettingsCustomMenuCard from '../settings/SettingsCustomMenuCard.vue'
import SettingsDefaultsCard from '../settings/SettingsDefaultsCard.vue'
import SettingsPurchaseCard from '../settings/SettingsPurchaseCard.vue'
import SettingsSiteCard from '../settings/SettingsSiteCard.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
  })
}))

const SelectStub = {
  props: ['id', 'name', 'modelValue', 'options', 'placeholder', 'ariaLabel', 'ariaLabelledby'],
  emits: ['update:modelValue'],
  template:
    '<select class="select" :id="id" :name="name" :aria-label="ariaLabel" :aria-labelledby="ariaLabelledby" :value="modelValue" @change="$emit(\'update:modelValue\', Number($event.target.value))"><option v-for="option in options" :key="option.value" :value="option.value">{{ option.label }}</option></select>'
}

const ToggleStub = {
  props: ['modelValue', 'disabled'],
  emits: ['update:modelValue'],
  template:
    '<input class="toggle" type="checkbox" :checked="modelValue" @change="$emit(\'update:modelValue\', $event.target.checked)" />'
}

const ImageUploadStub = {
  props: ['modelValue'],
  emits: ['update:modelValue'],
  template:
    '<button class="image-upload" type="button" @click="$emit(\'update:modelValue\', \'next-value\')">upload</button>'
}

function createForm(overrides: Record<string, unknown> = {}) {
  return {
    default_balance: 0,
    default_concurrency: 1,
    default_subscriptions: [],
    backend_mode_enabled: false,
    site_name: 'Sub2API',
    site_subtitle: 'Subtitle',
    frontend_theme: 'factory',
    api_base_url: '',
    custom_endpoints: [],
    contact_info: '',
    doc_url: '',
    site_logo: '',
    home_content: '',
    hide_ccs_import_button: false,
    purchase_subscription_enabled: false,
    purchase_subscription_url: '',
    custom_menu_items: [],
    ...overrides
  }
}

describe('settings general cards', () => {
  it('wires default subscriptions actions through the extracted card', async () => {
    const form = createForm({
      default_subscriptions: [{ group_id: 10, validity_days: 30 }]
    }) as any

    const wrapper = mount(SettingsDefaultsCard, {
      props: {
        form,
        defaultSubscriptionGroupOptions: [
          {
            value: 10,
            label: 'Starter',
            description: 'Starter plan',
            platform: 'openai',
            subscriptionType: 'subscription',
            rate: 1
          }
        ],
        toDefaultSubscriptionGroupOption: (option: any) => option
      },
      global: {
        stubs: {
          Select: SelectStub,
          GroupBadge: true,
          GroupOptionItem: true
        }
      }
    })

    await wrapper.find('button.btn-secondary').trigger('click')
    await wrapper.find('.select').setValue('10')
    await wrapper.findAll('input[type="number"]').at(-1)?.setValue('45')
    await wrapper.findAll('button').at(-1)?.trigger('click')

    expect(wrapper.emitted('add-default-subscription')).toHaveLength(1)
    expect(form.default_subscriptions[0].group_id).toBe(10)
    expect(form.default_subscriptions[0].validity_days).toBe(45)
    expect(wrapper.emitted('remove-default-subscription')?.[0]).toEqual([0])
  })

  it('wires site endpoint actions and image upload through the extracted card', async () => {
    const form = createForm({
      custom_endpoints: [{ name: 'Docs', endpoint: 'https://docs.example.com', description: 'Docs' }]
    }) as any

    const wrapper = mount(SettingsSiteCard, {
      props: {
        form
      },
      global: {
        stubs: {
          Select: SelectStub,
          Toggle: ToggleStub,
          ImageUpload: ImageUploadStub
        }
      }
    })

    expect(wrapper.text()).toContain('admin.settings.site.backendMode')
    expect(wrapper.text()).toContain('admin.settings.site.frontendTheme')
    expect(wrapper.text()).toContain('admin.settings.site.frontendThemeHint')
    expect(wrapper.text()).not.toContain('Frontend Theme')
    expect(wrapper.findAll('.toggle').map((node) => node.attributes('aria-label'))).toEqual([
      'admin.settings.site.backendMode',
      'admin.settings.site.hideCcsImportButton'
    ])

    await wrapper.find('.toggle').setValue(true)
    await wrapper.get('.settings-site-card__remove-button').trigger('click')
    await wrapper.get('.image-upload').trigger('click')
    await wrapper.get('.settings-site-card__add-button').trigger('click')

    expect(form.backend_mode_enabled).toBe(true)
    expect(form.site_logo).toBe('next-value')
    expect(wrapper.emitted('remove-endpoint')?.[0]).toEqual([0])
    expect(wrapper.emitted('add-endpoint')).toHaveLength(1)
  })

  it('keeps purchase settings bound and exposes integration docs through the extracted card', async () => {
    const form = createForm() as any

    const wrapper = mount(SettingsPurchaseCard, {
      props: {
        form
      },
      global: {
        stubs: {
          Toggle: ToggleStub
        }
      }
    })

    expect(wrapper.get('.toggle').attributes('aria-label')).toBe(
      'admin.settings.purchase.enabled'
    )

    await wrapper.find('.toggle').setValue(true)
    await wrapper.find('input[type="url"]').setValue('https://billing.example.com')

    expect(form.purchase_subscription_enabled).toBe(true)
    expect(form.purchase_subscription_url).toBe('https://billing.example.com')
    expect(wrapper.get('a').attributes('href')).toContain('ADMIN_PAYMENT_INTEGRATION_API.md')
  })

  it('keeps custom menu item actions wired through the extracted card', async () => {
    const menuForm = createForm({
      custom_menu_items: [
        {
          id: 'docs',
          label: 'Docs',
          icon_svg: '',
          url: 'https://docs.example.com',
          visibility: 'user',
          sort_order: 0
        },
        {
          id: 'admin',
          label: 'Admin',
          icon_svg: '',
          url: 'https://admin.example.com',
          visibility: 'admin',
          sort_order: 1
        }
      ]
    }) as any

    const menuWrapper = mount(SettingsCustomMenuCard, {
      props: {
        form: menuForm
      },
      global: {
        stubs: {
          ImageUpload: ImageUploadStub
        }
      }
    })

    await menuWrapper.find('input[type="text"]').setValue('Documentation')
    await menuWrapper.find('select').setValue('admin')
    await menuWrapper.find('input[type="url"]').setValue('https://docs.example.com/v2')
    await menuWrapper.find('.image-upload').trigger('click')
    await menuWrapper.get('button[title="admin.settings.customMenu.moveDown"]').trigger('click')
    await menuWrapper.get('button[title="admin.settings.customMenu.moveUp"]').trigger('click')
    await menuWrapper.get('button[title="admin.settings.customMenu.remove"]').trigger('click')
    await menuWrapper.get('button.w-full').trigger('click')

    expect(menuForm.custom_menu_items[0].label).toBe('Documentation')
    expect(menuForm.custom_menu_items[0].visibility).toBe('admin')
    expect(menuForm.custom_menu_items[0].url).toBe('https://docs.example.com/v2')
    expect(menuForm.custom_menu_items[0].icon_svg).toBe('next-value')
    expect(menuWrapper.emitted('move-item')).toEqual([[0, 1], [1, -1]])
    expect(menuWrapper.emitted('remove-item')?.[0]).toEqual([0])
    expect(menuWrapper.emitted('add-item')).toHaveLength(1)
  })
})
