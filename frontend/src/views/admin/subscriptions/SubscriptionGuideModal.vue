<template>
  <teleport to="body">
    <transition name="modal">
      <div
        v-if="show"
        class="subscription-guide-modal fixed inset-0 z-50 flex items-center justify-center"
        @mousedown.self="emit('close')"
      >
        <div class="subscription-guide-modal__overlay fixed inset-0" @click="emit('close')"></div>
        <div class="subscription-guide-modal__panel relative w-full overflow-y-auto">
          <button
            type="button"
            class="subscription-guide-modal__close absolute"
            @click="emit('close')"
          >
            <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>

          <h2 class="subscription-guide-modal__title mb-4 text-lg font-bold">{{ t('admin.subscriptions.guide.title') }}</h2>
          <p class="subscription-guide-modal__description mb-5 text-sm">{{ t('admin.subscriptions.guide.subtitle') }}</p>

          <div class="subscription-guide-modal__section mb-5">
            <h3 class="subscription-guide-modal__section-title mb-2 flex items-center gap-2 text-sm font-semibold">
              <span class="subscription-guide-modal__step-index text-xs font-bold">1</span>
              {{ t('admin.subscriptions.guide.step1.title') }}
            </h3>
            <ol class="subscription-guide-modal__list ml-8 list-decimal space-y-1 text-sm">
              <li>{{ t('admin.subscriptions.guide.step1.line1') }}</li>
              <li>{{ t('admin.subscriptions.guide.step1.line2') }}</li>
              <li>{{ t('admin.subscriptions.guide.step1.line3') }}</li>
            </ol>
            <div class="ml-8 mt-2">
              <router-link
                to="/admin/groups"
                class="subscription-guide-modal__link inline-flex items-center gap-1 text-sm font-medium"
                @click="emit('close')"
              >
                {{ t('admin.subscriptions.guide.step1.link') }}
                <Icon name="arrowRight" size="xs" />
              </router-link>
            </div>
          </div>

          <div class="subscription-guide-modal__section mb-5">
            <h3 class="subscription-guide-modal__section-title mb-2 flex items-center gap-2 text-sm font-semibold">
              <span class="subscription-guide-modal__step-index text-xs font-bold">2</span>
              {{ t('admin.subscriptions.guide.step2.title') }}
            </h3>
            <ol class="subscription-guide-modal__list ml-8 list-decimal space-y-1 text-sm">
              <li>{{ t('admin.subscriptions.guide.step2.line1') }}</li>
              <li>{{ t('admin.subscriptions.guide.step2.line2') }}</li>
              <li>{{ t('admin.subscriptions.guide.step2.line3') }}</li>
            </ol>
          </div>

          <div class="subscription-guide-modal__section mb-5">
            <h3 class="subscription-guide-modal__section-title mb-2 flex items-center gap-2 text-sm font-semibold">
              <span class="subscription-guide-modal__step-index text-xs font-bold">3</span>
              {{ t('admin.subscriptions.guide.step3.title') }}
            </h3>
            <div class="subscription-guide-modal__table-wrap ml-8 overflow-hidden">
              <table class="w-full text-sm">
                <tbody>
                  <tr
                    v-for="(row, index) in guideActionRows"
                    :key="index"
                    class="subscription-guide-modal__table-row last:border-0"
                  >
                    <td class="subscription-guide-modal__table-key whitespace-nowrap font-medium">{{ row.action }}</td>
                    <td class="subscription-guide-modal__table-value">{{ row.desc }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <div class="subscription-guide-modal__tip text-xs">
            {{ t('admin.subscriptions.guide.tip') }}
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
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

const { t } = useI18n()

const guideActionRows = computed(() => [
  { action: t('admin.subscriptions.guide.actions.adjust'), desc: t('admin.subscriptions.guide.actions.adjustDesc') },
  { action: t('admin.subscriptions.guide.actions.resetQuota'), desc: t('admin.subscriptions.guide.actions.resetQuotaDesc') },
  { action: t('admin.subscriptions.guide.actions.revoke'), desc: t('admin.subscriptions.guide.actions.revokeDesc') }
])
</script>

<style scoped>
.subscription-guide-modal__overlay {
  background: var(--theme-overlay-strong);
}

.subscription-guide-modal {
  padding: var(--theme-markdown-block-padding);
}

.subscription-guide-modal__panel {
  max-width: min(100%, var(--theme-dialog-width-wide-sm));
  max-height: 85vh;
  padding: var(--theme-auth-callback-card-padding);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 78%, transparent);
  border-radius: var(--theme-subscription-panel-radius);
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow-hover);
}

.subscription-guide-modal__close {
  inset-block-start: calc(var(--theme-auth-callback-card-padding) * 0.67);
  inset-inline-end: calc(var(--theme-auth-callback-card-padding) * 0.67);
  color: color-mix(in srgb, var(--theme-page-muted) 72%, var(--theme-surface));
  transition: color 0.2s ease;
}

.subscription-guide-modal__close:hover {
  color: var(--theme-page-text);
}

.subscription-guide-modal__title,
.subscription-guide-modal__section-title,
.subscription-guide-modal__table-key {
  color: var(--theme-page-text);
}

.subscription-guide-modal__description,
.subscription-guide-modal__list,
.subscription-guide-modal__table-value {
  color: var(--theme-page-muted);
}

.subscription-guide-modal__step-index {
  background: color-mix(in srgb, var(--theme-accent-soft) 90%, var(--theme-surface));
  color: var(--theme-accent);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 2rem;
  width: 2rem;
  border-radius: 999px;
}

.subscription-guide-modal__table-key,
.subscription-guide-modal__table-value {
  padding:
    var(--theme-settings-card-panel-padding)
    var(--theme-settings-card-header-padding-x);
}

.subscription-guide-modal__link {
  color: var(--theme-accent);
}

.subscription-guide-modal__link:hover {
  color: color-mix(in srgb, var(--theme-accent) 84%, var(--theme-page-text));
}

.subscription-guide-modal__table-wrap {
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 74%, transparent);
  border-radius: var(--theme-subscription-panel-radius);
}

.subscription-guide-modal__table-row {
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.subscription-guide-modal__table-key {
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
}

.subscription-guide-modal__tip {
  padding: calc(var(--theme-markdown-block-padding) * 0.75);
  border-radius: var(--theme-button-radius);
  background: color-mix(in srgb, rgb(var(--theme-info-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
}
</style>
