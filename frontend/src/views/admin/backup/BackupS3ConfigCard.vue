<template>
  <div class="backup-s3-config-card card">
    <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
      <div>
        <h3 class="backup-s3-config-card__title text-base font-semibold">
          {{ t('admin.backup.s3.title') }}
        </h3>
        <p class="backup-s3-config-card__description mt-1 text-sm">
          {{ t('admin.backup.s3.descriptionPrefix') }}
          <button
            type="button"
            class="backup-s3-config-card__link underline"
            @click="emit('open-guide')"
          >
            Cloudflare R2
          </button>
          {{ t('admin.backup.s3.descriptionSuffix') }}
        </p>
      </div>
    </div>

    <div class="grid grid-cols-1 gap-3 md:grid-cols-2">
      <div>
        <label class="backup-s3-config-card__field-label mb-1 block text-xs font-medium">
          {{ t('admin.backup.s3.endpoint') }}
        </label>
        <input
          v-model="form.endpoint"
          class="input w-full"
          placeholder="https://<account_id>.r2.cloudflarestorage.com"
        />
      </div>
      <div>
        <label class="backup-s3-config-card__field-label mb-1 block text-xs font-medium">
          {{ t('admin.backup.s3.region') }}
        </label>
        <input v-model="form.region" class="input w-full" placeholder="auto" />
      </div>
      <div>
        <label class="backup-s3-config-card__field-label mb-1 block text-xs font-medium">
          {{ t('admin.backup.s3.bucket') }}
        </label>
        <input v-model="form.bucket" class="input w-full" />
      </div>
      <div>
        <label class="backup-s3-config-card__field-label mb-1 block text-xs font-medium">
          {{ t('admin.backup.s3.prefix') }}
        </label>
        <input v-model="form.prefix" class="input w-full" placeholder="backups/" />
      </div>
      <div>
        <label class="backup-s3-config-card__field-label mb-1 block text-xs font-medium">
          {{ t('admin.backup.s3.accessKeyId') }}
        </label>
        <input v-model="form.access_key_id" class="input w-full" />
      </div>
      <div>
        <label class="backup-s3-config-card__field-label mb-1 block text-xs font-medium">
          {{ t('admin.backup.s3.secretAccessKey') }}
        </label>
        <input
          v-model="form.secret_access_key"
          type="password"
          class="input w-full"
          :placeholder="secretConfigured ? t('admin.backup.s3.secretConfigured') : ''"
        />
      </div>
      <label class="backup-s3-config-card__checkbox inline-flex items-center gap-2 text-sm md:col-span-2">
        <input v-model="form.force_path_style" type="checkbox" />
        <span>{{ t('admin.backup.s3.forcePathStyle') }}</span>
      </label>
    </div>

    <div class="mt-4 flex flex-wrap gap-2">
      <button type="button" class="btn btn-secondary btn-sm" :disabled="testing" @click="emit('test')">
        {{ testing ? t('common.loading') : t('admin.backup.s3.testConnection') }}
      </button>
      <button type="button" class="btn btn-primary btn-sm" :disabled="saving" @click="emit('save')">
        {{ saving ? t('common.loading') : t('common.save') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { BackupS3Config } from '@/api/admin/backup'

defineProps<{
  form: BackupS3Config
  secretConfigured: boolean
  saving: boolean
  testing: boolean
}>()

const emit = defineEmits<{
  'open-guide': []
  test: []
  save: []
}>()

const { t } = useI18n()
</script>

<style scoped>
.backup-s3-config-card__title,
.backup-s3-config-card__field-label,
.backup-s3-config-card__checkbox {
  color: var(--theme-page-text);
}

.backup-s3-config-card__description {
  color: var(--theme-page-muted);
}

.backup-s3-config-card__link {
  color: var(--theme-accent);
}

.backup-s3-config-card__link:hover {
  color: color-mix(in srgb, var(--theme-accent) 82%, var(--theme-page-text));
}

.backup-s3-config-card {
  padding: var(--theme-settings-card-panel-padding);
}
</style>
