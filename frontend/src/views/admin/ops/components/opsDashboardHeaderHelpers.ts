import type { OpsDashboardOverview, OpsJobHeartbeat, OpsMetricThresholds } from '@/api/admin/ops'

export type ThresholdLevel = 'normal' | 'warning' | 'critical'
export type DiagnosisSeverity = 'critical' | 'warning' | 'info'

export interface DiagnosisItem {
  type: DiagnosisSeverity
  message: string
  impact: string
  action?: string
}

interface BuildDiagnosisReportOptions {
  overview?: OpsDashboardOverview | null
  isSystemIdle: boolean
  healthScore: number | null
  t: (key: string, params?: Record<string, unknown>) => string
}

type OpsStatusTone = 'ok' | 'warning' | 'critical' | 'unknown'

interface StatusDisplay {
  label: string
  className: string
}

interface JobsStatusDisplay extends StatusDisplay {
  status: 'ok' | 'warn' | 'unknown'
  warnCount: number
}

export function formatCustomTimeRangeLabel(startTime: string, endTime: string): string {
  const start = new Date(startTime)
  const end = new Date(endTime)
  const formatDate = (value: Date) => {
    const month = String(value.getMonth() + 1).padStart(2, '0')
    const day = String(value.getDate()).padStart(2, '0')
    const hour = String(value.getHours()).padStart(2, '0')
    const minute = String(value.getMinutes()).padStart(2, '0')
    return `${month}-${day} ${hour}:${minute}`
  }
  return `${formatDate(start)} ~ ${formatDate(end)}`
}

export function getSLAThresholdLevel(
  slaPercent: number | null,
  thresholds?: OpsMetricThresholds | null
): ThresholdLevel {
  if (slaPercent == null) return 'normal'
  const threshold = thresholds?.sla_percent_min
  if (threshold == null) return 'normal'

  const warningBuffer = 0.1
  if (slaPercent < threshold) return 'critical'
  if (slaPercent < threshold + warningBuffer) return 'warning'
  return 'normal'
}

export function getTTFTThresholdLevel(
  ttftMs: number | null,
  thresholds?: OpsMetricThresholds | null
): ThresholdLevel {
  if (ttftMs == null) return 'normal'
  const threshold = thresholds?.ttft_p99_ms_max
  if (threshold == null) return 'normal'
  if (ttftMs >= threshold) return 'critical'
  if (ttftMs >= threshold * 0.8) return 'warning'
  return 'normal'
}

export function getRequestErrorRateThresholdLevel(
  errorRatePercent: number | null,
  thresholds?: OpsMetricThresholds | null
): ThresholdLevel {
  if (errorRatePercent == null) return 'normal'
  const threshold = thresholds?.request_error_rate_percent_max
  if (threshold == null) return 'normal'
  if (errorRatePercent >= threshold) return 'critical'
  if (errorRatePercent >= threshold * 0.8) return 'warning'
  return 'normal'
}

export function getUpstreamErrorRateThresholdLevel(
  upstreamErrorRatePercent: number | null,
  thresholds?: OpsMetricThresholds | null
): ThresholdLevel {
  if (upstreamErrorRatePercent == null) return 'normal'
  const threshold = thresholds?.upstream_error_rate_percent_max
  if (threshold == null) return 'normal'
  if (upstreamErrorRatePercent >= threshold) return 'critical'
  if (upstreamErrorRatePercent >= threshold * 0.8) return 'warning'
  return 'normal'
}

export function getThresholdColorClass(level: ThresholdLevel): string {
  switch (level) {
    case 'critical':
      return 'ops-dashboard-header__tone ops-dashboard-header__tone--critical'
    case 'warning':
      return 'ops-dashboard-header__tone ops-dashboard-header__tone--warning'
    default:
      return 'ops-dashboard-header__tone ops-dashboard-header__tone--healthy'
  }
}

