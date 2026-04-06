<template>
  <div class="version-badge relative">
    <!-- Admin: Full version badge with dropdown -->
    <template v-if="isAdmin">
      <button
        @click="toggleDropdown"
        class="version-badge__trigger"
        :class="[
          hasUpdate ? 'version-badge__trigger--update' : 'version-badge__trigger--idle'
        ]"
        :title="hasUpdate ? t('version.updateAvailable') : t('version.upToDate')"
      >
        <span v-if="currentVersion" class="font-medium">v{{ currentVersion }}</span>
        <span
          v-else
          class="version-badge__trigger-skeleton h-3 w-12 animate-pulse font-medium"
        ></span>
        <!-- Update indicator -->
        <span v-if="hasUpdate" class="version-badge__indicator">
          <span
            class="version-badge__indicator-ping"
          ></span>
          <span class="version-badge__indicator-dot"></span>
        </span>
      </button>

      <!-- Dropdown -->
      <transition name="dropdown">
        <div
          v-if="dropdownOpen"
          ref="dropdownRef"
          class="version-badge__panel absolute left-0 z-50 mt-2 overflow-hidden"
        >
          <!-- Header with refresh button -->
          <div
            class="version-badge__panel-header"
          >
            <span class="version-badge__panel-label text-sm font-medium">{{
              t('version.currentVersion')
            }}</span>
            <button
              @click="refreshVersion(true)"
              class="version-badge__refresh"
              :disabled="loading"
              :title="t('version.refresh')"
            >
              <Icon
                name="refresh"
                size="sm"
                :stroke-width="2"
                :class="{ 'animate-spin': loading }"
              />
            </button>
          </div>

          <div class="version-badge__panel-body">
            <!-- Loading state -->
            <div v-if="loading" class="version-badge__loading-state">
              <svg class="version-badge__loading-spinner h-6 w-6 animate-spin" fill="none" viewBox="0 0 24 24">
                <circle
                  class="opacity-25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  stroke-width="4"
                ></circle>
                <path
                  class="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                ></path>
              </svg>
            </div>

            <!-- Content -->
            <template v-else>
              <!-- Version display - centered and prominent -->
              <div class="mb-4 text-center">
                <div class="inline-flex items-center gap-2">
                  <span
                    v-if="currentVersion"
                    class="version-badge__version-value font-bold"
                    >v{{ currentVersion }}</span
                  >
                  <span v-else class="version-badge__version-placeholder font-bold">--</span>
                  <!-- Show check mark when up to date -->
                  <span
                    v-if="!hasUpdate"
                    class="version-badge__check-badge flex h-5 w-5 items-center justify-center"
                  >
                    <svg
                      class="version-badge__check-icon h-3 w-3"
                      fill="currentColor"
                      viewBox="0 0 20 20"
                    >
                      <path
                        fill-rule="evenodd"
                        d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                        clip-rule="evenodd"
                      />
                    </svg>
                  </span>
                </div>
                <p class="version-badge__version-meta mt-1 text-xs">
                  {{
                    hasUpdate
                      ? t('version.latestVersion') + ': v' + latestVersion
                      : t('version.upToDate')
                  }}
                </p>
              </div>

              <!-- Priority 1: Update error (must check before hasUpdate) -->
              <div v-if="updateError" class="space-y-2">
                <div
                  class="version-badge__status-card version-badge__status-card--danger"
                >
                  <div
                    class="version-badge__status-icon"
                  >
                    <Icon
                      name="x"
                      size="sm"
                      :stroke-width="2"
                      class="version-badge__status-symbol"
                    />
                  </div>
                  <div class="version-badge__status-content">
                    <p class="version-badge__status-title text-sm font-medium">
                      {{ t('version.updateFailed') }}
                    </p>
                    <p class="version-badge__status-text truncate text-xs">
                      {{ updateError }}
                    </p>
                  </div>
                </div>

                <!-- Retry button -->
                <button
                  @click="handleUpdate"
                  :disabled="updating"
                  class="version-badge__action version-badge__action--danger"
                >
                  {{ t('version.retry') }}
                </button>
              </div>

              <!-- Priority 2: Update success - need restart -->
              <div v-else-if="updateSuccess && needRestart" class="space-y-2">
                <div
                  class="version-badge__status-card version-badge__status-card--success"
                >
                  <div
                    class="version-badge__status-icon"
                  >
                    <svg
                      class="version-badge__status-symbol h-4 w-4"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      stroke-width="2"
                    >
                      <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                    </svg>
                  </div>
                  <div class="version-badge__status-content">
                    <p class="version-badge__status-title text-sm font-medium">
                      {{ t('version.updateComplete') }}
                    </p>
                    <p class="version-badge__status-text text-xs">
                      {{ t('version.restartRequired') }}
                    </p>
                  </div>
                </div>

                <!-- Restart button with countdown -->
                <button
                  @click="handleRestart"
                  :disabled="restarting"
                  class="version-badge__action version-badge__action--success"
                >
                  <svg
                    v-if="restarting"
                    class="h-4 w-4 animate-spin"
                    fill="none"
                    viewBox="0 0 24 24"
                  >
                    <circle
                      class="opacity-25"
                      cx="12"
                      cy="12"
                      r="10"
                      stroke="currentColor"
                      stroke-width="4"
                    ></circle>
                    <path
                      class="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                    ></path>
                  </svg>
                  <svg
                    v-else
                    class="h-4 w-4"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    stroke-width="2"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                    />
                  </svg>
                  <template v-if="restarting">
                    <span>{{ t('version.restarting') }}</span>
                    <span v-if="restartCountdown > 0" class="tabular-nums"
                      >({{ restartCountdown }}s)</span
                    >
                  </template>
                  <span v-else>{{ t('version.restartNow') }}</span>
                </button>
              </div>

              <!-- Priority 3: Update available for source build - show git pull hint -->
              <div v-else-if="hasUpdate && !isReleaseBuild" class="space-y-2">
                <a
                  v-if="releaseInfo?.html_url && releaseInfo.html_url !== '#'"
                  :href="releaseInfo.html_url"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="version-badge__status-card version-badge__status-card--warning version-badge__status-card--interactive group"
                >
                  <div
                    class="version-badge__status-icon"
                  >
                    <Icon
                      name="download"
                      size="sm"
                      :stroke-width="2"
                      class="version-badge__status-symbol"
                    />
                  </div>
                  <div class="version-badge__status-content">
                    <p class="version-badge__status-title text-sm font-medium">
                      {{ t('version.updateAvailable') }}
                    </p>
                    <p class="version-badge__status-text text-xs">
                      v{{ latestVersion }}
                    </p>
                  </div>
                  <svg
                    class="version-badge__arrow h-4 w-4 transition-transform group-hover:translate-x-0.5"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    stroke-width="2"
                  >
                    <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
                  </svg>
                </a>
                <!-- Source build hint -->
                <div
                  class="version-badge__status-card version-badge__status-card--info version-badge__status-card--compact"
                >
                  <svg
                    class="version-badge__status-symbol h-3.5 w-3.5 flex-shrink-0"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    stroke-width="2"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                    />
                  </svg>
                  <p class="version-badge__status-text text-xs">
                    {{ t('version.sourceModeHint') }}
                  </p>
                </div>
              </div>

              <!-- Priority 4: Update available for release build - show update button -->
              <div v-else-if="hasUpdate && isReleaseBuild" class="space-y-2">
                <!-- Update info card -->
                <div
                  class="version-badge__status-card version-badge__status-card--warning"
                >
                  <div
                    class="version-badge__status-icon"
                  >
                    <Icon
                      name="download"
                      size="sm"
                      :stroke-width="2"
                      class="version-badge__status-symbol"
                    />
                  </div>
                  <div class="version-badge__status-content">
                    <p class="version-badge__status-title text-sm font-medium">
                      {{ t('version.updateAvailable') }}
                    </p>
                    <p class="version-badge__status-text text-xs">
                      v{{ latestVersion }}
                    </p>
                  </div>
                </div>

                <!-- Update button -->
                <button
                  @click="handleUpdate"
                  :disabled="updating"
                  class="version-badge__action version-badge__action--primary"
                >
                  <svg v-if="updating" class="h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24">
                    <circle
                      class="opacity-25"
                      cx="12"
                      cy="12"
                      r="10"
                      stroke="currentColor"
                      stroke-width="4"
                    ></circle>
                    <path
                      class="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                    ></path>
                  </svg>
                  <Icon v-else name="download" size="sm" :stroke-width="2" />
                  {{ updating ? t('version.updating') : t('version.updateNow') }}
                </button>

                <!-- View release link -->
                <a
                  v-if="releaseInfo?.html_url && releaseInfo.html_url !== '#'"
                  :href="releaseInfo.html_url"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="version-badge__link version-badge__link--inline"
                >
                  {{ t('version.viewChangelog') }}
                  <Icon name="externalLink" size="xs" :stroke-width="2" />
                </a>
              </div>

              <!-- Priority 5: Up to date - show GitHub link -->
              <a
                v-else-if="releaseInfo?.html_url && releaseInfo.html_url !== '#'"
                :href="releaseInfo.html_url"
                target="_blank"
                rel="noopener noreferrer"
                class="version-badge__link version-badge__link--compact"
              >
                <svg class="h-4 w-4" fill="currentColor" viewBox="0 0 24 24">
                  <path
                    fill-rule="evenodd"
                    clip-rule="evenodd"
                    d="M12 2C6.477 2 2 6.477 2 12c0 4.42 2.865 8.17 6.839 9.49.5.092.682-.217.682-.482 0-.237-.008-.866-.013-1.7-2.782.604-3.369-1.34-3.369-1.34-.454-1.156-1.11-1.464-1.11-1.464-.908-.62.069-.608.069-.608 1.003.07 1.531 1.03 1.531 1.03.892 1.529 2.341 1.087 2.91.831.092-.646.35-1.086.636-1.336-2.22-.253-4.555-1.11-4.555-4.943 0-1.091.39-1.984 1.029-2.683-.103-.253-.446-1.27.098-2.647 0 0 .84-.269 2.75 1.025A9.578 9.578 0 0112 6.836c.85.004 1.705.114 2.504.336 1.909-1.294 2.747-1.025 2.747-1.025.546 1.377.203 2.394.1 2.647.64.699 1.028 1.592 1.028 2.683 0 3.842-2.339 4.687-4.566 4.935.359.309.678.919.678 1.852 0 1.336-.012 2.415-.012 2.743 0 .267.18.578.688.48C19.138 20.167 22 16.418 22 12c0-5.523-4.477-10-10-10z"
                  />
                </svg>
                {{ t('version.viewRelease') }}
              </a>
            </template>
          </div>
        </div>
      </transition>
    </template>

    <!-- Non-admin: Simple static version text -->
    <span v-else-if="version" class="version-badge__text text-xs">
      v{{ version }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import { performUpdate, restartService } from '@/api/admin/system'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

const props = defineProps<{
  version?: string
}>()

const authStore = useAuthStore()
const appStore = useAppStore()

const isAdmin = computed(() => authStore.isAdmin)

const dropdownOpen = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)

