type DeferredTask = () => void

interface DeferredTaskOptions {
  timeout?: number
}

type DeferredTaskHandle = number

function runTask(task: DeferredTask): void {
  task()
}

/**
 * Schedule non-critical work after the first paint/idle period so that
 * initial rendering keeps priority on the main thread.
 */
export function scheduleDeferredTask(
  task: DeferredTask,
  options: DeferredTaskOptions = {}
): () => void {
  if (typeof window === 'undefined') {
    runTask(task)
    return () => {}
  }

  const timeout = options.timeout ?? 1500
  let cancelled = false
  let handle: DeferredTaskHandle | null = null
  let fallbackTimer: number | null = null

  const invoke = () => {
    if (cancelled) {
      return
    }

    if (fallbackTimer !== null) {
      window.clearTimeout(fallbackTimer)
      fallbackTimer = null
    }

    runTask(task)
  }

  if (typeof window.requestIdleCallback === 'function') {
    handle = window.requestIdleCallback(() => {
      handle = null
      invoke()
    }, { timeout })
  } else {
    handle = window.requestAnimationFrame(() => {
      handle = null
      fallbackTimer = window.setTimeout(() => {
        fallbackTimer = null
        invoke()
      }, 0)
    })
  }

  return () => {
    cancelled = true

    if (handle !== null) {
      if (typeof window.cancelIdleCallback === 'function' && typeof window.requestIdleCallback === 'function') {
        window.cancelIdleCallback(handle as number)
      } else {
        window.cancelAnimationFrame(handle as number)
      }
    }

    if (fallbackTimer !== null) {
      window.clearTimeout(fallbackTimer)
    }
  }
}
