import type {
  Proxy,
  ProxyQualityCheckItem,
  QualityCategoryScores
} from '@/types'

type Translate = (key: string, params?: Record<string, unknown>) => string

export interface ProxyQualityCategoryEntry {
  key: keyof QualityCategoryScores
  label: string
  weight: number
  score: number
}

export function formatProxyLocation(proxy: Pick<Proxy, 'country' | 'city'>) {
  const parts = [proxy.country, proxy.city].filter(Boolean) as string[]
  return parts.join(' · ')
}

export function buildProxyFlagUrl(code: string) {
  return `https://unpkg.com/flag-icons/flags/4x3/${code.toLowerCase()}.svg`
}

export function getQualityStatusClass(status: ProxyQualityCheckItem['status']) {
  if (status === 'pass') {
    return 'badge-success'
  }
  if (status === 'warn') {
    return 'badge-warning'
  }
  if (status === 'challenge') {
    return 'badge-danger'
  }
  if (status === 'skip') {
    return 'badge-gray'
  }
  return 'badge-danger'
}

export function getQualityStatusLabel(status: ProxyQualityCheckItem['status'], t: Translate) {
  if (status === 'pass') {
    return t('admin.proxies.qualityStatusPass')
  }
  if (status === 'warn') {
    return t('admin.proxies.qualityStatusWarn')
  }
  if (status === 'challenge') {
    return t('admin.proxies.qualityStatusChallenge')
  }
  if (status === 'skip') {
    return t('admin.proxies.qualityStatusSkip')
  }
  return t('admin.proxies.qualityStatusFail')
}

export function getQualityOverallClass(status?: Proxy['quality_status']) {
  if (status === 'healthy') {
    return 'badge-success'
  }
  if (status === 'warn') {
    return 'badge-warning'
  }
  if (status === 'challenge') {
    return 'badge-danger'
  }
  return 'badge-danger'
}

export function getQualityOverallLabel(status: Proxy['quality_status'] | undefined, t: Translate) {
  if (status === 'healthy') {
    return t('admin.proxies.qualityStatusHealthy')
  }
  if (status === 'warn') {
    return t('admin.proxies.qualityStatusWarn')
  }
  if (status === 'challenge') {
    return t('admin.proxies.qualityStatusChallenge')
  }
  return t('admin.proxies.qualityStatusFail')
}

export function getQualityTargetLabel(target: string, t: Translate) {
  switch (target) {
    case 'base_connectivity':
      return t('admin.proxies.qualityTargetBase')
    case 'openai':
      return 'OpenAI'
    case 'anthropic':
      return 'Anthropic'
    case 'gemini':
      return 'Gemini'
    case 'ip_type':
      return t('admin.proxies.qualityTargetIPType')
    case 'abuse_check':
      return t('admin.proxies.qualityTargetAbuse')
    case 'dns_leak':
      return t('admin.proxies.qualityTargetDNSLeak')
    default:
      return target
  }
}

export function getIpTypeBadgeClass(ipType: string) {
  switch (ipType) {
    case 'residential':
      return 'badge-success'
    case 'mobile':
      return 'badge-info'
    case 'datacenter':
      return 'badge-danger'
    case 'vpn':
      return 'badge-warning'
    case 'tor':
      return 'badge-danger'
    default:
      return 'badge-gray'
  }
}

export function getIpTypeLabel(ipType: string, t: Translate) {
  switch (ipType) {
    case 'residential':
      return t('admin.proxies.qualityIPTypeResidential')
    case 'mobile':
      return t('admin.proxies.qualityIPTypeMobile')
    case 'datacenter':
      return t('admin.proxies.qualityIPTypeDatacenter')
    case 'vpn':
      return t('admin.proxies.qualityIPTypeVPN')
    case 'tor':
      return t('admin.proxies.qualityIPTypeTor')
    default:
      return ipType
  }
}

export function getDnsLeakBadgeClass(risk: string) {
  if (risk === 'possible') {
    return 'badge-warning'
  }
  if (risk === 'detected') {
    return 'badge-danger'
  }
  return 'badge-gray'
}

export function getDnsLeakLabel(risk: string, t: Translate) {
  switch (risk) {
    case 'none':
      return t('admin.proxies.qualityDNSLeakNone')
    case 'possible':
      return t('admin.proxies.qualityDNSLeakPossible')
    case 'detected':
      return t('admin.proxies.qualityDNSLeakDetected')
    default:
      return risk
  }
}

export function getProxyScoreBarColor(score: number) {
  if (score >= 80) {
    return 'theme-progress-fill--success'
  }
  if (score >= 60) {
    return 'theme-progress-fill--warning'
  }
  if (score >= 40) {
    return 'theme-progress-fill--brand-orange'
  }
  return 'theme-progress-fill--danger'
}

export function buildProxyQualityCategoryEntries(
  categoryScores: QualityCategoryScores | undefined,
  t: Translate
): ProxyQualityCategoryEntry[] {
  if (!categoryScores) {
    return []
  }

  return [
    {
      key: 'reachability',
      label: t('admin.proxies.qualityCategoryReachability'),
      weight: 30,
      score: categoryScores.reachability
    },
    {
      key: 'ip_risk',
      label: t('admin.proxies.qualityCategoryIPRisk'),
      weight: 25,
      score: categoryScores.ip_risk
    },
    {
      key: 'ip_type',
      label: t('admin.proxies.qualityCategoryIPType'),
      weight: 20,
      score: categoryScores.ip_type
    },
    {
      key: 'abuse_history',
      label: t('admin.proxies.qualityCategoryAbuse'),
      weight: 15,
      score: categoryScores.abuse_history
    },
    {
      key: 'latency',
      label: t('admin.proxies.qualityCategoryLatency'),
      weight: 10,
      score: categoryScores.latency
    }
  ]
}
