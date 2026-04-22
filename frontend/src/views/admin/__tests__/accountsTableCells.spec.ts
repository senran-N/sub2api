import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import AccountCapacityCell from '@/components/account/AccountCapacityCell.vue'
import type { Account } from '@/types'
import AccountActionsCell from '../accounts/AccountActionsCell.vue'
import AccountExpiresCell from '../accounts/AccountExpiresCell.vue'
import AccountLastUsedCell from '../accounts/AccountLastUsedCell.vue'
import AccountNameCell from '../accounts/AccountNameCell.vue'
import AccountNotesCell from '../accounts/AccountNotesCell.vue'
import AccountPlatformTypeCell from '../accounts/AccountPlatformTypeCell.vue'
import AccountProxyCell from '../accounts/AccountProxyCell.vue'
import AccountRateMultiplierCell from '../accounts/AccountRateMultiplierCell.vue'
import AccountSchedulableToggle from '../accounts/AccountSchedulableToggle.vue'

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

vi.mock('@/utils/format', () => ({
  formatRelativeTime: (value: string | null) => `relative:${value}`,
  formatDateTime: () => '2026-01-01 00:00'
}))

function createAccount(overrides: Partial<Account> = {}): Account {
  return {
    id: 1,
    name: 'Main Account',
    notes: 'Primary account notes',
    platform: 'antigravity',
    type: 'oauth',
    credentials: {
      plan_type: 'pro'
    },
    extra: {
      email_address: 'main@example.com',
      load_code_assist: {
        paidTier: { id: 'g1-pro-tier' }
      }
    },
    proxy_id: null,
    concurrency: 1,
    priority: 8,
    rate_multiplier: 1.25,
    status: 'active',
    error_message: null,
    last_used_at: '2026-04-04T00:00:00Z',
    expires_at: 1,
    auto_pause_on_expired: true,
    created_at: '2026-04-01T00:00:00Z',
    updated_at: '2026-04-01T00:00:00Z',
    proxy: null,
    schedulable: true,
    rate_limited_at: null,
    rate_limit_reset_at: null,
    overload_until: null,
    temp_unschedulable_until: null,
    temp_unschedulable_reason: null,
    session_window_start: null,
    session_window_end: null,
    session_window_status: null,
    ...overrides
  }
}

describe('account table cells', () => {
  it('renders name, notes, proxy, and rate cells', () => {
    const account = createAccount({
      proxy: {
        id: 2,
        name: 'Tokyo Proxy',
        country_code: 'JP'
      } as any
    })

    const nameWrapper = mount(AccountNameCell, {
      props: {
        account
      }
    })
    expect(nameWrapper.text()).toContain('Main Account')
    expect(nameWrapper.text()).toContain('main@example.com')

    const notesWrapper = mount(AccountNotesCell, {
      props: {
        notes: account.notes
      }
    })
    expect(notesWrapper.text()).toContain('Primary account notes')

    const proxyWrapper = mount(AccountProxyCell, {
      props: {
        proxy: account.proxy
      }
    })
    expect(proxyWrapper.text()).toContain('Tokyo Proxy')
    expect(proxyWrapper.text()).toContain('JP')

    const rateWrapper = mount(AccountRateMultiplierCell, {
      props: {
        rateMultiplier: account.rate_multiplier
      }
    })
    expect(rateWrapper.text()).toContain('1.25x')
  })

  it('renders platform, expiration, and last-used state', () => {
    const account = createAccount()

    const platformWrapper = mount(AccountPlatformTypeCell, {
      props: {
        account
      },
      global: {
        stubs: {
          PlatformTypeBadge: {
            template: '<div>badge</div>'
          }
        }
      }
    })
    expect(platformWrapper.text()).toContain('badge')
    expect(platformWrapper.text()).toContain('admin.accounts.tier.pro')

    const expiresWrapper = mount(AccountExpiresCell, {
      props: {
        account,
        value: 1
      }
    })
    expect(expiresWrapper.text()).toContain('2026-01-01 00:00')
    expect(expiresWrapper.text()).toContain('admin.accounts.expired')
    expect(expiresWrapper.text()).toContain('admin.accounts.autoPauseOnExpired')

    const lastUsedWrapper = mount(AccountLastUsedCell, {
      props: {
        value: account.last_used_at
      }
    })
    expect(lastUsedWrapper.text()).toContain('relative:2026-04-04T00:00:00Z')
  })

  it('renders Grok runtime tier, quota, and capability hints', () => {
    const account = createAccount({
      name: 'Grok Session',
      platform: 'grok',
      type: 'session',
      credentials: {
        session_token: 'secret'
      },
      extra: {
        grok: {
          auth_mode: 'session',
          auth_fingerprint: 'sha256:ab12...cd34',
          tier: {
            normalized: 'heavy'
          },
          sync_state: {
            last_sync_at: '2026-04-20T00:30:00Z',
            last_probe_at: '2026-04-20T01:00:00Z',
            last_probe_error: 'API returned 401 Unauthorized',
            last_probe_status_code: 401
          },
          capabilities: {
            operations: ['chat', 'video']
          },
          quota_windows: {
            auto: {
              remaining: 17,
              total: 150,
              source: 'sync'
            },
            heavy: {
              remaining: 3,
              total: 20,
              source: 'sync'
            }
          }
        }
      }
    })

    const nameWrapper = mount(AccountNameCell, {
      props: {
        account
      }
    })
    expect(nameWrapper.text()).toContain('admin.accounts.grok.runtime.lastSyncAt')
    expect(nameWrapper.text()).toContain('relative:2026-04-20T00:30:00Z')
    expect(nameWrapper.text()).toContain('admin.accounts.grok.runtime.lastProbeError')
    expect(nameWrapper.text()).toContain('admin.accounts.grok.runtime.probeFailedWithCode')
    expect(nameWrapper.text()).not.toContain('API returned 401 Unauthorized')

    const platformWrapper = mount(AccountPlatformTypeCell, {
      props: {
        account
      },
      global: {
        stubs: {
          PlatformTypeBadge: {
            template: '<div>badge</div>'
          }
        }
      }
    })
    expect(platformWrapper.text()).toContain('admin.accounts.grok.runtime.tiers.heavy')

    const capacityWrapper = mount(AccountCapacityCell, {
      props: {
        account
      }
    })
    expect(capacityWrapper.text()).toContain('admin.accounts.grok.runtime.windows.auto')
    expect(capacityWrapper.text()).toContain('17')
    expect(capacityWrapper.text()).toContain('150')
    expect(capacityWrapper.text()).toContain('admin.accounts.grok.runtime.capabilities.video')
  })

  it('emits schedulable and action events', async () => {
    const account = createAccount()

    const toggleWrapper = mount(AccountSchedulableToggle, {
      props: {
        account,
        loading: false
      }
    })
    await toggleWrapper.find('button').trigger('click')
    expect(toggleWrapper.emitted('toggle')?.[0]).toEqual([account])

    const actionsWrapper = mount(AccountActionsCell, {
      props: {
        account
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })
    const buttons = actionsWrapper.findAll('button')
    await buttons[0].trigger('click')
    await buttons[1].trigger('click')
    await buttons[2].trigger('click')

    expect(actionsWrapper.emitted('edit')?.[0]).toEqual([account])
    expect(actionsWrapper.emitted('delete')?.[0]).toEqual([account])
    expect(actionsWrapper.emitted('open-menu')?.[0]?.[0]).toEqual(account)
  })
})