// Use store's cached version state
const loading = computed(() => appStore.versionLoading)
const currentVersion = computed(() => appStore.currentVersion || props.version || '')
const latestVersion = computed(() => appStore.latestVersion)
const hasUpdate = computed(() => appStore.hasUpdate)
const releaseInfo = computed(() => appStore.releaseInfo)
const buildType = computed(() => appStore.buildType)

// Update process states (local to this component)
const updating = ref(false)
const restarting = ref(false)
const needRestart = ref(false)
const updateError = ref('')
const updateSuccess = ref(false)
const restartCountdown = ref(0)

// Only show update check for release builds (binary/docker deployment)
const isReleaseBuild = computed(() => buildType.value === 'release')

function toggleDropdown() {
  dropdownOpen.value = !dropdownOpen.value
}

function closeDropdown() {
  dropdownOpen.value = false
}

async function refreshVersion(force = true) {
  if (!isAdmin.value) return

  // Reset update states when refreshing
  updateError.value = ''
  updateSuccess.value = false
  needRestart.value = false

  await appStore.fetchVersion(force)
}

async function handleUpdate() {
  if (updating.value) return

  updating.value = true
  updateError.value = ''
  updateSuccess.value = false

  try {
    const result = await performUpdate()
    updateSuccess.value = true
    needRestart.value = result.need_restart
    // Clear version cache to reflect update completed
    appStore.clearVersionCache()
  } catch (error: unknown) {
    const err = error as { response?: { data?: { message?: string } }; message?: string }
    updateError.value = err.response?.data?.message || err.message || t('version.updateFailed')
  } finally {
    updating.value = false
  }
}

