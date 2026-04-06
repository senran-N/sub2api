import { describe, expect, it } from 'vitest'
import {
  buildDiagnosisReport,
  buildGoroutineStatusDisplay,
  buildJobsStatusDisplay,
  buildPoolUsageDisplay,
  formatTimeShort,
  formatCustomTimeRangeLabel,
  getRequestErrorRateThresholdLevel,
  getSLAThresholdLevel,
  getThresholdColorClass,
  getTTFTThresholdLevel,
  getUpstreamErrorRateThresholdLevel
} from '../opsDashboardHeaderHelpers'

describe('opsDashboardHeaderHelpers', () => {
  it('formats custom time range labels', () => {
    expect(
      formatCustomTimeRangeLabel('2026-04-05T10:00:00.000Z', '2026-04-05T11:30:00.000Z')
    ).toMatch(/04-05 \d{2}:\d{2} ~ 04-05 \d{2}:\d{2}/)
  })

  it('evaluates threshold levels', () => {
    expect(getSLAThresholdLevel(94.8, { sla_percent_min: 95 } as any)).toBe('critical')
    expect(getSLAThresholdLevel(95.05, { sla_percent_min: 95 } as any)).toBe('warning')
    expect(getTTFTThresholdLevel(900, { ttft_p99_ms_max: 1000 } as any)).toBe('warning')
    expect(
      getRequestErrorRateThresholdLevel(4.1, { request_error_rate_percent_max: 4 } as any)
    ).toBe('critical')
    expect(
      getUpstreamErrorRateThresholdLevel(2.0, { upstream_error_rate_percent_max: 2.5 } as any)
    ).toBe('warning')
  })

  it('maps threshold levels to classes', () => {
    expect(getThresholdColorClass('critical')).toContain('ops-dashboard-header__tone--critical')
    expect(getThresholdColorClass('warning')).toContain('ops-dashboard-header__tone--warning')
    expect(getThresholdColorClass('normal')).toContain('ops-dashboard-header__tone--healthy')
  })

  it('returns idle diagnosis when the system has no traffic', () => {
    const report = buildDiagnosisReport({
      overview: {
        sla: 1,
        error_rate: 0,
        upstream_error_rate: 0,
        ttft: {},
        qps: { current: 0, peak: 0, avg: 0 },
        tps: { current: 0, peak: 0, avg: 0 }
      } as any,
      isSystemIdle: true,
      healthScore: 100,
      t: (key) => key
    })

    expect(report).toEqual([
      {
        type: 'info',
        message: 'admin.ops.diagnosis.idle',
        impact: 'admin.ops.diagnosis.idleImpact'
      }
    ])
  })

  it('builds resource and reliability diagnosis items in priority order', () => {
    const report = buildDiagnosisReport({
      overview: {
        sla: 0.875,
        error_rate: 0.041,
        upstream_error_rate: 0.061,
        ttft: { p99_ms: 780 },
        system_metrics: {
          db_ok: false,
          redis_ok: false,
          cpu_usage_percent: 96.2,
          memory_usage_percent: 88.4
        },
        qps: { current: 2, peak: 2, avg: 2 },
        tps: { current: 3, peak: 3, avg: 3 }
      } as any,
      isSystemIdle: false,
      healthScore: 52,
      t: (key, params) => `${key}:${JSON.stringify(params ?? {})}`
    })

    expect(report.map((item) => item.message)).toEqual([
      'admin.ops.diagnosis.dbDown:{}',
      'admin.ops.diagnosis.redisDown:{}',
      'admin.ops.diagnosis.cpuCritical:{"usage":"96.2"}',
      'admin.ops.diagnosis.memoryHigh:{"usage":"88.4"}',
      'admin.ops.diagnosis.ttftHigh:{"ttft":"780"}',
      'admin.ops.diagnosis.upstreamCritical:{"rate":"6.10"}',
      'admin.ops.diagnosis.errorHigh:{"rate":"4.10"}',
      'admin.ops.diagnosis.slaCritical:{"sla":"87.50"}',
      'admin.ops.diagnosis.healthCritical:{"score":52}'
    ])
    expect(report[0]?.type).toBe('critical')
    expect(report[1]?.type).toBe('warning')
  })

  it('returns a healthy info item when no diagnosis rules are triggered', () => {
    const report = buildDiagnosisReport({
      overview: {
        sla: 0.995,
        error_rate: 0.001,
        upstream_error_rate: 0.001,
        ttft: { p99_ms: 200 },
        system_metrics: {
          db_ok: true,
          redis_ok: true,
          cpu_usage_percent: 20,
          memory_usage_percent: 30
        },
        qps: { current: 8, peak: 10, avg: 6 },
        tps: { current: 12, peak: 15, avg: 9 }
      } as any,
      isSystemIdle: false,
      healthScore: 96,
      t: (key) => key
    })

    expect(report).toEqual([
      {
        type: 'info',
        message: 'admin.ops.diagnosis.healthy',
        impact: 'admin.ops.diagnosis.healthyImpact'
      }
    ])
  })

  it('formats short timestamps defensively', () => {
    expect(formatTimeShort('2026-04-05T10:00:00.000Z')).toMatch(/\d{1,2}:\d{2}:\d{2}/)
    expect(formatTimeShort('not-a-date')).toBe('-')
    expect(formatTimeShort(null)).toBe('-')
  })

  it('builds pool usage displays from availability and utilization', () => {
    expect(buildPoolUsageDisplay(false, 50, (key) => key)).toEqual({
      label: 'FAIL',
      className: 'theme-text-danger'
    })
    expect(buildPoolUsageDisplay(true, 75, (key) => key)).toEqual({
      label: '75%',
      className: 'theme-text-warning'
    })
    expect(buildPoolUsageDisplay(true, null, (key) => key)).toEqual({
      label: 'admin.ops.ok',
      className: 'theme-text-success'
    })
  })

  it('builds goroutine and jobs status displays', () => {
    expect(buildGoroutineStatusDisplay(16_000, (key) => key)).toEqual({
      status: 'critical',
      label: 'common.critical',
      className: 'theme-text-danger'
    })

    expect(
      buildJobsStatusDisplay(
        [
          {
            job_name: 'cleanup',
            updated_at: '2026-04-05T10:00:00.000Z',
            last_success_at: '2026-04-05T09:50:00.000Z',
            last_error_at: '2026-04-05T09:55:00.000Z'
          },
          {
            job_name: 'sync',
            updated_at: '2026-04-05T10:00:00.000Z',
            last_success_at: '2026-04-05T09:58:00.000Z',
            last_error_at: null
          }
        ],
        (key) => key
      )
    ).toEqual({
      status: 'warn',
      warnCount: 1,
      label: 'common.warning',
      className: 'theme-text-warning'
    })
  })
})
