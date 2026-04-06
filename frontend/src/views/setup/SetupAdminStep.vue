<template>
  <div class="space-y-6">
    <div class="setup-step-header">
      <h2 class="setup-step-title">
        {{ t('setup.admin.title') }}
      </h2>
      <p class="setup-step-description">
        {{ t('setup.admin.description') }}
      </p>
    </div>

    <div>
      <label class="input-label">{{ t('setup.admin.email') }}</label>
      <input
        v-model="email"
        type="email"
        class="input"
        placeholder="admin@example.com"
      />
    </div>

    <div>
      <label class="input-label">{{ t('setup.admin.password') }}</label>
      <input
        v-model="password"
        type="password"
        class="input"
        :placeholder="t('setup.admin.passwordPlaceholder')"
      />
    </div>

    <div>
      <label class="input-label">{{ t('setup.admin.confirmPassword') }}</label>
      <input
        v-model="confirmPasswordModel"
        type="password"
        class="input"
        :placeholder="t('setup.admin.confirmPasswordPlaceholder')"
      />
      <p
        v-if="confirmPassword && admin.password !== confirmPassword"
        class="input-error-text"
      >
        {{ t('setup.admin.passwordMismatch') }}
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { AdminConfig } from '@/api/setup'

const props = defineProps<{
  admin: AdminConfig
  confirmPassword: string
}>()

const emit = defineEmits<{
  'update:admin': [value: AdminConfig]
  'update:confirm-password': [value: string]
}>()

const { t } = useI18n()

const updateAdmin = (patch: Partial<AdminConfig>) => {
  emit('update:admin', {
    ...props.admin,
    ...patch
  })
}

const email = computed({
  get: () => props.admin.email,
  set: (value: string) => updateAdmin({ email: value })
})

const password = computed({
  get: () => props.admin.password,
  set: (value: string) => updateAdmin({ password: value })
})

const confirmPasswordModel = computed({
  get: () => props.confirmPassword,
  set: (value: string) => emit('update:confirm-password', value)
})
</script>
