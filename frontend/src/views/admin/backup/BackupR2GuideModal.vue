<template>
  <teleport to="body">
    <transition name="modal">
      <div
        v-if="show"
        class="fixed inset-0 z-50 flex items-center justify-center p-4"
        @mousedown.self="emit('close')"
      >
        <div class="fixed inset-0 bg-black/50" @click="emit('close')"></div>
        <div class="relative max-h-[85vh] w-full max-w-2xl overflow-y-auto rounded-xl bg-white p-6 shadow-2xl dark:bg-dark-800">
          <button
            type="button"
            class="absolute right-4 top-4 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"
            @click="emit('close')"
          >
            <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>

          <h2 class="mb-4 text-lg font-bold text-gray-900 dark:text-white">
            {{ t('admin.backup.r2Guide.title') }}
          </h2>
          <p class="mb-4 text-sm text-gray-500 dark:text-gray-400">
            {{ t('admin.backup.r2Guide.intro') }}
          </p>

          <div class="mb-5">
            <h3 class="mb-2 flex items-center gap-2 text-sm font-semibold text-gray-900 dark:text-white">
              <span class="flex h-6 w-6 items-center justify-center rounded-full bg-primary-100 text-xs font-bold text-primary-700 dark:bg-primary-900/40 dark:text-primary-300">1</span>
              {{ t('admin.backup.r2Guide.step1.title') }}
            </h3>
            <ol class="ml-8 list-decimal space-y-1 text-sm text-gray-600 dark:text-gray-300">
              <li>{{ t('admin.backup.r2Guide.step1.line1') }}</li>
              <li>{{ t('admin.backup.r2Guide.step1.line2') }}</li>
              <li>{{ t('admin.backup.r2Guide.step1.line3') }}</li>
            </ol>
          </div>

          <div class="mb-5">
            <h3 class="mb-2 flex items-center gap-2 text-sm font-semibold text-gray-900 dark:text-white">
              <span class="flex h-6 w-6 items-center justify-center rounded-full bg-primary-100 text-xs font-bold text-primary-700 dark:bg-primary-900/40 dark:text-primary-300">2</span>
              {{ t('admin.backup.r2Guide.step2.title') }}
            </h3>
            <ol class="ml-8 list-decimal space-y-1 text-sm text-gray-600 dark:text-gray-300">
              <li>{{ t('admin.backup.r2Guide.step2.line1') }}</li>
              <li>{{ t('admin.backup.r2Guide.step2.line2') }}</li>
              <li>{{ t('admin.backup.r2Guide.step2.line3') }}</li>
              <li>{{ t('admin.backup.r2Guide.step2.line4') }}</li>
            </ol>
            <div class="mt-2 rounded-lg bg-amber-50 p-3 text-xs text-amber-700 dark:bg-amber-900/20 dark:text-amber-300">
              {{ t('admin.backup.r2Guide.step2.warning') }}
            </div>
          </div>

          <div class="mb-5">
            <h3 class="mb-2 flex items-center gap-2 text-sm font-semibold text-gray-900 dark:text-white">
              <span class="flex h-6 w-6 items-center justify-center rounded-full bg-primary-100 text-xs font-bold text-primary-700 dark:bg-primary-900/40 dark:text-primary-300">3</span>
              {{ t('admin.backup.r2Guide.step3.title') }}
            </h3>
            <p class="ml-8 text-sm text-gray-600 dark:text-gray-300">
              {{ t('admin.backup.r2Guide.step3.desc') }}
            </p>
            <code class="ml-8 mt-1 block rounded bg-gray-100 px-3 py-2 text-xs text-gray-800 dark:bg-dark-700 dark:text-gray-200">
              https://&lt;{{ t('admin.backup.r2Guide.step3.accountId') }}&gt;.r2.cloudflarestorage.com
            </code>
          </div>

          <div class="mb-5">
            <h3 class="mb-2 flex items-center gap-2 text-sm font-semibold text-gray-900 dark:text-white">
              <span class="flex h-6 w-6 items-center justify-center rounded-full bg-primary-100 text-xs font-bold text-primary-700 dark:bg-primary-900/40 dark:text-primary-300">4</span>
              {{ t('admin.backup.r2Guide.step4.title') }}
            </h3>
            <div class="ml-8 overflow-hidden rounded-lg border border-gray-200 dark:border-dark-600">
              <table class="w-full text-sm">
                <tbody>
                  <tr
                    v-for="(row, index) in r2ConfigRows"
                    :key="index"
                    class="border-b border-gray-100 dark:border-dark-700 last:border-0"
                  >
                    <td class="whitespace-nowrap bg-gray-50 px-3 py-2 font-medium text-gray-700 dark:bg-dark-700 dark:text-gray-300">
                      {{ row.field }}
                    </td>
                    <td class="px-3 py-2 text-gray-600 dark:text-gray-400">
                      <code class="text-xs">{{ row.value }}</code>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <div class="rounded-lg bg-green-50 p-3 text-xs text-green-700 dark:bg-green-900/20 dark:text-green-300">
            {{ t('admin.backup.r2Guide.freeTier') }}
          </div>

          <div class="mt-4 text-right">
            <button type="button" class="btn btn-primary btn-sm" @click="emit('close')">
              {{ t('common.close') }}
            </button>
          </div>
        </div>
      </div>
    </transition>
  </teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { buildBackupR2ConfigRows } from '../backupView'

defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

const { t } = useI18n()
const r2ConfigRows = computed(() => buildBackupR2ConfigRows(t))
</script>

<style scoped>
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}
.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
</style>