async function handleRestart() {
  if (restarting.value) return

  restarting.value = true
  restartCountdown.value = 8

  try {
    await restartService()
    // Service will restart, page will reload automatically or show disconnected
  } catch (error) {
    // Expected - connection will be lost during restart
    console.log('Service restarting...')
  }

  // Start countdown
  const countdownInterval = setInterval(() => {
    restartCountdown.value--
    if (restartCountdown.value <= 0) {
      clearInterval(countdownInterval)
      // Try to check if service is back before reload
      checkServiceAndReload()
    }
  }, 1000)
}

async function checkServiceAndReload() {
  const maxRetries = 5
  const retryDelay = 1000

  for (let i = 0; i < maxRetries; i++) {
    try {
      const response = await fetch('/health', {
        method: 'GET',
        cache: 'no-cache'
      })
      if (response.ok) {
        // Service is back, reload page
        window.location.reload()
        return
      }
    } catch {
      // Service not ready yet
    }

    if (i < maxRetries - 1) {
      await new Promise((resolve) => setTimeout(resolve, retryDelay))
    }
  }

  // After retries, reload anyway
  window.location.reload()
}

function handleClickOutside(event: MouseEvent) {
  const target = event.target as Node
  const button = (event.target as Element).closest('button')
  if (dropdownRef.value && !dropdownRef.value.contains(target) && !button?.contains(target)) {
    closeDropdown()
  }
}

