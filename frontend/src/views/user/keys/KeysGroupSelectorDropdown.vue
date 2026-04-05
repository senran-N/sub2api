<template>
  <Teleport to="body">
    <div v-if="show && position">
      <div class="fixed inset-0 z-[100000019]" @click="emit('close')"></div>
      <div
        class="animate-in fade-in slide-in-from-top-2 fixed z-[100000020] w-[calc(100vw-2rem)] overflow-hidden rounded-xl bg-white shadow-lg ring-1 ring-black/5 duration-200 dark:bg-dark-800 dark:ring-white/10 sm:w-max sm:min-w-[380px]"
        style="pointer-events: auto !important;"
        :style="{
          top: position.top !== undefined ? `${position.top}px` : undefined,
          bottom: position.bottom !== undefined ? `${position.bottom}px` : undefined,
          left: `${position.left}px`
        }"
      >
        <div class="border-b border-gray-100 p-2 dark:border-dark-700">
          <div class="relative">
            <svg
              class="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400"
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
              class="w-full rounded-lg border border-gray-200 bg-gray-50 py-1.5 pl-8 pr-3 text-sm text-gray-900 placeholder-gray-400 outline-none focus:border-primary-300 focus:ring-1 focus:ring-primary-300 dark:border-dark-600 dark:bg-dark-700 dark:text-white dark:placeholder-gray-500 dark:focus:border-primary-600 dark:focus:ring-primary-600"
              :placeholder="t('keys.searchGroup')"
              @click.stop
              @input="emit('update:searchQuery', ($event.target as HTMLInputElement).value)"
            />
          </div>
        </div>

        <div class="max-h-80 overflow-y-auto p-1.5">
          <button
            v-for="option in options"
            :key="option.value"
            class="flex w-full items-center justify-between rounded-lg border-b border-gray-100 px-3 py-2.5 text-sm transition-colors last:border-0 dark:border-dark-700"
            :class="
              selectedGroupId === option.value
                ? 'bg-primary-50 dark:bg-primary-900/20'
                : 'hover:bg-gray-100 dark:hover:bg-dark-700'
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

          <div v-if="options.length === 0" class="py-4 text-center text-sm text-gray-400 dark:text-gray-500">
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
