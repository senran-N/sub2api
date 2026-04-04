import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import SettingsLinuxdoCard from '../settings/SettingsLinuxdoCard.vue'
import SettingsRegistrationCard from '../settings/SettingsRegistrationCard.vue'
import SettingsTurnstileCard from '../settings/SettingsTurnstileCard.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
  })
}))

const ToggleStub = {
  props: ['modelValue', 'disabled'],
  emits: ['update:modelValue'],
  template:
    '<input class="toggle" type="checkbox" :checked="modelValue" :disabled="disabled" @change="$emit(\'update:modelValue\', $event.target.checked)" />'
}

function createForm(overrides: Record<string, unknown> = {}) {
  return {
    registration_enabled: true,
    email_verify_enabled: false,
    promo_code_enabled: true,
    invitation_code_enabled: false,
    password_reset_enabled: false,
    frontend_url: '',
    totp_enabled: false,
    totp_encryption_key_configured: false,
    turnstile_enabled: false,
    turnstile_site_key: '',
    turnstile_secret_key: '',
    turnstile_secret_key_configured: true,
    linuxdo_connect_enabled: false,
    linuxdo_connect_client_id: '',
    linuxdo_connect_client_secret: '',
    linuxdo_connect_client_secret_configured: true,
    linuxdo_connect_redirect_url: '',
    ...overrides
  }
}

describe('settings security cards', () => {
  it('keeps registration conditionals and draft events wired', async () => {
    const form = createForm({
      email_verify_enabled: true,
      password_reset_enabled: true
    }) as any

    const wrapper = mount(SettingsRegistrationCard, {
      props: {
        form,
        tags: ['example.com'],
        draft: ' Foo.Bar '
      },
      global: {
        stubs: {
          Toggle: ToggleStub,
          Icon: true
        }
      }
    })

    expect(wrapper.text()).toContain('admin.settings.registration.frontendUrl')
    expect(wrapper.text()).toContain('admin.settings.registration.totpKeyNotConfigured')

    const textInput = wrapper.find('input[type="text"]')
    await textInput.setValue('New.Domain')
    await textInput.trigger('keydown', { key: 'Enter' })
    await textInput.trigger('paste')
    await textInput.trigger('blur')
    await wrapper.find('button').trigger('click')

    expect(wrapper.emitted('update:draft')?.[0]).toEqual(['New.Domain'])
    expect(wrapper.emitted('draft-input')).toHaveLength(1)
    expect(wrapper.emitted('draft-keydown')).toHaveLength(1)
    expect(wrapper.emitted('draft-paste')).toHaveLength(1)
    expect(wrapper.emitted('commit-draft')).toHaveLength(1)
    expect(wrapper.emitted('remove-tag')?.[0]).toEqual(['example.com'])
  })

  it('shows turnstile fields only when enabled', async () => {
    const form = createForm() as any

    const wrapper = mount(SettingsTurnstileCard, {
      props: { form },
      global: {
        stubs: {
          Toggle: ToggleStub
        }
      }
    })

    expect(wrapper.text()).not.toContain('admin.settings.turnstile.siteKey')

    await wrapper.find('.toggle').setValue(true)

    expect(wrapper.text()).toContain('admin.settings.turnstile.siteKey')
    expect(form.turnstile_enabled).toBe(true)
  })

  it('emits linuxdo quick set action and shows suggestion when enabled', async () => {
    const wrapper = mount(SettingsLinuxdoCard, {
      props: {
        form: createForm({
          linuxdo_connect_enabled: true
        }) as any,
        redirectUrlSuggestion: 'https://sub2api.example.com/callback'
      },
      global: {
        stubs: {
          Toggle: ToggleStub
        }
      }
    })

    expect(wrapper.text()).toContain('https://sub2api.example.com/callback')

    await wrapper.find('button').trigger('click')

    expect(wrapper.emitted('quick-set-copy')).toHaveLength(1)
  })
})
