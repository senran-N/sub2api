import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import SettingsEmailDisabledCard from '../settings/SettingsEmailDisabledCard.vue'
import SettingsSmtpCard from '../settings/SettingsSmtpCard.vue'
import SettingsTestEmailCard from '../settings/SettingsTestEmailCard.vue'

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
    email_verify_enabled: true,
    smtp_host: '',
    smtp_port: 587,
    smtp_username: '',
    smtp_password: '',
    smtp_password_configured: true,
    smtp_from_email: '',
    smtp_from_name: '',
    smtp_use_tls: true,
    ...overrides
  }
}

describe('settings email cards', () => {
  it('shows the disabled email hint in the extracted card', () => {
    const wrapper = mount(SettingsEmailDisabledCard, {
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    expect(wrapper.text()).toContain('admin.settings.emailTabDisabledTitle')
    expect(wrapper.text()).toContain('admin.settings.emailTabDisabledHint')
  })

  it('keeps smtp field bindings and actions wired through the extracted card', async () => {
    const form = createForm() as any
    const wrapper = mount(SettingsSmtpCard, {
      props: {
        form,
        testing: false,
        disabled: false
      },
      global: {
        stubs: {
          Toggle: ToggleStub
        }
      }
    })

    const inputs = wrapper.findAll('input')
    await inputs[0].setValue('smtp.example.com')
    await inputs[1].setValue('465')
    await inputs[2].setValue('mailer')
    await inputs[3].setValue('secret')
    await inputs[3].trigger('keydown')
    await inputs[3].trigger('paste')
    await inputs[4].setValue('noreply@example.com')
    await inputs[5].setValue('Sub2API Mailer')
    await wrapper.find('.toggle').setValue(false)
    await wrapper.get('button').trigger('click')

    expect(form.smtp_host).toBe('smtp.example.com')
    expect(form.smtp_port).toBe(465)
    expect(form.smtp_username).toBe('mailer')
    expect(form.smtp_password).toBe('secret')
    expect(form.smtp_from_email).toBe('noreply@example.com')
    expect(form.smtp_from_name).toBe('Sub2API Mailer')
    expect(form.smtp_use_tls).toBe(false)
    expect(wrapper.emitted('password-interaction')).toHaveLength(2)
    expect(wrapper.emitted('test-connection')).toHaveLength(1)
  })

  it('keeps test email v-model and send action wired through the extracted card', async () => {
    const wrapper = mount(SettingsTestEmailCard, {
      props: {
        modelValue: '',
        sending: false,
        disabled: false
      }
    })

    await wrapper.find('input[type="email"]').setValue('ops@example.com')

    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual(['ops@example.com'])

    await wrapper.setProps({
      modelValue: 'ops@example.com'
    })
    await wrapper.get('button').trigger('click')

    expect(wrapper.emitted('send')).toHaveLength(1)
  })
})
