<template>
  <Teleport to="body">
    <div v-if="show" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div class="fixed inset-0 bg-black/50" @click="emit('close')"></div>
      <div class="relative z-10 w-full max-w-lg rounded-xl bg-white shadow-xl dark:bg-dark-800">
        <div
          class="flex items-center justify-between border-b border-gray-200 px-5 py-4 dark:border-dark-600"
        >
          <div class="flex items-center gap-3">
            <div
              class="flex h-10 w-10 items-center justify-center rounded-full bg-green-100 dark:bg-green-900/30"
            >
              <svg
                class="h-5 w-5 text-green-600 dark:text-green-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M5 13l4 4L19 7"
                />
              </svg>
            </div>
            <div>
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">
                {{ t('admin.redeem.generatedSuccessfully') }}
              </h2>
              <p class="text-sm text-gray-500 dark:text-gray-400">
                {{ t('admin.redeem.codesCreated', { count }) }}
              </p>
            </div>
          </div>
          <button
            class="rounded-lg p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-dark-700 dark:hover:text-gray-300"
            @click="emit('close')"
          >
            <Icon name="x" size="md" :stroke-width="2" />
          </button>
        </div>

        <div class="p-5">
          <div class="relative">
            <textarea
              readonly
              :value="codesText"
              :style="{ height: textareaHeight }"
              class="w-full resize-none rounded-lg border border-gray-200 bg-gray-50 p-3 font-mono text-sm text-gray-800 focus:outline-none dark:border-dark-600 dark:bg-dark-700 dark:text-gray-200"
            ></textarea>
          </div>
        </div>

        <div
          class="flex justify-end gap-2 rounded-b-xl border-t border-gray-200 bg-gray-50 px-5 py-4 dark:border-dark-600 dark:bg-dark-700/50"
        >
          <button
            :class="[
              'btn flex items-center gap-2 transition-all',
              copiedAll ? 'btn-success' : 'btn-secondary'
            ]"
            @click="emit('copy')"
          >
            <Icon v-if="!copiedAll" name="copy" size="sm" :stroke-width="2" />
            <svg v-else class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M5 13l4 4L19 7"
              />
            </svg>
            {{ copiedAll ? t('admin.redeem.copied') : t('admin.redeem.copyAll') }}
          </button>
          <button class="btn btn-primary flex items-center gap-2" @click="emit('download')">
            <Icon name="download" size="sm" :stroke-width="2" />
            {{ t('admin.redeem.download') }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  show: boolean
  count: number
  codesText: string
  textareaHeight: string
  copiedAll: boolean
}>()

const emit = defineEmits<{
  close: []
  copy: []
  download: []
}>()

const { t } = useI18n()
</script>