export function buildDiagnosisReport({
  overview,
  isSystemIdle,
  healthScore,
  t
}: BuildDiagnosisReportOptions): DiagnosisItem[] {
  if (!overview) {
    return []
  }

  const report: DiagnosisItem[] = []

  if (isSystemIdle) {
    report.push({
      type: 'info',
      message: t('admin.ops.diagnosis.idle'),
      impact: t('admin.ops.diagnosis.idleImpact')
    })
    return report
  }

  const systemMetrics = overview.system_metrics
  if (systemMetrics) {
    if (systemMetrics.db_ok === false) {
      report.push({
        type: 'critical',
        message: t('admin.ops.diagnosis.dbDown'),
        impact: t('admin.ops.diagnosis.dbDownImpact'),
        action: t('admin.ops.diagnosis.dbDownAction')
      })
    }

    if (systemMetrics.redis_ok === false) {
      report.push({
        type: 'warning',
        message: t('admin.ops.diagnosis.redisDown'),
        impact: t('admin.ops.diagnosis.redisDownImpact'),
        action: t('admin.ops.diagnosis.redisDownAction')
      })
    }

    const cpuPct = systemMetrics.cpu_usage_percent ?? 0
    if (cpuPct > 90) {
      report.push({
        type: 'critical',
        message: t('admin.ops.diagnosis.cpuCritical', { usage: cpuPct.toFixed(1) }),
        impact: t('admin.ops.diagnosis.cpuCriticalImpact'),
        action: t('admin.ops.diagnosis.cpuCriticalAction')
      })
    } else if (cpuPct > 80) {
      report.push({
        type: 'warning',
        message: t('admin.ops.diagnosis.cpuHigh', { usage: cpuPct.toFixed(1) }),
        impact: t('admin.ops.diagnosis.cpuHighImpact'),
        action: t('admin.ops.diagnosis.cpuHighAction')
      })
    }

    const memoryPct = systemMetrics.memory_usage_percent ?? 0
    if (memoryPct > 90) {
      report.push({
        type: 'critical',
        message: t('admin.ops.diagnosis.memoryCritical', { usage: memoryPct.toFixed(1) }),
        impact: t('admin.ops.diagnosis.memoryCriticalImpact'),
        action: t('admin.ops.diagnosis.memoryCriticalAction')
      })
    } else if (memoryPct > 85) {
      report.push({
        type: 'warning',
        message: t('admin.ops.diagnosis.memoryHigh', { usage: memoryPct.toFixed(1) }),
        impact: t('admin.ops.diagnosis.memoryHighImpact'),
        action: t('admin.ops.diagnosis.memoryHighAction')
      })
    }
  }

  const ttftP99 = overview.ttft?.p99_ms ?? 0
  if (ttftP99 > 500) {
    report.push({
      type: 'warning',
      message: t('admin.ops.diagnosis.ttftHigh', { ttft: ttftP99.toFixed(0) }),
      impact: t('admin.ops.diagnosis.ttftHighImpact'),
      action: t('admin.ops.diagnosis.ttftHighAction')
    })
  }

  const upstreamRatePct = (overview.upstream_error_rate ?? 0) * 100
  if (upstreamRatePct > 5) {
    report.push({
      type: 'critical',
      message: t('admin.ops.diagnosis.upstreamCritical', { rate: upstreamRatePct.toFixed(2) }),
      impact: t('admin.ops.diagnosis.upstreamCriticalImpact'),
      action: t('admin.ops.diagnosis.upstreamCriticalAction')
    })
  } else if (upstreamRatePct > 2) {
    report.push({
      type: 'warning',
      message: t('admin.ops.diagnosis.upstreamHigh', { rate: upstreamRatePct.toFixed(2) }),
      impact: t('admin.ops.diagnosis.upstreamHighImpact'),
      action: t('admin.ops.diagnosis.upstreamHighAction')
    })
  }

  const errorPct = (overview.error_rate ?? 0) * 100
  if (errorPct > 3) {
    report.push({
      type: 'critical',
      message: t('admin.ops.diagnosis.errorHigh', { rate: errorPct.toFixed(2) }),
      impact: t('admin.ops.diagnosis.errorHighImpact'),
      action: t('admin.ops.diagnosis.errorHighAction')
    })
  } else if (errorPct > 0.5) {
    report.push({
      type: 'warning',
      message: t('admin.ops.diagnosis.errorElevated', { rate: errorPct.toFixed(2) }),
      impact: t('admin.ops.diagnosis.errorElevatedImpact'),
      action: t('admin.ops.diagnosis.errorElevatedAction')
    })
  }

  const slaPct = (overview.sla ?? 0) * 100
  if (slaPct < 90) {
    report.push({
      type: 'critical',
      message: t('admin.ops.diagnosis.slaCritical', { sla: slaPct.toFixed(2) }),
      impact: t('admin.ops.diagnosis.slaCriticalImpact'),
      action: t('admin.ops.diagnosis.slaCriticalAction')
    })
  } else if (slaPct < 98) {
    report.push({
      type: 'warning',
      message: t('admin.ops.diagnosis.slaLow', { sla: slaPct.toFixed(2) }),
      impact: t('admin.ops.diagnosis.slaLowImpact'),
      action: t('admin.ops.diagnosis.slaLowAction')
    })
  }

  if (healthScore != null) {
    if (healthScore < 60) {
      report.push({
        type: 'critical',
        message: t('admin.ops.diagnosis.healthCritical', { score: healthScore }),
        impact: t('admin.ops.diagnosis.healthCriticalImpact'),
        action: t('admin.ops.diagnosis.healthCriticalAction')
      })
    } else if (healthScore < 90) {
      report.push({
        type: 'warning',
        message: t('admin.ops.diagnosis.healthLow', { score: healthScore }),
        impact: t('admin.ops.diagnosis.healthLowImpact'),
        action: t('admin.ops.diagnosis.healthLowAction')
      })
    }
  }

  if (report.length === 0) {
    report.push({
      type: 'info',
      message: t('admin.ops.diagnosis.healthy'),
      impact: t('admin.ops.diagnosis.healthyImpact')
    })
  }

  return report
}

