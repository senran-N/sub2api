<template>
  <div class="relative" ref="containerRef">
    <button
      type="button"
      @click="toggle"
      :disabled="disabled"
      :class="[
        'select-trigger',
        isOpen && 'select-trigger-open',
        disabled && 'select-trigger-disabled'
      ]"
    >
      <span class="select-value">
        {{ selectedLabel }}
      </span>
      <span class="select-icon">
        <Icon
          name="chevronDown"
          size="md"
          :class="['transition-transform duration-200', isOpen && 'rotate-180']"
        />
      </span>
    </button>

    <Transition name="select-dropdown">
      <div v-if="isOpen" class="select-dropdown">
        <!-- Search and Batch Test Header -->
        <div class="select-header">
          <div class="select-search">
            <Icon name="search" size="sm" class="select-search-icon" />
            <input
              ref="searchInputRef"
              v-model="searchQuery"
              type="text"
              :placeholder="t('admin.proxies.searchProxies')"
              class="select-search-input"
              @click.stop
            />
          </div>
          <button
            v-if="proxies.length > 0"
            type="button"
            @click.stop="handleBatchTest"
            :disabled="batchTesting"
            class="batch-test-btn"
            :title="t('admin.proxies.batchTest')"
          >
            <svg v-if="batchTesting" class="h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24">
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
            <Icon v-else name="play" size="sm" />
          </button>
        </div>

        <!-- Options list -->
        <div class="select-options">
          <!-- No Proxy option -->
          <div
            @click="selectOption(null)"
            :class="['select-option', modelValue === null && 'select-option-selected']"
          >
            <span class="select-option-label">{{ t('admin.accounts.noProxy') }}</span>
            <Icon v-if="modelValue === null" name="check" size="sm" class="select-option-check" />
          </div>

          <!-- Proxy options -->
          <div
            v-for="proxy in filteredProxies"
            :key="proxy.id"
            @click="selectOption(proxy.id)"
            :class="['select-option', modelValue === proxy.id && 'select-option-selected']"
          >
            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-2">
                <span class="truncate font-medium">{{ proxy.name }}</span>
                <!-- Account count badge -->
                <span
                  v-if="proxy.account_count !== undefined"
                  class="theme-chip theme-chip--compact theme-chip--neutral flex-shrink-0"
                >
                  {{ proxy.account_count }}
                </span>
                <!-- Test result badges -->
                <template v-if="testResults[proxy.id]">
                  <span
                    v-if="testResults[proxy.id].success"
                    class="theme-chip theme-chip--compact theme-chip--success flex-shrink-0"
                  >
                    <span v-if="testResults[proxy.id].country">{{
                      testResults[proxy.id].country
                    }}</span>
                    <span v-if="testResults[proxy.id].latency_ms"
                      >{{ testResults[proxy.id].latency_ms }}ms</span
                    >
                  </span>
                  <span
                    v-else
                    class="theme-chip theme-chip--compact theme-chip--danger flex-shrink-0"
                  >
                    {{ t('admin.proxies.testFailed') }}
                  </span>
                </template>
              </div>
              <div class="select-option-meta truncate text-xs">
                {{ proxy.protocol }}://{{ proxy.host }}:{{ proxy.port }}
              </div>
            </div>

            <!-- Individual test button -->
            <button
              type="button"
              @click.stop="handleTestProxy(proxy)"
              :disabled="testingProxyIds.has(proxy.id)"
              class="test-btn"
              :title="t('admin.proxies.testConnection')"
            >
              <svg
                v-if="testingProxyIds.has(proxy.id)"
                class="h-3.5 w-3.5 animate-spin"
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
              <Icon v-else name="play" size="xs" />
            </button>

            <Icon
              v-if="modelValue === proxy.id"
              name="check"
              size="sm"
              class="select-option-check flex-shrink-0"
            />
          </div>

          <!-- Empty state -->
          <div v-if="filteredProxies.length === 0 && searchQuery" class="select-empty">
            {{ t('common.noOptionsFound') }}
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import Icon from '@/components/icons/Icon.vue'
import type { Proxy } from '@/types'

const { t } = useI18n()

interface ProxyTestResult {
  success: boolean
  message: string
  latency_ms?: number
  ip_address?: string
  city?: string
  region?: string
  country?: string
}

interface Props {
  modelValue: number | null
  proxies: Proxy[]
  disabled?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false
})

const emit = defineEmits<{
  'update:modelValue': [value: number | null]
}>()

const isOpen = ref(false)
const searchQuery = ref('')
const containerRef = ref<HTMLElement | null>(null)
const searchInputRef = ref<HTMLInputElement | null>(null)

// Test state
const testResults = reactive<Record<number, ProxyTestResult>>({})
const testingProxyIds = reactive(new Set<number>())
const batchTesting = ref(false)

const selectedProxy = computed(() => {
  if (props.modelValue === null) return null
  return props.proxies.find((p) => p.id === props.modelValue) || null
})

const selectedLabel = computed(() => {
  if (!selectedProxy.value) {
    return t('admin.accounts.noProxy')
  }
  const proxy = selectedProxy.value
  return `${proxy.name} (${proxy.protocol}://${proxy.host}:${proxy.port})`
})

const filteredProxies = computed(() => {
  if (!searchQuery.value) {
    return props.proxies
  }
  const query = searchQuery.value.toLowerCase()
  return props.proxies.filter((proxy) => {
    const name = proxy.name.toLowerCase()
    const host = proxy.host.toLowerCase()
    return name.includes(query) || host.includes(query)
  })
})

const toggle = () => {
  if (props.disabled) return
  isOpen.value = !isOpen.value
  if (isOpen.value) {
    nextTick(() => {
      searchInputRef.value?.focus()
    })
  }
}

