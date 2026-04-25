<template>
  <div>
    <label class="input-label">{{ t("admin.accounts.platform") }}</label>
    <div class="segmented-control mt-2 flex" data-tour="account-form-platform">
      <button
        type="button"
        @click="selectPlatform('anthropic')"
        :class="buttonClasses('anthropic')"
      >
        <Icon name="sparkles" size="sm" />
        Anthropic
      </button>
      <button
        type="button"
        @click="selectPlatform('openai')"
        :class="buttonClasses('openai')"
      >
        <svg
          class="h-4 w-4"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="1.5"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M3.75 13.5l10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75z"
          />
        </svg>
        OpenAI
      </button>
      <button
        type="button"
        @click="selectPlatform('gemini')"
        :class="buttonClasses('gemini')"
      >
        <svg
          class="h-4 w-4"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="1.5"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M12 2l1.5 6.5L20 10l-6.5 1.5L12 18l-1.5-6.5L4 10l6.5-1.5L12 2z"
          />
        </svg>
        Gemini
      </button>
      <button
        type="button"
        @click="selectPlatform('grok')"
        :class="buttonClasses('grok')"
      >
        <svg
          class="h-4 w-4"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="1.75"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M13 2L4 14h6l-1 8 9-12h-6l1-8z"
          />
        </svg>
        Grok
      </button>
      <button
        type="button"
        @click="selectPlatform('antigravity')"
        :class="buttonClasses('antigravity')"
      >
        <Icon name="cloud" size="sm" />
        Antigravity
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import type { AccountPlatform } from "@/types";
import Icon from "@/components/icons/Icon.vue";

const props = defineProps<{
  modelValue: AccountPlatform;
}>();

const emit = defineEmits<{
  "update:modelValue": [platform: AccountPlatform];
}>();

const { t } = useI18n();

function selectPlatform(platform: AccountPlatform) {
  emit("update:modelValue", platform);
}

function buttonClasses(platform: AccountPlatform) {
  return [
    "create-account-modal__platform-button create-account-modal__platform-button-control flex flex-1 items-center justify-center gap-2 text-sm font-medium transition-all",
    props.modelValue === platform
      ? `create-account-modal__platform-button--active create-account-modal__platform-button--${platform}`
      : "create-account-modal__platform-button--idle",
  ];
}
</script>

<style scoped>
.create-account-modal__platform-button {
  color: var(--theme-page-muted);
}

.create-account-modal__platform-button-control {
  border-radius: calc(var(--theme-button-radius) - 2px);
  padding: 0.625rem 1rem;
}

.create-account-modal__platform-button--idle:hover {
  color: var(--theme-page-text);
}

.create-account-modal__platform-button--active {
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow);
}

.create-account-modal__platform-button--anthropic {
  color: color-mix(
    in srgb,
    rgb(var(--theme-brand-orange-rgb)) 84%,
    var(--theme-page-text)
  );
}

.create-account-modal__platform-button--openai {
  color: color-mix(
    in srgb,
    rgb(var(--theme-success-rgb)) 84%,
    var(--theme-page-text)
  );
}

.create-account-modal__platform-button--gemini {
  color: color-mix(
    in srgb,
    rgb(var(--theme-info-rgb)) 84%,
    var(--theme-page-text)
  );
}

.create-account-modal__platform-button--antigravity {
  color: color-mix(
    in srgb,
    rgb(var(--theme-brand-purple-rgb)) 84%,
    var(--theme-page-text)
  );
}
</style>
