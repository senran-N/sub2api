import { describe, expect, it } from 'vitest'
import { buildSoraTabs, resolveSoraDashboardPath } from '../sora/soraView'

describe('soraView helpers', () => {
  it('resolves dashboard paths by role', () => {
    expect(resolveSoraDashboardPath(true)).toBe('/admin/dashboard')
    expect(resolveSoraDashboardPath(false)).toBe('/dashboard')
  })

  it('builds the available sora tabs in order', () => {
    const t = (key: string) => key
    expect(buildSoraTabs(t)).toEqual([
      { key: 'generate', label: 'sora.tabGenerate' },
      { key: 'library', label: 'sora.tabLibrary' }
    ])
  })
})
