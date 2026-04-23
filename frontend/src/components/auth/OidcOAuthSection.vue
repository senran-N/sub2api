<template>
  <div class="space-y-4">
    <button type="button" :disabled="disabled" class="btn btn-secondary w-full" @click="startLogin">
      <svg
        class="oidc-oauth-section__icon mr-2"
        viewBox="0 0 24 24"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        width="1em"
        height="1em"
        aria-hidden="true"
      >
        <path d="M12 3l7 4v5c0 4.25-2.5 8.16-6.4 9.83L12 22l-.6-.17C7.5 20.16 5 16.25 5 12V7l7-4z" class="oidc-oauth-section__shield"/>
        <path d="M9 12.5l2 2 4-4" class="oidc-oauth-section__check"/>
      </svg>
      {{ buttonLabel }}
    </button>

    <div v-if="props.showDivider !== false" class="flex items-center gap-3">
      <div class="oidc-oauth-section__divider h-px flex-1"></div>
      <span class="theme-text-muted text-xs">
        {{ t('auth.oidc.orContinue') }}
      </span>
      <div class="oidc-oauth-section__divider h-px flex-1"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'

const props = defineProps<{
  disabled?: boolean
  providerName?: string
  showDivider?: boolean
}>()

const route = useRoute()
const { t } = useI18n()

const buttonLabel = computed(() => {
  const providerName = props.providerName?.trim()
  if (providerName) {
    return t('auth.oidc.signInNamed', { provider: providerName })
  }
  return t('auth.oidc.signIn')
})

function startLogin(): void {
  const redirectTo = (route.query.redirect as string) || '/dashboard'
  const apiBase = (import.meta.env.VITE_API_BASE_URL as string | undefined) || '/api/v1'
  const normalized = apiBase.replace(/\/$/, '')
  const startURL = `${normalized}/auth/oauth/oidc/start?redirect=${encodeURIComponent(redirectTo)}`
  window.location.href = startURL
}
</script>

<style scoped>
.oidc-oauth-section__icon {
  width: 20px;
  height: 20px;
}

.oidc-oauth-section__shield {
  fill: color-mix(in srgb, rgb(var(--theme-primary-rgb)) 16%, var(--theme-surface));
  stroke: color-mix(in srgb, rgb(var(--theme-primary-rgb)) 78%, var(--theme-page-text));
  stroke-width: 1.5;
}

.oidc-oauth-section__check {
  stroke: color-mix(in srgb, rgb(var(--theme-primary-rgb)) 82%, white);
  stroke-linecap: round;
  stroke-linejoin: round;
  stroke-width: 2;
}

.oidc-oauth-section__divider {
  background: color-mix(in srgb, var(--theme-card-border) 82%, transparent);
}
</style>
