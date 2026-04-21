<template>
  <Teleport to="body">
    <Transition name="modal">
      <div
        v-if="show"
        class="modal-overlay sm:items-center"
        :class="{ 'items-end': true, 'sm:items-center': true }"
        :style="zIndexStyle"
        :aria-labelledby="dialogId"
        role="dialog"
        aria-modal="true"
        @click.self="handleClose"
      >
        <!-- Modal panel -->
        <div ref="dialogRef" :class="['modal-content modal-sheet', widthClasses]" @click.stop>
          <!-- Drag Handle (mobile only) -->
          <div class="flex justify-center pt-2 pb-0 sm:hidden">
            <div class="modal-sheet__handle h-1 w-10 rounded-full"></div>
          </div>
          <!-- Header -->
          <div class="modal-header">
            <h2 :id="dialogId" class="modal-title">
              {{ title }}
            </h2>
            <button
              @click="emit('close')"
              class="modal-close-button transition-colors"
              aria-label="Close modal"
            >
              <Icon name="x" size="md" />
            </button>
          </div>

          <!-- Body -->
          <div class="modal-body">
            <slot></slot>
          </div>

          <!-- Footer -->
          <div v-if="$slots.footer" class="modal-footer">
            <slot name="footer"></slot>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, watch, onMounted, onUnmounted, ref, nextTick } from 'vue'
import Icon from '@/components/icons/Icon.vue'
import { lockBodyScroll as acquireBodyScrollLock, unlockBodyScroll as releaseBodyScrollLock } from '@/utils/bodyScrollLock'

// 生成唯一ID以避免多个对话框时ID冲突
let dialogIdCounter = 0
const dialogId = `modal-title-${++dialogIdCounter}`

// 焦点管理
const dialogRef = ref<HTMLElement | null>(null)
let previousActiveElement: HTMLElement | null = null
let bodyScrollLocked = false

type DialogWidth = 'narrow' | 'normal' | 'wide' | 'extra-wide' | 'full'

interface Props {
  show: boolean
  title: string
  width?: DialogWidth
  closeOnEscape?: boolean
  closeOnClickOutside?: boolean
  zIndex?: number
}

interface Emits {
  (e: 'close'): void
}

const props = withDefaults(defineProps<Props>(), {
  width: 'normal',
  closeOnEscape: true,
  closeOnClickOutside: false,
  zIndex: 50
})

const emit = defineEmits<Emits>()

// Custom z-index style (overrides the default z-50 from CSS)
const zIndexStyle = computed(() => {
  return props.zIndex !== 50 ? { zIndex: props.zIndex } : undefined
})

const widthClasses = computed(() => {
  // Width guidance: narrow=confirm/short prompts, normal=standard forms,
  // wide=multi-section forms or rich content, extra-wide=analytics/tables,
  // full=full-screen or very dense layouts.
  const widths: Record<DialogWidth, string> = {
    narrow: 'modal-width--narrow',
    normal: 'modal-width--normal',
    wide: 'modal-width--wide',
    'extra-wide': 'modal-width--extra-wide',
    full: 'modal-width--full'
  }
  return widths[props.width]
})

const handleClose = () => {
  if (props.closeOnClickOutside) {
    emit('close')
  }
}

const handleEscape = (event: KeyboardEvent) => {
  if (props.show && props.closeOnEscape && event.key === 'Escape') {
    emit('close')
  }
}

const lockBodyScroll = () => {
  if (bodyScrollLocked) return
  bodyScrollLocked = true
  acquireBodyScrollLock()
}

const unlockBodyScroll = () => {
  if (!bodyScrollLocked) return false
  bodyScrollLocked = false
  return releaseBodyScrollLock()
}

// Prevent body scroll when modal is open and manage focus
watch(
  () => props.show,
  async (isOpen) => {
    if (isOpen) {
      // 保存当前焦点元素
      previousActiveElement = document.activeElement as HTMLElement
      lockBodyScroll()

      // 等待DOM更新后设置焦点到对话框
      await nextTick()
      if (dialogRef.value) {
        const firstFocusable = dialogRef.value.querySelector<HTMLElement>(
          'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
        )
        firstFocusable?.focus()
      }
    } else {
      const shouldRestoreFocus = unlockBodyScroll()
      // 仅在最后一个对话框关闭时恢复之前的焦点，避免多层对话框互相抢焦点
      if (
        shouldRestoreFocus &&
        previousActiveElement &&
        typeof previousActiveElement.focus === 'function'
      ) {
        previousActiveElement.focus()
      }
      previousActiveElement = null
    }
  },
  { immediate: true }
)

onMounted(() => {
  document.addEventListener('keydown', handleEscape)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleEscape)
  unlockBodyScroll()
})
</script>

<style scoped>
.modal-sheet__handle {
  background: color-mix(in srgb, var(--theme-page-muted) 28%, transparent);
}

.modal-width--narrow {
  max-width: min(100%, var(--theme-dialog-width-narrow));
}

.modal-width--normal {
  max-width: min(100%, var(--theme-dialog-width-normal));
}

.modal-width--wide,
.modal-width--extra-wide,
.modal-width--full {
  width: 100%;
  max-width: 100%;
}

.modal-close-button {
  margin-inline-end: calc(var(--theme-button-padding-y) * -0.2);
  padding: calc(var(--theme-button-padding-y) * 0.8);
  border-radius: calc(var(--theme-button-radius) + 4px);
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
}

.modal-close-button:hover {
  background: var(--theme-button-ghost-hover-bg);
  color: var(--theme-page-text);
}

@media (min-width: 640px) {
  .modal-width--wide {
    max-width: var(--theme-dialog-width-wide-sm);
  }

  .modal-width--extra-wide {
    max-width: var(--theme-dialog-width-extra-sm);
  }

  .modal-width--full {
    max-width: var(--theme-dialog-width-full-sm);
  }
}

@media (min-width: 768px) {
  .modal-width--wide {
    max-width: var(--theme-dialog-width-wide-md);
  }

  .modal-width--extra-wide {
    max-width: var(--theme-dialog-width-extra-md);
  }

  .modal-width--full {
    max-width: var(--theme-dialog-width-full-md);
  }
}

@media (min-width: 1024px) {
  .modal-width--wide {
    max-width: var(--theme-dialog-width-wide-lg);
  }

  .modal-width--extra-wide {
    max-width: var(--theme-dialog-width-extra-lg);
  }

  .modal-width--full {
    max-width: var(--theme-dialog-width-full-lg);
  }
}

@media (min-width: 1280px) {
  .modal-width--extra-wide {
    max-width: var(--theme-dialog-width-extra-xl);
  }

  .modal-width--full {
    max-width: var(--theme-dialog-width-full-xl);
  }
}

/* Mobile bottom sheet styling */
@media (max-width: 639px) {
  .modal-sheet {
    border-bottom-left-radius: 0;
    border-bottom-right-radius: 0;
    border-top-left-radius: var(--theme-dialog-mobile-radius);
    border-top-right-radius: var(--theme-dialog-mobile-radius);
    max-height: 92vh;
    width: 100%;
    max-width: 100% !important;
    margin: 0;
  }
}
</style>
