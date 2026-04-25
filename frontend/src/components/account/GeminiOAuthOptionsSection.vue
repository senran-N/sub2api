<template>
  <div class="mt-4">
    <label class="input-label">{{
      t("admin.accounts.oauth.gemini.oauthTypeLabel")
    }}</label>
    <div class="mt-2 grid grid-cols-2 gap-3">
      <CreateAccountChoiceCard
        :selected="oauthType === 'google_one'"
        tone="purple"
        icon="user"
        :title="t('admin.accounts.gemini.oauthType.googleOneTitle')"
        :description="t('admin.accounts.gemini.oauthType.googleOneDesc')"
        @select="emit('update:oauthType', 'google_one')"
      >
        <template #meta>
          <span class="gemini-oauth-options-section__tone-tag gemini-oauth-options-section__tone-tag--purple">
            {{ t("admin.accounts.gemini.oauthType.badges.personal") }}
          </span>
          <span class="gemini-oauth-options-section__tone-tag gemini-oauth-options-section__tone-tag--emerald">
            {{ t("admin.accounts.gemini.oauthType.badges.noGcp") }}
          </span>
        </template>
      </CreateAccountChoiceCard>

      <CreateAccountChoiceCard
        :selected="oauthType === 'code_assist'"
        tone="blue"
        icon="cloud"
        :title="t('admin.accounts.gemini.oauthType.codeAssistTitle')"
        :description="t('admin.accounts.gemini.oauthType.codeAssistDesc')"
        @select="emit('update:oauthType', 'code_assist')"
      >
        <template #meta>
          <span class="gemini-oauth-options-section__tone-tag gemini-oauth-options-section__tone-tag--blue">
            {{ t("admin.accounts.gemini.oauthType.badges.enterprise") }}
          </span>
          <span class="gemini-oauth-options-section__tone-tag gemini-oauth-options-section__tone-tag--emerald">
            {{ t("admin.accounts.gemini.oauthType.badges.highConcurrency") }}
          </span>
        </template>
      </CreateAccountChoiceCard>
    </div>

    <p
      v-if="oauthType === 'code_assist'"
      class="gemini-oauth-options-section__description mt-2 text-xs"
    >
      {{ t("admin.accounts.gemini.oauthType.codeAssistRequirement") }}
      <a
        :href="gcpProjectHelpLink"
        class="gemini-oauth-options-section__link ml-1"
        target="_blank"
        rel="noreferrer"
      >
        {{ t("admin.accounts.gemini.oauthType.gcpProjectLink") }}
      </a>
    </p>

    <div class="mt-3">
      <button
        type="button"
        class="gemini-oauth-options-section__inline-toggle flex items-center gap-2 text-sm"
        @click="showAdvancedOAuth = !showAdvancedOAuth"
      >
        <Icon
          name="chevronRight"
          size="sm"
          :stroke-width="2"
          :class="['transition-transform', showAdvancedOAuth ? 'rotate-90' : '']"
        />
        <span>
          {{
            t(
              showAdvancedOAuth
                ? "admin.accounts.gemini.oauthType.hideAdvanced"
                : "admin.accounts.gemini.oauthType.showAdvanced",
            )
          }}
        </span>
      </button>
    </div>

    <div v-if="showAdvancedOAuth" class="group relative mt-3">
      <CreateAccountChoiceCard
        :selected="oauthType === 'ai_studio'"
        tone="amber"
        icon="sparkles"
        :disabled="!aiStudioOauthEnabled"
        :title="t('admin.accounts.gemini.oauthType.customTitle')"
        :description="t('admin.accounts.gemini.oauthType.customDesc')"
        @select="emit('update:oauthType', 'ai_studio')"
      >
        <template #meta>
          <span class="gemini-oauth-options-section__description basis-full text-xs">
            {{ t("admin.accounts.gemini.oauthType.customRequirement") }}
          </span>
          <span class="gemini-oauth-options-section__tone-tag gemini-oauth-options-section__tone-tag--amber">
            {{ t("admin.accounts.gemini.oauthType.badges.orgManaged") }}
          </span>
          <span class="gemini-oauth-options-section__tone-tag gemini-oauth-options-section__tone-tag--amber">
            {{ t("admin.accounts.gemini.oauthType.badges.adminRequired") }}
          </span>
          <span
            v-if="!aiStudioOauthEnabled"
            class="gemini-oauth-options-section__tone-tag gemini-oauth-options-section__tone-tag--amber"
          >
            {{
              t("admin.accounts.oauth.gemini.aiStudioNotConfiguredShort")
            }}
          </span>
        </template>
      </CreateAccountChoiceCard>

      <div
        v-if="!aiStudioOauthEnabled"
        class="gemini-oauth-options-section__notice-tooltip pointer-events-none absolute right-0 top-full z-50 mt-2 text-xs opacity-0 shadow-lg transition-opacity group-hover:opacity-100"
      >
        {{ t("admin.accounts.oauth.gemini.aiStudioNotConfiguredTip") }}
      </div>
    </div>

    <div class="mt-4">
      <label class="input-label">{{
        t("admin.accounts.gemini.tier.label")
      }}</label>
      <div class="mt-2">
        <select
          v-if="oauthType === 'google_one'"
          :value="tierGoogleOne"
          class="input"
          @change="updateTierGoogleOne"
        >
          <option value="google_one_free">
            {{ t("admin.accounts.gemini.tier.googleOne.free") }}
          </option>
          <option value="google_ai_pro">
            {{ t("admin.accounts.gemini.tier.googleOne.pro") }}
          </option>
          <option value="google_ai_ultra">
            {{ t("admin.accounts.gemini.tier.googleOne.ultra") }}
          </option>
        </select>

        <select
          v-else-if="oauthType === 'code_assist'"
          :value="tierGcp"
          class="input"
          @change="updateTierGcp"
        >
          <option value="gcp_standard">
            {{ t("admin.accounts.gemini.tier.gcp.standard") }}
          </option>
          <option value="gcp_enterprise">
            {{ t("admin.accounts.gemini.tier.gcp.enterprise") }}
          </option>
        </select>

        <select
          v-else
          :value="tierAiStudio"
          class="input"
          @change="updateTierAiStudio"
        >
          <option value="aistudio_free">
            {{ t("admin.accounts.gemini.tier.aiStudio.free") }}
          </option>
          <option value="aistudio_paid">
            {{ t("admin.accounts.gemini.tier.aiStudio.paid") }}
          </option>
        </select>
      </div>
      <p class="input-hint">{{ t("admin.accounts.gemini.tier.hint") }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import CreateAccountChoiceCard from "@/components/account/CreateAccountChoiceCard.vue";
import Icon from "@/components/icons/Icon.vue";
import type {
  GeminiAIStudioTier,
  GeminiGcpTier,
  GeminiGoogleOneTier,
  GeminiOAuthType,
} from "@/components/account/createAccountModalHelpers";

defineProps<{
  oauthType: GeminiOAuthType;
  aiStudioOauthEnabled: boolean;
  gcpProjectHelpLink: string;
  tierGoogleOne: GeminiGoogleOneTier;
  tierGcp: GeminiGcpTier;
  tierAiStudio: GeminiAIStudioTier;
}>();

const emit = defineEmits<{
  "update:oauthType": [value: GeminiOAuthType];
  "update:tierGoogleOne": [value: GeminiGoogleOneTier];
  "update:tierGcp": [value: GeminiGcpTier];
  "update:tierAiStudio": [value: GeminiAIStudioTier];
}>();

const { t } = useI18n();
const showAdvancedOAuth = ref(false);

const updateTierGoogleOne = (event: Event) => {
  emit(
    "update:tierGoogleOne",
    (event.target as HTMLSelectElement).value as GeminiGoogleOneTier,
  );
};

const updateTierGcp = (event: Event) => {
  emit(
    "update:tierGcp",
    (event.target as HTMLSelectElement).value as GeminiGcpTier,
  );
};

const updateTierAiStudio = (event: Event) => {
  emit(
    "update:tierAiStudio",
    (event.target as HTMLSelectElement).value as GeminiAIStudioTier,
  );
};
</script>

<style scoped>
.gemini-oauth-options-section__description {
  color: var(--theme-page-muted);
}

.gemini-oauth-options-section__inline-toggle {
  color: var(--theme-page-muted);
}

.gemini-oauth-options-section__inline-toggle:hover {
  color: var(--theme-page-text);
}

.gemini-oauth-options-section__link {
  color: color-mix(
    in srgb,
    rgb(var(--theme-info-rgb)) 84%,
    var(--theme-page-text)
  );
}

.gemini-oauth-options-section__link:hover {
  text-decoration: underline;
}

.gemini-oauth-options-section__tone-tag {
  display: inline-flex;
  align-items: center;
  border-radius: var(--theme-button-radius);
  padding: 0.125rem 0.5rem;
  background: color-mix(
    in srgb,
    rgb(var(--gemini-oauth-options-tone-rgb)) 16%,
    var(--theme-surface)
  );
  color: color-mix(
    in srgb,
    rgb(var(--gemini-oauth-options-tone-rgb)) 88%,
    var(--theme-page-text)
  );
}

.gemini-oauth-options-section__tone-tag--purple {
  --gemini-oauth-options-tone-rgb: var(--theme-brand-purple-rgb);
}

.gemini-oauth-options-section__tone-tag--emerald {
  --gemini-oauth-options-tone-rgb: var(--theme-success-rgb);
}

.gemini-oauth-options-section__tone-tag--blue {
  --gemini-oauth-options-tone-rgb: var(--theme-info-rgb);
}

.gemini-oauth-options-section__tone-tag--amber {
  --gemini-oauth-options-tone-rgb: var(--theme-warning-rgb);
}

.gemini-oauth-options-section__notice-tooltip {
  width: 20rem;
  padding: 0.5rem 0.75rem;
  border-radius: var(--theme-button-radius);
  border-color: color-mix(in srgb, var(--theme-card-border) 68%, transparent);
  background: color-mix(
    in srgb,
    rgb(var(--theme-warning-rgb)) 10%,
    var(--theme-surface)
  );
  color: color-mix(
    in srgb,
    rgb(var(--theme-warning-rgb)) 84%,
    var(--theme-page-text)
  );
}
</style>
