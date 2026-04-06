<template>
  <Teleport to="body">
    <div v-if="show && position">
      <div class="fixed inset-0 z-[9998]" @click="emit('close')"></div>
      <div
        ref="menuRef"
        class="account-action-menu fixed z-[9999] overflow-hidden"
        :style="menuStyle"
        @click.stop
      >
        <div class="account-action-menu__body">
          <template v-if="account">
            <button @click="$emit('test', account); $emit('close')" class="account-action-menu__item">
              <Icon name="play" size="sm" class="account-action-menu__icon account-action-menu__icon--success" :stroke-width="2" />
              {{ t('admin.accounts.testConnection') }}
            </button>
            <button @click="$emit('stats', account); $emit('close')" class="account-action-menu__item">
              <Icon name="chart" size="sm" class="account-action-menu__icon account-action-menu__icon--info" />
              {{ t('admin.accounts.viewStats') }}
            </button>
            <button @click="$emit('schedule', account); $emit('close')" class="account-action-menu__item">
              <Icon name="clock" size="sm" class="account-action-menu__icon account-action-menu__icon--warning" />
              {{ t('admin.scheduledTests.schedule') }}
            </button>
            <template v-if="account.type === 'oauth' || account.type === 'setup-token'">
              <button @click="$emit('reauth', account); $emit('close')" class="account-action-menu__item account-action-menu__item--accent">
                <Icon name="link" size="sm" />
                {{ t('admin.accounts.reAuthorize') }}
              </button>
              <button @click="$emit('refresh-token', account); $emit('close')" class="account-action-menu__item account-action-menu__item--brand">
                <Icon name="refresh" size="sm" />
                {{ t('admin.accounts.refreshToken') }}
              </button>
            </template>
            <button v-if="supportsPrivacy" @click="$emit('set-privacy', account); $emit('close')" class="account-action-menu__item account-action-menu__item--success">
              <Icon name="shield" size="sm" />
              {{ t('admin.accounts.setPrivacy') }}
            </button>
            <div v-if="hasRecoverableState" class="account-action-menu__divider"></div>
            <button v-if="hasRecoverableState" @click="$emit('recover-state', account); $emit('close')" class="account-action-menu__item account-action-menu__item--success">
              <Icon name="sync" size="sm" />
              {{ t('admin.accounts.recoverState') }}
            </button>
            <button v-if="hasQuotaLimit" @click="$emit('reset-quota', account); $emit('close')" class="account-action-menu__item account-action-menu__item--info">
              <Icon name="refresh" size="sm" />
              {{ t('admin.accounts.resetQuota') }}
            </button>
          </template>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@/components/icons'
import { useDocumentThemeVersion } from '@/composables/useDocumentThemeVersion'
import type { Account } from '@/types'
import { clampFloatingPanelPosition, readThemePixelValue } from '@/utils/floatingPanel'

const props = defineProps<{ show: boolean; account: Account | null; position: { top: number; left: number } | null }>()
const emit = defineEmits(['close', 'test', 'stats', 'schedule', 'reauth', 'refresh-token', 'recover-state', 'reset-quota', 'set-privacy'])
const { t } = useI18n()
const themeVersion = useDocumentThemeVersion()
const menuRef = ref<HTMLElement | null>(null)
const menuStyle = ref<Record<string, string>>({})

const updateMenuPosition = () => {
  if (!props.show || !props.position) {
    menuStyle.value = {}
    return
  }

  const padding = readThemePixelValue('--theme-floating-panel-viewport-padding', 8)
  const panelWidth = menuRef.value?.offsetWidth ?? readThemePixelValue('--theme-account-action-menu-width', 208)
  const panelHeight = menuRef.value?.offsetHeight ?? readThemePixelValue('--theme-account-action-menu-estimated-height', 240)
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

const isRateLimited = computed(() => {
  if (props.account?.rate_limit_reset_at && new Date(props.account.rate_limit_reset_at) > new Date()) {
    return true
  }
  const modelLimits = (props.account?.extra as Record<string, unknown> | undefined)?.model_rate_limits as
    | Record<string, { rate_limit_reset_at: string }>
    | undefined
  if (modelLimits) {
    const now = new Date()
    return Object.values(modelLimits).some(info => new Date(info.rate_limit_reset_at) > now)
  }
  return false
})
const isOverloaded = computed(() => props.account?.overload_until && new Date(props.account.overload_until) > new Date())
const isTempUnschedulable = computed(() => props.account?.temp_unschedulable_until && new Date(props.account.temp_unschedulable_until) > new Date())
const hasRecoverableState = computed(() => {
  return props.account?.status === 'error' || Boolean(isRateLimited.value) || Boolean(isOverloaded.value) || Boolean(isTempUnschedulable.value)
})
const isAntigravityOAuth = computed(() => props.account?.platform === 'antigravity' && props.account?.type === 'oauth')
const isOpenAIOAuth = computed(() => props.account?.platform === 'openai' && props.account?.type === 'oauth')
const supportsPrivacy = computed(() => isAntigravityOAuth.value || isOpenAIOAuth.value)
const hasQuotaLimit = computed(() => {
  return (props.account?.type === 'apikey' || props.account?.type === 'bedrock') && (
    (props.account?.quota_limit ?? 0) > 0 ||
    (props.account?.quota_daily_limit ?? 0) > 0 ||
    (props.account?.quota_weekly_limit ?? 0) > 0
  )
})

const handleKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Escape') emit('close')
}

watch(
  () => props.show,
  (visible) => {
    if (visible) {
      window.addEventListener('keydown', handleKeydown)
    } else {
      window.removeEventListener('keydown', handleKeydown)
    }
  },
  { immediate: true }
)

watch(
  [() => props.show, () => props.position?.top, () => props.position?.left, themeVersion],
  async () => {
    if (!props.show || !props.position) {
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
  window.removeEventListener('keydown', handleKeydown)
  window.removeEventListener('resize', updateMenuPosition)
})
</script>

<style scoped>
.account-action-menu {
  width: min(
    var(--theme-account-action-menu-width),
    calc(100vw - (var(--theme-floating-panel-viewport-padding) + var(--theme-floating-panel-viewport-padding)))
  );
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 74%, transparent);
  border-radius: calc(var(--theme-surface-radius) + 4px);
  background: var(--theme-surface);
  box-shadow: var(--theme-dropdown-shadow);
}

.account-action-menu__body {
  padding: var(--theme-dropdown-padding-y) 0;
}

.account-action-menu__item {
  display: flex;
  width: 100%;
  align-items: center;
  gap: 0.5rem;
  padding: var(--theme-dropdown-item-padding-y) var(--theme-dropdown-item-padding-x);
  font-size: 0.875rem;
  color: var(--theme-page-text);
  transition:
    background-color 0.2s ease,
    color 0.2s ease;
}

.account-action-menu__item:hover {
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
}

.account-action-menu__item--accent {
  color: var(--theme-accent);
}

.account-action-menu__item--brand {
  color: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 84%, var(--theme-page-text));
}

.account-action-menu__item--success {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.account-action-menu__item--info {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
}

.account-action-menu__icon--success {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.account-action-menu__icon--info {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
}

.account-action-menu__icon--warning {
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
}

.account-action-menu__divider {
  margin: var(--theme-dropdown-padding-y) 0;
  border-top: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}
</style>
