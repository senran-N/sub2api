import { afterEach, describe, expect, it } from 'vitest'

import {
  getThemeChartAlphaPalette,
  getThemeChartPalette,
  getThemeChartTokens,
  getThemeChartTooltipColors,
  getThemeDoughnutChartConfig,
  getThemeLineChartConfig,
  readThemeCssVariable
} from '../themeStyles'

function setThemeVariables(entries: Record<string, string>) {
  Object.entries(entries).forEach(([name, value]) => {
    document.documentElement.style.setProperty(name, value)
  })
}

afterEach(() => {
  document.documentElement.removeAttribute('style')
})

describe('themeStyles', () => {
  it('reads typed chart palette domains from CSS variables', () => {
    setThemeVariables({
      '--theme-chart-seq-1-rgb': '1 2 3',
      '--theme-chart-seq-2-rgb': '4 5 6',
      '--theme-chart-seq-3-rgb': '7 8 9',
      '--theme-chart-seq-4-rgb': '10 11 12',
      '--theme-chart-seq-5-rgb': '13 14 15',
      '--theme-chart-seq-6-rgb': '16 17 18',
      '--theme-chart-seq-7-rgb': '19 20 21',
      '--theme-chart-seq-8-rgb': '22 23 24',
      '--theme-chart-seq-9-rgb': '25 26 27',
      '--theme-chart-seq-10-rgb': '28 29 30',
      '--theme-chart-seq-11-rgb': '31 32 33',
      '--theme-chart-seq-12-rgb': '34 35 36',
      '--theme-surface-contrast': '#111111',
      '--theme-surface-contrast-text': '#f5f5f5',
      '--theme-chart-donut-cutout': '70',
      '--theme-chart-donut-border-radius': '6',
      '--theme-chart-donut-spacing': '5',
      '--theme-chart-point-radius': '2',
      '--theme-chart-point-hover-radius': '7',
    })

    expect(getThemeChartPalette()).toEqual({
      colors: [
        'rgb(1 2 3)',
        'rgb(4 5 6)',
        'rgb(7 8 9)',
        'rgb(10 11 12)',
        'rgb(13 14 15)',
        'rgb(16 17 18)',
        'rgb(19 20 21)',
        'rgb(22 23 24)',
        'rgb(25 26 27)',
        'rgb(28 29 30)',
        'rgb(31 32 33)',
        'rgb(34 35 36)'
      ]
    })

    expect(getThemeChartAlphaPalette(0.2)).toEqual({
      alpha: 0.2,
      colors: [
        'rgb(1 2 3 / 0.2)',
        'rgb(4 5 6 / 0.2)',
        'rgb(7 8 9 / 0.2)',
        'rgb(10 11 12 / 0.2)',
        'rgb(13 14 15 / 0.2)',
        'rgb(16 17 18 / 0.2)',
        'rgb(19 20 21 / 0.2)',
        'rgb(22 23 24 / 0.2)',
        'rgb(25 26 27 / 0.2)',
        'rgb(28 29 30 / 0.2)',
        'rgb(31 32 33 / 0.2)',
        'rgb(34 35 36 / 0.2)'
      ]
    })

    expect(getThemeChartTooltipColors()).toEqual({
      background: '#111111',
      text: '#f5f5f5'
    })
    expect(getThemeDoughnutChartConfig()).toEqual({
      cutout: '70%',
      borderRadius: 6,
      spacing: 5,
      hoverOffset: 8
    })
    expect(getThemeLineChartConfig()).toEqual({
      pointRadius: 2,
      pointHoverRadius: 7
    })
  })

  it('returns a single typed chart token bundle', () => {
    setThemeVariables({
      '--theme-chart-seq-1-rgb': '1 2 3',
      '--theme-surface-contrast': '#101010',
      '--theme-surface-contrast-text': '#fafafa'
    })

    const tokens = getThemeChartTokens(0.3)
    expect(tokens.palette.colors[0]).toBe('rgb(1 2 3)')
    expect(tokens.alphaPalette).toEqual({
      alpha: 0.3,
      colors: expect.arrayContaining(['rgb(1 2 3 / 0.3)'])
    })
    expect(tokens.tooltip).toEqual({
      background: '#101010',
      text: '#fafafa'
    })
  })

  it('falls back to default tokens when CSS variables are absent', () => {
    expect(readThemeCssVariable('--theme-accent')).toBe('#C43C00')
    expect(getThemeChartPalette().colors[0]).toBe('rgb(37 99 235)')
  })
})
