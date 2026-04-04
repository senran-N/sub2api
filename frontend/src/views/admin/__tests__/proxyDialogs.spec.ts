import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import type { ProxyAccountSummary, ProxyQualityCheckResult } from '@/types'
import ProxyAccountsDialog from '../proxies/ProxyAccountsDialog.vue'
import ProxyQualityReportDialog from '../proxies/ProxyQualityReportDialog.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) =>
      params ? `${key}:${JSON.stringify(params)}` : key
  })
}))

function createQualityReport(
  overrides: Partial<ProxyQualityCheckResult> = {}
): ProxyQualityCheckResult {
  return {
    proxy_id: 1,
    score: 91,
    grade: 'A',
    summary: 'healthy',
    exit_ip: '1.1.1.1',
    country: 'US',
    base_latency_ms: 120,
    passed_count: 2,
    warn_count: 0,
    failed_count: 0,
    challenge_count: 0,
    checked_at: 1712000000,
    items: [
      {
        target: 'openai',
        status: 'pass',
        http_status: 200,
        latency_ms: 111,
        message: 'ok'
      }
    ],
    category_scores: {
      reachability: 95,
      ip_risk: 80,
      ip_type: 75,
      abuse_history: 70,
      latency: 60
    },
    ...overrides
  }
}

describe('proxy dialogs', () => {
  it('renders proxy quality report details', () => {
    const wrapper = mount(ProxyQualityReportDialog, {
      props: {
        show: true,
        proxyName: 'Edge Proxy',
        report: createQualityReport()
      },
      global: {
        stubs: {
          BaseDialog: {
            props: ['show', 'title', 'width'],
            template: '<div><slot /><slot name="footer" /></div>'
          }
        }
      }
    })

    expect(wrapper.text()).toContain('Edge Proxy')
    expect(wrapper.text()).toContain('healthy')
    expect(wrapper.text()).toContain('91')
    expect(wrapper.text()).toContain('admin.proxies.qualityStatusPass')
    expect(wrapper.text()).toContain('OpenAI')
  })

  it('renders accounts dialog loading, empty, and table states', () => {
    const loadingWrapper = mount(ProxyAccountsDialog, {
      props: {
        show: true,
        proxyName: 'Edge Proxy',
        loading: true,
        accounts: []
      },
      global: {
        stubs: {
          BaseDialog: {
            props: ['show', 'title', 'width'],
            template: '<div><slot /><slot name="footer" /></div>'
          },
          Icon: true,
          PlatformTypeBadge: true
        }
      }
    })
    expect(loadingWrapper.text()).toContain('common.loading')

    const emptyWrapper = mount(ProxyAccountsDialog, {
      props: {
        show: true,
        proxyName: 'Edge Proxy',
        loading: false,
        accounts: []
      },
      global: {
        stubs: {
          BaseDialog: {
            props: ['show', 'title', 'width'],
            template: '<div><slot /><slot name="footer" /></div>'
          },
          Icon: true,
          PlatformTypeBadge: true
        }
      }
    })
    expect(emptyWrapper.text()).toContain('admin.proxies.accountsEmpty')

    const accounts: ProxyAccountSummary[] = [
      {
        id: 1,
        name: 'Acct',
        platform: 'openai',
        type: 'api',
        notes: 'note'
      }
    ]

    const dataWrapper = mount(ProxyAccountsDialog, {
      props: {
        show: true,
        proxyName: 'Edge Proxy',
        loading: false,
        accounts
      },
      global: {
        stubs: {
          BaseDialog: {
            props: ['show', 'title', 'width'],
            template: '<div><slot /><slot name="footer" /></div>'
          },
          Icon: true,
          PlatformTypeBadge: true
        }
      }
    })

    expect(dataWrapper.text()).toContain('Acct')
    expect(dataWrapper.text()).toContain('note')
  })
})
