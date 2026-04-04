<template>
  <Teleport to="body">
    <Transition name="dm-drawer-mask">
      <div
        v-if="open"
        class="fixed inset-0 z-[54] bg-black/40 backdrop-blur-sm"
        @click="emit('close')"
      ></div>
    </Transition>

    <Transition name="dm-drawer-panel">
      <div
        v-if="open"
        class="fixed inset-y-0 right-0 z-[55] flex h-full w-full max-w-2xl flex-col border-l border-gray-200 bg-white shadow-2xl dark:border-dark-700 dark:bg-dark-900"
      >
        <div class="flex items-center justify-between border-b border-gray-200 px-4 py-3 dark:border-dark-700">
          <h4 class="text-sm font-semibold text-gray-900 dark:text-white">
            {{ creating ? t('admin.settings.soraS3.createTitle') : t('admin.settings.soraS3.editTitle') }}
          </h4>
          <button
            type="button"
            class="rounded p-1 text-gray-500 hover:bg-gray-100 hover:text-gray-700 dark:text-gray-400 dark:hover:bg-dark-800 dark:hover:text-gray-200"
            @click="emit('close')"
          >
            ✕
          </button>
        </div>

        <div class="flex-1 overflow-y-auto p-4">
          <div class="grid grid-cols-1 gap-3 md:grid-cols-2">
            <input
              v-model="form.profile_id"
              class="input w-full"
              :placeholder="t('admin.settings.soraS3.profileID')"
              :disabled="!creating"
            />
            <input
              v-model="form.name"
              class="input w-full"
              :placeholder="t('admin.settings.soraS3.profileName')"
            />
            <label class="inline-flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300 md:col-span-2">
              <input v-model="form.enabled" type="checkbox" />
              <span>{{ t('admin.settings.soraS3.enabled') }}</span>
            </label>
            <input v-model="form.endpoint" class="input w-full" :placeholder="t('admin.settings.soraS3.endpoint')" />
            <input v-model="form.region" class="input w-full" :placeholder="t('admin.settings.soraS3.region')" />
            <input v-model="form.bucket" class="input w-full" :placeholder="t('admin.settings.soraS3.bucket')" />
            <input v-model="form.prefix" class="input w-full" :placeholder="t('admin.settings.soraS3.prefix')" />
            <input v-model="form.access_key_id" class="input w-full" :placeholder="t('admin.settings.soraS3.accessKeyId')" />
            <input
              v-model="form.secret_access_key"
              type="password"
              class="input w-full"
              :placeholder="form.secret_access_key_configured ? t('admin.settings.soraS3.secretConfigured') : t('admin.settings.soraS3.secretAccessKey')"
            />
            <input v-model="form.cdn_url" class="input w-full" :placeholder="t('admin.settings.soraS3.cdnUrl')" />
            <div>
              <input
                v-model.number="form.default_storage_quota_gb"
                type="number"
                min="0"
                step="0.1"
                class="input w-full"
                :placeholder="t('admin.settings.soraS3.defaultQuota')"
              />
              <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                {{ t('admin.settings.soraS3.defaultQuotaHint') }}
              </p>
            </div>
            <label class="inline-flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
              <input v-model="form.force_path_style" type="checkbox" />
              <span>{{ t('admin.settings.soraS3.forcePathStyle') }}</span>
            </label>
            <label
              v-if="creating"
              class="inline-flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300 md:col-span-2"
            >
              <input v-model="form.set_active" type="checkbox" />
              <span>{{ t('admin.settings.soraS3.setActive') }}</span>
            </label>
          </div>
        </div>

        <div class="flex flex-wrap justify-end gap-2 border-t border-gray-200 p-4 dark:border-dark-700">
          <button type="button" class="btn btn-secondary btn-sm" @click="emit('close')">
            {{ t('common.cancel') }}
          </button>
          <button
            type="button"
            class="btn btn-secondary btn-sm"
            :disabled="testing || !form.enabled"
            @click="emit('test')"
          >
            {{ testing ? t('common.loading') : t('admin.settings.soraS3.testConnection') }}
          </button>
          <button type="button" class="btn btn-primary btn-sm" :disabled="saving" @click="emit('save')">
            {{ saving ? t('common.loading') : t('admin.settings.soraS3.saveProfile') }}
          </button>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { SoraS3ProfileForm } from '../dataManagementView'

defineProps<{
  open: boolean
  creating: boolean
  saving: boolean
  testing: boolean
  form: SoraS3ProfileForm
}>()

const emit = defineEmits<{
  close: []
  test: []
  save: []
}>()

const { t } = useI18n()
</script>

<style scoped>
.dm-drawer-mask-enter-active,
.dm-drawer-mask-leave-active {
  transition: opacity 0.2s ease;
}

.dm-drawer-mask-enter-from,
.dm-drawer-mask-leave-to {
  opacity: 0;
}

.dm-drawer-panel-enter-active,
.dm-drawer-panel-leave-active {
  transition:
    transform 0.24s cubic-bezier(0.22, 1, 0.36, 1),
    opacity 0.2s ease;
}

.dm-drawer-panel-enter-from,
.dm-drawer-panel-leave-to {
  opacity: 0.96;
  transform: translateX(100%);
}

@media (prefers-reduced-motion: reduce) {
  .dm-drawer-mask-enter-active,
  .dm-drawer-mask-leave-active,
  .dm-drawer-panel-enter-active,
  .dm-drawer-panel-leave-active {
    transition-duration: 0s;
  }
}
</style>
