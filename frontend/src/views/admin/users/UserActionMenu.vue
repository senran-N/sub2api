<template>
  <Teleport to="body">
    <div
      v-if="user && position"
      ref="menuRef"
      class="user-action-menu fixed z-[9999] overflow-hidden"
      :style="menuStyle"
    >
      <div class="user-action-menu__content">
        <button
          class="user-action-menu__item flex w-full items-center gap-2 text-sm"
          @click="emitAndClose('api-keys', user)"
        >
          <Icon name="key" size="sm" class="user-action-menu__icon" :stroke-width="2" />
          {{ t('admin.users.apiKeys') }}
        </button>

        <button
          class="user-action-menu__item flex w-full items-center gap-2 text-sm"
          @click="emitAndClose('groups', user)"
        >
          <Icon name="users" size="sm" class="user-action-menu__icon" :stroke-width="2" />
          {{ t('admin.users.groups') }}
        </button>

        <div class="user-action-menu__divider my-1"></div>

        <button
          class="user-action-menu__item flex w-full items-center gap-2 text-sm"
          @click="emitAndClose('deposit', user)"
        >
          <Icon name="plus" size="sm" class="user-action-menu__icon user-action-menu__icon--success" :stroke-width="2" />
          {{ t('admin.users.deposit') }}
        </button>

        <button
          class="user-action-menu__item flex w-full items-center gap-2 text-sm"
          @click="emitAndClose('withdraw', user)"
        >
          <svg class="user-action-menu__icon user-action-menu__icon--warning h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 12H4" />
          </svg>
          {{ t('admin.users.withdraw') }}
        </button>

        <button
          class="user-action-menu__item flex w-full items-center gap-2 text-sm"
          @click="emitAndClose('history', user)"
        >
          <Icon name="dollar" size="sm" class="user-action-menu__icon" :stroke-width="2" />
          {{ t('admin.users.balanceHistory') }}
        </button>

        <div class="user-action-menu__divider my-1"></div>

        <button
          v-if="user.role !== 'admin'"
          class="user-action-menu__item user-action-menu__item--danger flex w-full items-center gap-2 text-sm"
          @click="emitAndClose('delete', user)"
        >
          <Icon name="trash" size="sm" :stroke-width="2" />
          {{ t('common.delete') }}
        </button>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { useDocumentThemeVersion } from '@/composables/useDocumentThemeVersion'
import type { AdminUser } from '@/types'
import { clampFloatingPanelPosition, readThemePixelValue } from '@/utils/floatingPanel'

const props = defineProps<{
  user: AdminUser | null
  position: { top: number; left: number } | null
}>()

const emit = defineEmits<{
  close: []
  'api-keys': [user: AdminUser]
  groups: [user: AdminUser]
  deposit: [user: AdminUser]
  withdraw: [user: AdminUser]
  history: [user: AdminUser]
  delete: [user: AdminUser]
}>()

const { t } = useI18n()
const themeVersion = useDocumentThemeVersion()
const menuRef = ref<HTMLElement | null>(null)
const menuStyle = ref<Record<string, string>>({})

const updateMenuPosition = () => {
  if (!props.user || !props.position) {
    menuStyle.value = {}
    return
  }

  const padding = readThemePixelValue('--theme-floating-panel-viewport-padding', 8)
  const panelWidth = menuRef.value?.offsetWidth ?? readThemePixelValue('--theme-user-action-menu-width', 200)
  const panelHeight = menuRef.value?.offsetHeight ?? readThemePixelValue('--theme-user-action-menu-estimated-height', 240)
  const nextPosition = clampFloatingPanelPosition(props.position, {
    panelWidth,
    panelHeight,
    padding
  })

  menuStyle.value = {
    top: `${nextPosition.top}px`,
    left: `${nextPosition.left}px`
  }
}

watch(
  [() => props.user, () => props.position?.top, () => props.position?.left, themeVersion],
  async () => {
    if (!props.user || !props.position) {
      menuStyle.value = {}
      return
    }

    menuStyle.value = {
      top: `${props.position.top}px`,
      left: `${props.position.left}px`
    }

    await nextTick()
    updateMenuPosition()
  },
  { immediate: true }
)

onMounted(() => {
  window.addEventListener('resize', updateMenuPosition)
})

onUnmounted(() => {
  window.removeEventListener('resize', updateMenuPosition)
})

function emitAndClose(
  event: 'api-keys' | 'groups' | 'deposit' | 'withdraw' | 'history' | 'delete',
  user: AdminUser
) {
  if (event === 'api-keys') {
    emit('api-keys', user)
  } else if (event === 'groups') {
    emit('groups', user)
  } else if (event === 'deposit') {
    emit('deposit', user)
  } else if (event === 'withdraw') {
    emit('withdraw', user)
  } else if (event === 'history') {
    emit('history', user)
  } else {
    emit('delete', user)
  }

  emit('close')
}
</script>

<style scoped>
.user-action-menu {
  width: min(
    var(--theme-user-action-menu-width),
    calc(100vw - (var(--theme-floating-panel-viewport-padding) + var(--theme-floating-panel-viewport-padding)))
  );
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 74%, transparent);
  border-radius: calc(var(--theme-surface-radius) + 4px);
  background: var(--theme-surface);
  box-shadow: var(--theme-dropdown-shadow);
}

.user-action-menu__item {
  padding: var(--theme-user-action-menu-item-padding-y) var(--theme-user-action-menu-item-padding-x);
  color: var(--theme-page-text);
  transition:
    background-color 0.2s ease,
    color 0.2s ease;
}

.user-action-menu__content {
  padding-block: var(--theme-user-action-menu-content-padding-y);
}

.user-action-menu__item:hover {
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
}

.user-action-menu__item--danger {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}

.user-action-menu__item--danger:hover {
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 9%, var(--theme-surface));
}

.user-action-menu__icon {
  color: var(--theme-page-muted);
}

.user-action-menu__icon--success {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.user-action-menu__icon--warning {
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
}

.user-action-menu__divider {
  border-top: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}
</style>