export function formatTimeShort(value?: string | null): string {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '-'
  return date.toLocaleTimeString()
}

function getOpsStatusClass(status: OpsStatusTone): string {
  switch (status) {
    case 'ok':
      return 'theme-text-success'
    case 'warning':
      return 'theme-text-warning'
    case 'critical':
      return 'theme-text-danger'
    default:
      return 'theme-text-strong'
  }
}

export function buildPoolUsageDisplay(
  ok: boolean | null | undefined,
  usagePercent: number | null,
  t: (key: string) => string
): StatusDisplay {
  if (ok === false) {
    return {
      label: 'FAIL',
      className: getOpsStatusClass('critical')
    }
  }

  if (usagePercent != null) {
    const status: OpsStatusTone =
      usagePercent >= 90 ? 'critical' : usagePercent >= 70 ? 'warning' : 'ok'
    return {
      label: `${usagePercent.toFixed(0)}%`,
      className: getOpsStatusClass(status)
    }
  }

  if (ok === true) {
    return {
      label: t('admin.ops.ok'),
      className: getOpsStatusClass('ok')
    }
  }

  return {
    label: t('admin.ops.noData'),
    className: getOpsStatusClass('unknown')
  }
}

export function buildGoroutineStatusDisplay(
  goroutineCount: number | null,
  t: (key: string) => string,
  warnThreshold = 8_000,
  criticalThreshold = 15_000
): StatusDisplay & { status: OpsStatusTone } {
  const status: OpsStatusTone =
    goroutineCount == null
      ? 'unknown'
      : goroutineCount >= criticalThreshold
        ? 'critical'
        : goroutineCount >= warnThreshold
          ? 'warning'
          : 'ok'

  const label =
    status === 'ok'
      ? t('admin.ops.ok')
      : status === 'warning'
        ? t('common.warning')
        : status === 'critical'
          ? t('common.critical')
          : t('admin.ops.noData')

  return {
    status,
    label,
    className: getOpsStatusClass(status)
  }
}

function isJobHeartbeatWarning(heartbeat: OpsJobHeartbeat | null | undefined) {
  if (!heartbeat) return false
  return Boolean(
    heartbeat.last_error_at &&
      (!heartbeat.last_success_at || heartbeat.last_error_at > heartbeat.last_success_at)
  )
}

export function buildJobsStatusDisplay(
  jobHeartbeats: Array<OpsJobHeartbeat | null | undefined>,
  t: (key: string) => string
): JobsStatusDisplay {
  if (!jobHeartbeats.length) {
    return {
      status: 'unknown',
      warnCount: 0,
      label: t('admin.ops.noData'),
      className: getOpsStatusClass('unknown')
    }
  }

  let warnCount = 0
  for (const heartbeat of jobHeartbeats) {
    if (isJobHeartbeatWarning(heartbeat)) {
      warnCount += 1
    }
  }

  const status = warnCount > 0 ? 'warn' : 'ok'
  return {
    status,
    warnCount,
    label: status === 'warn' ? t('common.warning') : t('admin.ops.ok'),
    className: getOpsStatusClass(status === 'warn' ? 'warning' : 'ok')
  }
}