onMounted(() => {
  if (isAdmin.value) {
    // Use cached version if available, otherwise fetch
    appStore.fetchVersion(false)
  }
  document.addEventListener('click', handleClickOutside)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.dropdown-enter-active,
.dropdown-leave-active {
  transition: all 0.2s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: scale(0.95) translateY(-4px);
}

.version-badge {
  --version-badge-danger-rgb: var(--theme-danger-rgb);
  --version-badge-success-rgb: var(--theme-success-rgb);
  --version-badge-warning-rgb: var(--theme-warning-rgb);
  --version-badge-info-rgb: var(--theme-info-rgb);
}

.version-badge__trigger,
.version-badge__panel,
.version-badge__status-card,
.version-badge__action,
.version-badge__refresh {
  border-radius: var(--theme-version-panel-radius);
}

.version-badge__trigger {
  display: flex;
  align-items: center;
  gap: calc(var(--theme-table-layout-gap) * 0.375);
  padding: calc(var(--theme-button-padding-y) * 0.4) calc(var(--theme-button-padding-x) * 0.4);
  font-size: 0.75rem;
  transition: background 0.2s ease, color 0.2s ease, border-color 0.2s ease;
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 72%, transparent);
}

.version-badge__trigger--idle {
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, transparent);
  color: var(--theme-page-muted);
}

.version-badge__trigger--idle:hover {
  background: color-mix(in srgb, var(--theme-button-secondary-hover-bg) 92%, transparent);
  color: var(--theme-page-text);
}

.version-badge__trigger--update {
  background: color-mix(in srgb, rgb(var(--version-badge-warning-rgb)) 14%, var(--theme-surface));
  color: rgb(var(--version-badge-warning-rgb));
}

.version-badge__trigger--update:hover {
  background: color-mix(in srgb, rgb(var(--version-badge-warning-rgb)) 20%, var(--theme-surface));
}

.version-badge__trigger-skeleton {
  border-radius: 9999px;
  background: color-mix(in srgb, var(--theme-page-border) 78%, transparent);
}

.version-badge__indicator-ping {
  position: absolute;
  display: inline-flex;
  width: 100%;
  height: 100%;
  border-radius: 999px;
  opacity: 0.75;
  background: color-mix(in srgb, rgb(var(--version-badge-warning-rgb)) 72%, transparent);
}

.version-badge__indicator {
  position: relative;
  display: flex;
  width: 0.5rem;
  height: 0.5rem;
}

.version-badge__indicator-dot {
  position: relative;
  display: inline-flex;
  width: 0.5rem;
  height: 0.5rem;
  border-radius: 999px;
  background: rgb(var(--version-badge-warning-rgb));
}

.version-badge__panel {
  width: min(calc(100vw - 2rem), var(--theme-version-panel-width));
  border: 1px solid var(--theme-dropdown-border);
  background: var(--theme-dropdown-bg);
  box-shadow: var(--theme-dropdown-shadow);
}

.version-badge__panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--theme-table-cell-padding-y) var(--theme-table-cell-padding-x);
  border-bottom: 1px solid var(--theme-page-border);
}

.version-badge__panel-body {
  padding: var(--theme-table-mobile-card-padding);
}

.version-badge__panel-label,
.version-badge__text,
.version-badge__version-meta,
.version-badge__link {
  color: var(--theme-page-muted);
}

.version-badge__refresh {
  padding: 0.375rem;
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
}

.version-badge__refresh:hover {
  background: var(--theme-dropdown-item-hover-bg);
  color: var(--theme-page-text);
}

.version-badge__loading-spinner {
  color: var(--theme-accent);
}

.version-badge__loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: calc(var(--theme-table-mobile-empty-padding) * 0.5) 0;
}

.version-badge__version-value {
  font-family: var(--theme-version-value-font);
  font-size: var(--theme-version-value-size);
  color: var(--theme-page-text);
}

.version-badge__version-placeholder {
  font-family: var(--theme-version-value-font);
  font-size: var(--theme-version-value-size);
  color: color-mix(in srgb, var(--theme-page-muted) 60%, transparent);
}

