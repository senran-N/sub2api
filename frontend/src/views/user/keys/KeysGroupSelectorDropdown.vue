<template>
  <Teleport to="body">
    <div v-if="show && position">
      <div class="fixed inset-0 z-[100000019]" @click="emit('close')"></div>
      <div
        class="keys-group-selector-dropdown fixed z-[100000020] animate-in fade-in slide-in-from-top-2 duration-200"
        style="pointer-events: auto !important;"
        :style="{
          top: position.top !== undefined ? `${position.top}px` : undefined,
          bottom: position.bottom !== undefined ? `${position.bottom}px` : undefined,
          left: `${position.left}px`
        }"
      >
        <div class="keys-group-selector-dropdown__header">
          <div class="relative">
            <svg
              class="keys-group-selector-dropdown__search-icon"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              stroke-width="2"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
              />
            </svg>
            <input
              :value="searchQuery"
              type="text"
              class="keys-group-selector-dropdown__search"
              :placeholder="t('keys.searchGroup')"
              @click.stop
              @input="emit('update:searchQuery', ($event.target as HTMLInputElement).value)"
            />
          </div>
        </div>

        <div class="keys-group-selector-dropdown__list">
          <button
            v-for="option in options"
            :key="option.value"
            class="keys-group-selector-dropdown__option"
            :class="
              selectedGroupId === option.value
                ? 'keys-group-selector-dropdown__option--selected'
                : 'keys-group-selector-dropdown__option--interactive'
            "
            :title="option.description || undefined"
            @click="emit('select', option.value)"
          >
            <GroupOptionItem
              :name="option.label"
              :platform="option.platform"
              :subscription-type="option.subscriptionType"
              :rate-multiplier="option.rate"
              :user-rate-multiplier="option.userRate"
              :description="option.description"
              :selected="selectedGroupId === option.value"
            />
          </button>

          <div v-if="options.length === 0" class="keys-group-selector-dropdown__empty">
            {{ t('keys.noGroupFound') }}
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import GroupOptionItem from '@/components/common/GroupOptionItem.vue'
import type { UserKeyGroupOption } from './keysForm'
import type { KeysOverlayPosition } from './keysOverlays'

defineProps<{
  show: boolean
  position: KeysOverlayPosition | null
  searchQuery: string
  options: UserKeyGroupOption[]
  selectedGroupId: number | null
}>()

const emit = defineEmits<{
  close: []
  select: [groupId: number]
  'update:searchQuery': [value: string]
}>()

const { t } = useI18n()
</script>

<style scoped>
.keys-group-selector-dropdown {
  width: calc(100vw - 2rem);
  max-width: calc(100vw - 2rem);
  overflow: hidden;
  border-radius: calc(var(--theme-surface-radius) + 4px);
  background: var(--theme-dropdown-bg);
  box-shadow: var(--theme-dropdown-shadow);
  border: 1px solid color-mix(in srgb, var(--theme-dropdown-border) 88%, transparent);
}

.keys-group-selector-dropdown__header {
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
  padding: calc(var(--theme-user-api-keys-dropdown-padding) + 0.125rem);
}

.keys-group-selector-dropdown__header,
.keys-group-selector-dropdown__option {
  border-color: color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.keys-group-selector-dropdown__search-icon {
  position: absolute;
  left: calc(var(--theme-user-api-keys-dropdown-padding) + 0.375rem);
  top: 50%;
  height: 1rem;
  width: 1rem;
  transform: translateY(-50%);
}

.keys-group-selector-dropdown__search-icon,
.keys-group-selector-dropdown__empty {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
}

.keys-group-selector-dropdown__search {
  width: 100%;
  border-radius: calc(var(--theme-button-radius) + 2px);
  padding: calc(var(--theme-user-api-keys-dropdown-padding) + 0.125rem) calc(var(--theme-user-api-keys-dropdown-padding) + 0.5rem) calc(var(--theme-user-api-keys-dropdown-padding) + 0.125rem) calc(var(--theme-user-api-keys-dropdown-padding) * 2 + 1rem);
  font-size: 0.875rem;
  outline: none;
}

.keys-group-selector-dropdown__search {
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 84%, transparent);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  color: var(--theme-page-text);
}

.keys-group-selector-dropdown__search::placeholder {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
}

.keys-group-selector-dropdown__search:focus {
  border-color: var(--theme-accent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--theme-accent-soft) 88%, transparent);
}

.keys-group-selector-dropdown__list {
  max-height: calc(var(--theme-user-api-keys-dropdown-max-height) + 4rem);
  overflow-y: auto;
  padding: var(--theme-user-api-keys-dropdown-padding);
}

.keys-group-selector-dropdown__option {
  display: flex;
  width: 100%;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
  border-radius: calc(var(--theme-button-radius) + 2px);
  padding: calc(var(--theme-user-api-keys-dropdown-padding) + 0.25rem) calc(var(--theme-user-api-keys-dropdown-padding) + 0.5rem);
  font-size: 0.875rem;
  transition: background-color 0.2s ease, color 0.2s ease;
}

.keys-group-selector-dropdown__option:last-child {
  border-bottom: 0;
}

.keys-group-selector-dropdown__empty {
  padding: calc(var(--theme-user-api-keys-dropdown-padding) * 2) 0;
  text-align: center;
  font-size: 0.875rem;
}

.keys-group-selector-dropdown__option--selected {
  background: color-mix(in srgb, var(--theme-accent-soft) 82%, var(--theme-surface));
}

.keys-group-selector-dropdown__option--interactive:hover {
  background: var(--theme-dropdown-item-hover-bg);
}

@media (min-width: 640px) {
  .keys-group-selector-dropdown {
    width: max-content;
    min-width: max(var(--theme-user-api-keys-dropdown-width), 23.75rem);
    max-width: min(90vw, 32rem);
  }
}
</style>
