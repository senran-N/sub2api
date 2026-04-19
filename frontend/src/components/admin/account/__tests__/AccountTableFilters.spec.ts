import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import AccountTableFilters from '../AccountTableFilters.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

describe('AccountTableFilters', () => {
  it('renders Grok-aware platform and type options', () => {
    const wrapper = mount(AccountTableFilters, {
      props: {
        searchQuery: '',
        filters: {
          platform: '',
          type: '',
          status: '',
          group: '',
          privacy_mode: ''
        },
        groups: []
      },
      global: {
        stubs: {
          SearchInput: {
            template: '<div />'
          },
          Select: {
            props: ['modelValue', 'options'],
            emits: ['update:modelValue', 'change'],
            template: '<div class="select-stub" :data-options="JSON.stringify(options)" />'
          }
        }
      }
    })

    const selects = wrapper.findAll('.select-stub')
    const platformOptions = JSON.parse(selects[0].attributes('data-options'))
    const typeOptions = JSON.parse(selects[1].attributes('data-options'))

    expect(platformOptions).toEqual([
      { value: '', label: 'admin.accounts.allPlatforms' },
      { value: 'anthropic', label: 'admin.accounts.platforms.anthropic' },
      { value: 'openai', label: 'admin.accounts.platforms.openai' },
      { value: 'gemini', label: 'admin.accounts.platforms.gemini' },
      { value: 'grok', label: 'admin.accounts.platforms.grok' },
      { value: 'antigravity', label: 'admin.accounts.platforms.antigravity' }
    ])

    expect(typeOptions).toEqual([
      { value: '', label: 'admin.accounts.allTypes' },
      { value: 'oauth', label: 'admin.accounts.oauthType' },
      { value: 'setup-token', label: 'admin.accounts.setupToken' },
      { value: 'apikey', label: 'admin.accounts.apiKey' },
      { value: 'upstream', label: 'admin.accounts.types.upstream' },
      { value: 'session', label: 'admin.accounts.types.session' },
      { value: 'bedrock', label: 'admin.accounts.bedrockLabel' }
    ])
  })

  it('renders privacy mode options and emits privacy_mode updates', async () => {
    const wrapper = mount(AccountTableFilters, {
      props: {
        searchQuery: '',
        filters: {
          platform: '',
          type: '',
          status: '',
          group: '',
          privacy_mode: ''
        },
        groups: []
      },
      global: {
        stubs: {
          SearchInput: {
            template: '<div />'
          },
          Select: {
            props: ['modelValue', 'options'],
            emits: ['update:modelValue', 'change'],
            template: '<div class="select-stub" :data-options="JSON.stringify(options)" />'
          }
        }
      }
    })

    const selects = wrapper.findAll('.select-stub')
    expect(selects).toHaveLength(5)

    const privacyOptions = JSON.parse(selects[3].attributes('data-options'))
    expect(privacyOptions).toEqual([
      { value: '', label: 'admin.accounts.allPrivacyModes' },
      { value: '__unset__', label: 'admin.accounts.privacyUnset' },
      { value: 'training_off', label: 'Privacy' },
      { value: 'training_set_cf_blocked', label: 'CF' },
      { value: 'training_set_failed', label: 'Fail' }
    ])
  })
})
