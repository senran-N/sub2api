<template>
  <div class="space-y-4">
    <div v-if="allowAuthModeChange">
      <label class="input-label">{{
        t("admin.accounts.bedrockAuthMode")
      }}</label>
      <div class="mt-2 flex gap-4">
        <label :class="getRadioOptionClasses(authMode === 'sigv4')">
          <input
            :checked="authMode === 'sigv4'"
            type="radio"
            value="sigv4"
            class="bedrock-credentials-section__radio-input"
            @change="emit('update:authMode', 'sigv4')"
          />
          <span class="bedrock-credentials-section__title text-sm">{{
            t("admin.accounts.bedrockAuthModeSigv4")
          }}</span>
        </label>
        <label :class="getRadioOptionClasses(authMode === 'apikey')">
          <input
            :checked="authMode === 'apikey'"
            type="radio"
            value="apikey"
            class="bedrock-credentials-section__radio-input"
            @change="emit('update:authMode', 'apikey')"
          />
          <span class="bedrock-credentials-section__title text-sm">{{
            t("admin.accounts.bedrockAuthModeApikey")
          }}</span>
        </label>
      </div>
    </div>

    <template v-if="authMode === 'sigv4'">
      <div>
        <label class="input-label">{{
          t("admin.accounts.bedrockAccessKeyId")
        }}</label>
        <input
          :value="accessKeyId"
          type="text"
          :required="credentialsRequired"
          class="input font-mono"
          placeholder="AKIA..."
          @input="emitInputValue($event, 'update:accessKeyId')"
        />
      </div>
      <div>
        <label class="input-label">{{
          t("admin.accounts.bedrockSecretAccessKey")
        }}</label>
        <input
          :value="secretAccessKey"
          type="password"
          :required="credentialsRequired"
          class="input font-mono"
          :placeholder="secretAccessKeyPlaceholder"
          @input="emitInputValue($event, 'update:secretAccessKey')"
        />
        <p v-if="secretAccessKeyHint" class="input-hint">
          {{ secretAccessKeyHint }}
        </p>
      </div>
      <div>
        <label class="input-label">{{
          t("admin.accounts.bedrockSessionToken")
        }}</label>
        <input
          :value="sessionToken"
          type="password"
          class="input font-mono"
          :placeholder="sessionTokenPlaceholder"
          @input="emitInputValue($event, 'update:sessionToken')"
        />
        <p class="input-hint">
          {{ t("admin.accounts.bedrockSessionTokenHint") }}
        </p>
      </div>
    </template>

    <div v-if="authMode === 'apikey'">
      <label class="input-label">{{
        t("admin.accounts.bedrockApiKeyInput")
      }}</label>
      <input
        :value="apiKeyValue"
        type="password"
        :required="credentialsRequired"
        class="input font-mono"
        :placeholder="apiKeyPlaceholder"
        @input="emitInputValue($event, 'update:apiKeyValue')"
      />
      <p v-if="apiKeyHint" class="input-hint">
        {{ apiKeyHint }}
      </p>
    </div>

    <div>
      <label class="input-label">{{
        t("admin.accounts.bedrockRegion")
      }}</label>
      <input
        v-if="regionControl === 'input'"
        :value="region"
        type="text"
        class="input"
        placeholder="us-east-1"
        @input="emitInputValue($event, 'update:region')"
      />
      <select
        v-else
        :value="region"
        class="input"
        @change="emitInputValue($event, 'update:region')"
      >
        <optgroup label="US">
          <option value="us-east-1">us-east-1 (N. Virginia)</option>
          <option value="us-east-2">us-east-2 (Ohio)</option>
          <option value="us-west-1">us-west-1 (N. California)</option>
          <option value="us-west-2">us-west-2 (Oregon)</option>
          <option value="us-gov-east-1">us-gov-east-1 (GovCloud US-East)</option>
          <option value="us-gov-west-1">us-gov-west-1 (GovCloud US-West)</option>
        </optgroup>
        <optgroup label="Europe">
          <option value="eu-west-1">eu-west-1 (Ireland)</option>
          <option value="eu-west-2">eu-west-2 (London)</option>
          <option value="eu-west-3">eu-west-3 (Paris)</option>
          <option value="eu-central-1">eu-central-1 (Frankfurt)</option>
          <option value="eu-central-2">eu-central-2 (Zurich)</option>
          <option value="eu-south-1">eu-south-1 (Milan)</option>
          <option value="eu-south-2">eu-south-2 (Spain)</option>
          <option value="eu-north-1">eu-north-1 (Stockholm)</option>
        </optgroup>
        <optgroup label="Asia Pacific">
          <option value="ap-northeast-1">ap-northeast-1 (Tokyo)</option>
          <option value="ap-northeast-2">ap-northeast-2 (Seoul)</option>
          <option value="ap-northeast-3">ap-northeast-3 (Osaka)</option>
          <option value="ap-south-1">ap-south-1 (Mumbai)</option>
          <option value="ap-south-2">ap-south-2 (Hyderabad)</option>
          <option value="ap-southeast-1">ap-southeast-1 (Singapore)</option>
          <option value="ap-southeast-2">ap-southeast-2 (Sydney)</option>
        </optgroup>
        <optgroup label="Canada">
          <option value="ca-central-1">ca-central-1 (Canada)</option>
        </optgroup>
        <optgroup label="South America">
          <option value="sa-east-1">sa-east-1 (S&atilde;o Paulo)</option>
        </optgroup>
      </select>
      <p class="input-hint">{{ t("admin.accounts.bedrockRegionHint") }}</p>
    </div>

    <div>
      <label class="bedrock-credentials-section__checkbox">
        <input
          :checked="forceGlobal"
          type="checkbox"
          class="bedrock-credentials-section__checkbox-input"
          @change="
            emit(
              'update:forceGlobal',
              ($event.target as HTMLInputElement).checked,
            )
          "
        />
        <span class="bedrock-credentials-section__title text-sm">{{
          t("admin.accounts.bedrockForceGlobal")
        }}</span>
      </label>
      <p class="input-hint mt-1">
        {{ t("admin.accounts.bedrockForceGlobalHint") }}
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import type { BedrockAuthMode } from "@/components/account/createAccountModalHelpers";

