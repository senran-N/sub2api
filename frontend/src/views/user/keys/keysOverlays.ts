export interface KeysOverlayPosition {
  top?: number
  bottom?: number
  left: number
}

export function buildKeysMoreMenuPosition(button: HTMLElement): {
  top: number
  left: number
} {
  const rect = button.getBoundingClientRect()
  const menuHeight = 180
  const maxLeft = Math.max(window.innerWidth - 200, 0)
  const spaceBelow = window.innerHeight - rect.bottom

  return {
    top: spaceBelow < menuHeight ? rect.top - menuHeight : rect.bottom + 4,
    left: Math.min(rect.left, maxLeft)
  }
}

export function buildKeysGroupDropdownPosition(
  button: HTMLElement,
  dropdownEstimatedHeight = 400
): KeysOverlayPosition {
  const rect = button.getBoundingClientRect()
  const spaceBelow = window.innerHeight - rect.bottom
  const spaceAbove = rect.top

  if (spaceBelow < dropdownEstimatedHeight && spaceAbove > spaceBelow) {
    return {
      bottom: window.innerHeight - rect.top + 4,
      left: rect.left
    }
  }

  return {
    top: rect.bottom + 4,
    left: rect.left
  }
}