.version-badge__check-badge {
  border-radius: var(--theme-version-icon-radius);
  background: color-mix(in srgb, rgb(var(--version-badge-success-rgb)) 14%, var(--theme-surface));
}

.version-badge__check-icon {
  color: rgb(var(--version-badge-success-rgb));
}

.version-badge__status-card {
  --version-badge-tone-rgb: var(--version-badge-info-rgb);
  display: flex;
  align-items: center;
  gap: var(--theme-table-layout-gap);
  padding: var(--theme-table-mobile-card-padding);
  border: 1px solid color-mix(in srgb, rgb(var(--version-badge-tone-rgb)) 26%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--version-badge-tone-rgb)) 10%, var(--theme-surface));
}

.version-badge__status-content {
  min-width: 0;
  flex: 1;
}

.version-badge__status-card--compact {
  gap: calc(var(--theme-table-layout-gap) * 0.5);
  padding: calc(var(--theme-table-mobile-card-padding) * 0.67);
}

.version-badge__status-card--danger {
  --version-badge-tone-rgb: var(--version-badge-danger-rgb);
}

.version-badge__status-card--success {
  --version-badge-tone-rgb: var(--version-badge-success-rgb);
}

.version-badge__status-card--warning {
  --version-badge-tone-rgb: var(--version-badge-warning-rgb);
}

.version-badge__status-card--info {
  --version-badge-tone-rgb: var(--version-badge-info-rgb);
}

.version-badge__status-card--interactive:hover {
  background: color-mix(in srgb, rgb(var(--version-badge-warning-rgb)) 14%, var(--theme-surface));
}

.version-badge__status-icon {
  display: flex;
  width: var(--theme-stat-icon-size);
  height: var(--theme-stat-icon-size);
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  border-radius: var(--theme-version-icon-radius);
  background: color-mix(in srgb, rgb(var(--version-badge-tone-rgb)) 16%, var(--theme-surface));
}

.version-badge__status-symbol,
.version-badge__arrow {
  color: rgb(var(--version-badge-tone-rgb));
}

.version-badge__status-title {
  color: color-mix(in srgb, rgb(var(--version-badge-tone-rgb)) 84%, var(--theme-page-text));
}

.version-badge__status-text {
  color: color-mix(in srgb, rgb(var(--version-badge-tone-rgb)) 76%, var(--theme-page-muted));
}

.version-badge__action {
  display: flex;
  width: 100%;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: calc(var(--theme-button-padding-y) * 0.8) var(--theme-button-padding-x);
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--theme-filled-text);
  transition:
    transform 0.2s ease,
    box-shadow 0.2s ease,
    background 0.2s ease;
}

.version-badge__action:hover {
  transform: translateY(-1px);
}

.version-badge__action--danger {
  background: linear-gradient(
    135deg,
    rgb(var(--version-badge-danger-rgb)),
    color-mix(in srgb, rgb(var(--version-badge-danger-rgb)) 72%, var(--theme-accent-strong))
  );
  box-shadow: 0 12px 28px color-mix(in srgb, rgb(var(--version-badge-danger-rgb)) 26%, transparent);
}

.version-badge__action--success {
  background: linear-gradient(
    135deg,
    rgb(var(--version-badge-success-rgb)),
    color-mix(in srgb, rgb(var(--version-badge-success-rgb)) 72%, var(--theme-accent-strong))
  );
  box-shadow: 0 12px 28px color-mix(in srgb, rgb(var(--version-badge-success-rgb)) 26%, transparent);
}

.version-badge__action--primary {
  background: linear-gradient(
    135deg,
    var(--theme-accent),
    color-mix(in srgb, var(--theme-accent-strong) 24%, var(--theme-accent) 76%)
  );
  box-shadow: 0 12px 28px color-mix(in srgb, var(--theme-accent) 28%, transparent);
}

.version-badge__link:hover {
  color: var(--theme-page-text);
}

.version-badge__link {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  transition: color 0.2s ease, background 0.2s ease;
}

.version-badge__link--inline {
  font-size: 0.75rem;
}

.version-badge__link--compact {
  padding: calc(var(--theme-button-padding-y) * 0.65) 0;
  font-size: 0.875rem;
}

.version-badge__trigger,
.version-badge__refresh,
.version-badge__action,
.version-badge__link {
  transition:
    background 0.2s ease,
    color 0.2s ease,
    border-color 0.2s ease,
    box-shadow 0.2s ease,
    transform 0.2s ease;
}

.line-clamp-3 {
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