type StringUpdateEvent =
  | "update:accessKeyId"
  | "update:secretAccessKey"
  | "update:sessionToken"
  | "update:apiKeyValue"
  | "update:region";

const props = withDefaults(defineProps<{
  authMode: BedrockAuthMode;
  accessKeyId: string;
  secretAccessKey: string;
  sessionToken: string;
  apiKeyValue: string;
  region: string;
  forceGlobal: boolean;
  allowAuthModeChange?: boolean;
  apiKeyHintKey?: string;
  apiKeyPlaceholderKey?: string;
  credentialsRequired?: boolean;
  regionControl?: "input" | "select";
  secretAccessKeyHintKey?: string;
  secretAccessKeyPlaceholderKey?: string;
  sessionTokenPlaceholderKey?: string;
}>(), {
  allowAuthModeChange: true,
  apiKeyHintKey: "",
  apiKeyPlaceholderKey: "",
  credentialsRequired: true,
  regionControl: "select",
  secretAccessKeyHintKey: "",
  secretAccessKeyPlaceholderKey: "",
  sessionTokenPlaceholderKey: "",
});

const emit = defineEmits<{
  "update:authMode": [value: BedrockAuthMode];
  "update:accessKeyId": [value: string];
  "update:secretAccessKey": [value: string];
  "update:sessionToken": [value: string];
  "update:apiKeyValue": [value: string];
  "update:region": [value: string];
  "update:forceGlobal": [value: boolean];
}>();

const { t } = useI18n();

const secretAccessKeyPlaceholder = computed(() =>
  props.secretAccessKeyPlaceholderKey ? t(props.secretAccessKeyPlaceholderKey) : "",
);
const secretAccessKeyHint = computed(() =>
  props.secretAccessKeyHintKey ? t(props.secretAccessKeyHintKey) : "",
);
const sessionTokenPlaceholder = computed(() =>
  props.sessionTokenPlaceholderKey ? t(props.sessionTokenPlaceholderKey) : "",
);
const apiKeyPlaceholder = computed(() =>
  props.apiKeyPlaceholderKey ? t(props.apiKeyPlaceholderKey) : "",
);
const apiKeyHint = computed(() =>
  props.apiKeyHintKey ? t(props.apiKeyHintKey) : "",
);

const getRadioOptionClasses = (isSelected: boolean) => [
  "bedrock-credentials-section__radio-option",
  isSelected && "bedrock-credentials-section__radio-option--active",
];

const emitInputValue = (event: Event, emitName: StringUpdateEvent) => {
  const value = (event.target as HTMLInputElement | HTMLSelectElement).value;
  switch (emitName) {
    case "update:accessKeyId":
      emit("update:accessKeyId", value);
      return;
    case "update:secretAccessKey":
      emit("update:secretAccessKey", value);
      return;
    case "update:sessionToken":
      emit("update:sessionToken", value);
      return;
    case "update:apiKeyValue":
      emit("update:apiKeyValue", value);
      return;
    case "update:region":
      emit("update:region", value);
      return;
  }
};
</script>

<style scoped>
.bedrock-credentials-section__radio-option,
.bedrock-credentials-section__checkbox {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
}

.bedrock-credentials-section__radio-option {
  color: var(--theme-page-muted);
}

.bedrock-credentials-section__radio-option--active,
.bedrock-credentials-section__title {
  color: var(--theme-page-text);
}

.bedrock-credentials-section__radio-input,
.bedrock-credentials-section__checkbox-input {
  color: var(--theme-accent);
}
</style>
