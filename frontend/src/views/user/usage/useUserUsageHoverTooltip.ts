import { ref } from 'vue'

export interface UserUsageHoverTooltipPosition {
  x: number
  y: number
}

export function useUserUsageHoverTooltip<T>() {
  const visible = ref(false)
  const position = ref<UserUsageHoverTooltipPosition>({ x: 0, y: 0 })
  const data = ref<T | null>(null)

  const show = (event: MouseEvent, nextData: T) => {
    const target = event.currentTarget

    if (!(target instanceof HTMLElement)) {
      throw new TypeError('Tooltip trigger must be an HTMLElement')
    }

    const rect = target.getBoundingClientRect()
    data.value = nextData
    position.value = {
      x: rect.right + 8,
      y: rect.top + rect.height / 2
    }
    visible.value = true
  }

  const hide = () => {
    visible.value = false
    data.value = null
  }

  return {
    visible,
    position,
    data,
    show,
    hide
  }
}
