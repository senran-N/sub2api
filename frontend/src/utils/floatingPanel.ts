export interface FloatingPanelPosition {
  top: number
  left: number
}

export interface FloatingPanelClampMetrics {
  panelWidth: number
  panelHeight: number
  viewportWidth?: number
  viewportHeight?: number
  padding?: number
}

export interface ContextMenuPositionMetrics {
  rect: Pick<DOMRect, 'left' | 'top' | 'bottom' | 'width'>
  pointerX: number
  pointerY: number
  viewportWidth: number
  viewportHeight: number
  panelWidth: number
  panelHeight: number
  padding?: number
  gap?: number
  mobileBreakpoint?: number
}

const DEFAULT_PADDING = 8
const DEFAULT_GAP = 4
const DEFAULT_MOBILE_BREAKPOINT = 768

export function readThemePixelValue(variableName: string, fallback: number): number {
  if (typeof window === 'undefined' || typeof document === 'undefined') {
    return fallback
  }

  const rawValue = getComputedStyle(document.documentElement).getPropertyValue(variableName).trim()
  if (!rawValue) {
    return fallback
  }

  const numericValue = Number.parseFloat(rawValue)
  if (!Number.isFinite(numericValue)) {
    return fallback
  }

  if (rawValue.endsWith('rem') || rawValue.endsWith('em')) {
    const rootFontSize = Number.parseFloat(getComputedStyle(document.documentElement).fontSize)
    return Number.isFinite(rootFontSize) ? numericValue * rootFontSize : fallback
  }

  return numericValue
}

export function calculateContextMenuPosition(
  metrics: ContextMenuPositionMetrics
): FloatingPanelPosition {
  const padding = metrics.padding ?? DEFAULT_PADDING
  const gap = metrics.gap ?? DEFAULT_GAP
  const mobileBreakpoint = metrics.mobileBreakpoint ?? DEFAULT_MOBILE_BREAKPOINT

  if (metrics.viewportWidth < mobileBreakpoint) {
    const left = Math.max(
      padding,
      Math.min(
        metrics.rect.left + metrics.rect.width / 2 - metrics.panelWidth / 2,
        metrics.viewportWidth - metrics.panelWidth - padding
      )
    )

    let top = metrics.rect.bottom + gap
    if (top + metrics.panelHeight > metrics.viewportHeight - padding) {
      top = metrics.rect.top - metrics.panelHeight - gap
      if (top < padding) {
        top = padding
      }
    }

    return { top, left }
  }

  const left = Math.max(
    padding,
    Math.min(
      metrics.pointerX - metrics.panelWidth,
      metrics.viewportWidth - metrics.panelWidth - padding
    )
  )
  let top = metrics.pointerY
  if (top + metrics.panelHeight > metrics.viewportHeight - padding) {
    top = metrics.viewportHeight - metrics.panelHeight - padding
  }

  return { top, left }
}

export function clampFloatingPanelPosition(
  desiredPosition: FloatingPanelPosition,
  metrics: FloatingPanelClampMetrics
): FloatingPanelPosition {
  const padding = metrics.padding ?? DEFAULT_PADDING
  const viewportWidth = metrics.viewportWidth ?? window.innerWidth
  const viewportHeight = metrics.viewportHeight ?? window.innerHeight

  const maxLeft = Math.max(padding, viewportWidth - metrics.panelWidth - padding)
  const maxTop = Math.max(padding, viewportHeight - metrics.panelHeight - padding)

  return {
    left: Math.min(Math.max(desiredPosition.left, padding), maxLeft),
    top: Math.min(Math.max(desiredPosition.top, padding), maxTop)
  }
}