const selectOption = (value: number | null) => {
  emit('update:modelValue', value)
  isOpen.value = false
  searchQuery.value = ''
}

const handleTestProxy = async (proxy: Proxy) => {
  if (testingProxyIds.has(proxy.id)) return

  testingProxyIds.add(proxy.id)
  try {
    const result = await adminAPI.proxies.testProxy(proxy.id)
    testResults[proxy.id] = result
  } catch (error: any) {
    testResults[proxy.id] = {
      success: false,
      message: error.response?.data?.detail || 'Test failed'
    }
  } finally {
    testingProxyIds.delete(proxy.id)
  }
}

const handleBatchTest = async () => {
  if (batchTesting.value || props.proxies.length === 0) return

  batchTesting.value = true

  // Test all proxies in parallel
  const testPromises = props.proxies.map(async (proxy) => {
    testingProxyIds.add(proxy.id)
    try {
      const result = await adminAPI.proxies.testProxy(proxy.id)
      testResults[proxy.id] = result
    } catch (error: any) {
      testResults[proxy.id] = {
        success: false,
        message: error.response?.data?.detail || 'Test failed'
      }
    } finally {
      testingProxyIds.delete(proxy.id)
    }
  })

  await Promise.all(testPromises)
  batchTesting.value = false
}

const handleClickOutside = (event: MouseEvent) => {
  if (containerRef.value && !containerRef.value.contains(event.target as Node)) {
    isOpen.value = false
    searchQuery.value = ''
  }
}

const handleEscape = (event: KeyboardEvent) => {
  if (event.key === 'Escape' && isOpen.value) {
    isOpen.value = false
    searchQuery.value = ''
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('keydown', handleEscape)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('keydown', handleEscape)
})
</script>

<style scoped>
.select-trigger {
  @apply flex w-full items-center justify-between gap-2;
  padding: var(--theme-proxy-selector-trigger-padding-y)
    var(--theme-proxy-selector-trigger-padding-x);
  font-size: 0.875rem;
  @apply transition-all duration-200;
  @apply cursor-pointer;
  background: var(--theme-input-bg);
  border: 1px solid var(--theme-input-border);
  color: var(--theme-input-text);
  border-radius: var(--theme-select-radius);
}

.select-trigger:hover {
  border-color: color-mix(in srgb, var(--theme-input-border) 60%, var(--theme-page-text));
}

.select-trigger-open {
  border-color: var(--theme-accent);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--theme-accent) 18%, transparent);
}

.select-trigger-disabled {
  cursor: not-allowed;
  background: color-mix(in srgb, var(--theme-surface-soft) 86%, var(--theme-input-bg));
  opacity: 0.6;
}

.select-value {
  @apply flex-1 truncate text-left;
}

.select-icon {
  @apply flex-shrink-0;
  color: var(--theme-input-placeholder);
}

.select-dropdown {
  @apply absolute z-[100] mt-2 w-full;
  @apply overflow-hidden;
  background: var(--theme-dropdown-bg);
  border: 1px solid var(--theme-dropdown-border);
  box-shadow: var(--theme-dropdown-shadow);
  border-radius: var(--theme-select-panel-radius);
}

.select-header {
  @apply flex items-center gap-2;
  padding: var(--theme-proxy-selector-header-padding-y)
    var(--theme-proxy-selector-header-padding-x);
  border-bottom: 1px solid var(--theme-page-border);
}

.select-search {
  @apply flex flex-1 items-center gap-2;
}

.select-search-input {
  @apply flex-1 bg-transparent text-sm;
  @apply focus:outline-none;
  color: var(--theme-input-text);
}

.select-search-input::placeholder {
  color: var(--theme-input-placeholder);
}

.select-search-icon {
  color: var(--theme-input-placeholder);
}

.batch-test-btn {
  @apply flex-shrink-0;
  padding: var(--theme-proxy-selector-action-padding);
  @apply transition-colors disabled:cursor-not-allowed disabled:opacity-50;
  color: var(--theme-page-muted);
  border-radius: var(--theme-select-action-radius);
}

.batch-test-btn:hover {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 10%, var(--theme-surface));
}

.select-options {
  @apply overflow-y-auto;
  max-height: var(--theme-proxy-selector-options-max-height);
  padding-block: var(--theme-proxy-selector-options-padding-y);
}

.select-option {
  @apply flex items-center justify-between gap-2;
  padding: var(--theme-proxy-selector-option-padding-y)
    var(--theme-proxy-selector-option-padding-x);
  font-size: 0.875rem;
  @apply cursor-pointer transition-colors duration-150;
  color: var(--theme-page-text);
}

.select-option:hover {
  background: var(--theme-dropdown-item-hover-bg);
}

.select-option-selected {
  background: color-mix(in srgb, var(--theme-accent-soft) 78%, var(--theme-surface));
  color: var(--theme-accent);
}

.select-option-label {
  @apply truncate;
}

.select-empty {
  padding: var(--theme-proxy-selector-empty-padding-y)
    var(--theme-proxy-selector-empty-padding-x);
  @apply text-center text-sm;
  color: var(--theme-page-muted);
}

.test-btn {
  @apply flex-shrink-0;
  padding: var(--theme-proxy-selector-test-button-padding);
  @apply transition-colors disabled:cursor-not-allowed disabled:opacity-50;
  color: var(--theme-input-placeholder);
  border-radius: var(--theme-select-action-radius);
}

.test-btn:hover {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 10%, var(--theme-surface));
}

.select-option-meta {
  color: var(--theme-page-muted);
}

.select-option-check {
  color: var(--theme-accent);
}

/* Dropdown animation */
.select-dropdown-enter-active,
.select-dropdown-leave-active {
  transition: all 0.2s ease;
}

.select-dropdown-enter-from,
.select-dropdown-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
