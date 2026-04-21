<template>
  <div class="space-y-4">
    <div>
      <label class="input-label">{{
        t("admin.accounts.grok.importMode")
      }}</label>
      <div class="grok-session-panel__mode-toggle mt-2 flex gap-2">
        <button
          type="button"
          class="grok-session-panel__mode-button flex-1 text-sm font-medium"
          :class="
            mode === 'single'
              ? 'grok-session-panel__mode-button--active'
              : 'grok-session-panel__mode-button--idle'
          "
          :disabled="submitting"
          data-testid="grok-session-mode-single"
          @click="$emit('update:mode', 'single')"
        >
          {{ t("admin.accounts.grok.singleInputMode") }}
        </button>
        <button
          type="button"
          class="grok-session-panel__mode-button flex-1 text-sm font-medium"
          :class="
            mode === 'batch'
              ? 'grok-session-panel__mode-button--active'
              : 'grok-session-panel__mode-button--idle'
          "
          :disabled="submitting"
          data-testid="grok-session-mode-batch"
          @click="$emit('update:mode', 'batch')"
        >
          {{ t("admin.accounts.grok.batchImportMode") }}
        </button>
      </div>
    </div>

    <div v-if="mode === 'single'">
      <label class="input-label">{{
        t("admin.accounts.grok.sessionToken")
      }}</label>
      <input
        :value="singleToken"
        type="password"
        class="input font-mono"
        :placeholder="t('admin.accounts.grok.sessionTokenPlaceholder')"
        data-testid="grok-session-single-input"
        :disabled="submitting"
        @input="
          $emit('update:singleToken', ($event.target as HTMLInputElement).value)
        "
      />
      <p class="input-hint">{{ t("admin.accounts.grok.sessionTokenHint") }}</p>
    </div>

    <div v-else class="space-y-4">
      <div class="grok-session-panel__card space-y-3">
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div>
            <label class="input-label">{{
              t("admin.accounts.grok.batchImportTitle")
            }}</label>
            <p class="grok-session-panel__muted mt-1 text-xs">
              {{ t("admin.accounts.grok.batchImportHint") }}
            </p>
          </div>
          <span
            class="grok-session-panel__count rounded-full px-3 py-1 text-xs font-semibold"
            data-testid="grok-session-batch-count"
          >
            {{
              t("admin.accounts.oauth.keysCount", { count: parsedLineCount })
            }}
          </span>
        </div>

        <textarea
          :value="batchInput"
          rows="6"
          class="input w-full resize-y font-mono text-sm"
          :placeholder="t('admin.accounts.grok.batchImportPlaceholder')"
          data-testid="grok-session-batch-input"
          :disabled="submitting"
          @input="
            $emit(
              'update:batchInput',
              ($event.target as HTMLTextAreaElement).value,
            )
          "
        ></textarea>

        <p class="input-hint">
          {{ t("admin.accounts.grok.batchImportFormats") }}
        </p>
      </div>

      <div class="grid gap-3 sm:grid-cols-2">
        <label class="grok-session-panel__option">
          <input
            type="checkbox"
            class="grok-session-panel__checkbox"
            :checked="dryRun"
            :disabled="submitting"
            data-testid="grok-session-batch-dry-run"
            @change="
              $emit(
                'update:dryRun',
                ($event.target as HTMLInputElement).checked,
              )
            "
          />
          <span>
            <span
              class="grok-session-panel__option-title block text-sm font-medium"
            >
              {{ t("admin.accounts.grok.batchImportDryRun") }}
            </span>
            <span class="grok-session-panel__muted text-xs">
              {{ t("admin.accounts.grok.batchImportDryRunHint") }}
            </span>
          </span>
        </label>

        <label class="grok-session-panel__option">
          <input
            type="checkbox"
            class="grok-session-panel__checkbox"
            :checked="testAfterCreate"
            :disabled="submitting"
            data-testid="grok-session-batch-test-after-create"
            @change="
              $emit(
                'update:testAfterCreate',
                ($event.target as HTMLInputElement).checked,
              )
            "
          />
          <span>
            <span
              class="grok-session-panel__option-title block text-sm font-medium"
            >
              {{ t("admin.accounts.grok.batchImportTestAfterCreate") }}
            </span>
            <span class="grok-session-panel__muted text-xs">
              {{ t("admin.accounts.grok.batchImportTestAfterCreateHint") }}
            </span>
          </span>
        </label>
      </div>

      <div class="grok-session-panel__notice text-xs">
        <p>{{ t("admin.accounts.grok.batchImportSettingsHint") }}</p>
        <p>{{ t("admin.accounts.grok.batchImportDedupeHint") }}</p>
      </div>

      <div
        v-if="result"
        class="grok-session-panel__card space-y-3"
        data-testid="grok-session-batch-result"
      >
        <div class="flex flex-wrap items-center gap-2">
          <!-- Semantic tones: total is neutral, created/skipped/invalid map to
               success/warning/danger, and the dry-run flag stays info-blue so
               the palette reflects each metric's meaning instead of a single
               uniform blue that clashes with the warm surface. -->
          <span class="theme-chip theme-chip--regular theme-chip--neutral">
            {{
              t("admin.accounts.grok.batchImportSummaryTotal", {
                count: result.total,
              })
            }}
          </span>
          <span class="theme-chip theme-chip--regular theme-chip--success">
            {{
              t("admin.accounts.grok.batchImportSummaryCreated", {
                count: result.created,
              })
            }}
          </span>
          <span class="theme-chip theme-chip--regular theme-chip--warning">
            {{
              t("admin.accounts.grok.batchImportSummarySkipped", {
                count: result.skipped,
              })
            }}
          </span>
          <span class="theme-chip theme-chip--regular theme-chip--danger">
            {{
              t("admin.accounts.grok.batchImportSummaryInvalid", {
                count: result.invalid,
              })
            }}
          </span>
          <span
            v-if="result.dry_run"
            class="theme-chip theme-chip--regular theme-chip--info"
          >
            {{ t("admin.accounts.grok.batchImportDryRunBadge") }}
          </span>
        </div>

        <div class="space-y-2">
          <div
            v-for="item in result.results"
            :key="`${item.line}-${item.fingerprint || item.name || item.reason || item.success}`"
            class="grok-session-panel__result-row"
          >
            <div class="flex flex-wrap items-center gap-2">
              <span class="grok-session-panel__result-line"
                >#{{ item.line }}</span
              >
              <span
                class="grok-session-panel__result-state"
                :class="
                  item.success
                    ? 'grok-session-panel__result-state--success'
                    : 'grok-session-panel__result-state--muted'
                "
              >
                {{
                  item.success
                    ? t("admin.accounts.grok.batchImportResultSuccess")
                    : t("admin.accounts.grok.batchImportResultSkipped")
                }}
              </span>
              <span v-if="item.name" class="text-sm font-medium">{{
                item.name
              }}</span>
            </div>
            <p
              v-if="item.fingerprint"
              class="grok-session-panel__fingerprint mt-1 text-xs"
            >
              {{ item.fingerprint }}
            </p>
            <p
              v-if="item.reason"
              class="grok-session-panel__muted mt-1 text-xs"
            >
              {{ item.reason }}
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import type { GrokSessionBatchImportResult } from "@/api/admin/accounts";
import { countMultilineEntries } from "./oauthAuthorizationFlowHelpers";

interface Props {
  mode: "single" | "batch";
  singleToken: string;
  batchInput: string;
  dryRun: boolean;
  testAfterCreate: boolean;
  result: GrokSessionBatchImportResult | null;
  submitting?: boolean;
}

defineEmits<{
  "update:mode": [value: "single" | "batch"];
  "update:singleToken": [value: string];
  "update:batchInput": [value: string];
  "update:dryRun": [value: boolean];
  "update:testAfterCreate": [value: boolean];
}>();

const { t } = useI18n();

const props = defineProps<Props>();

const parsedLineCount = computed(() => countMultilineEntries(props.batchInput));
</script>

<style scoped>
.grok-session-panel__mode-toggle {
  border-radius: calc(var(--theme-button-radius) + 4px);
  padding: 0.25rem;
  background: color-mix(
    in srgb,
    var(--theme-surface-soft) 88%,
    var(--theme-surface)
  );
  border: 1px solid
    color-mix(in srgb, var(--theme-card-border) 72%, transparent);
}

.grok-session-panel__mode-button {
  border-radius: calc(var(--theme-button-radius) + 1px);
  padding: 0.625rem 0.875rem;
  transition:
    background-color 0.18s ease,
    color 0.18s ease,
    box-shadow 0.18s ease;
}

.grok-session-panel__mode-button--active {
  background: var(--theme-surface);
  color: var(--theme-page-text);
  box-shadow: var(--theme-card-shadow);
}

.grok-session-panel__mode-button--idle {
  color: var(--theme-page-muted);
}

.grok-session-panel__mode-button--idle:hover {
  color: var(--theme-page-text);
}

.grok-session-panel__card {
  border-radius: var(--theme-surface-radius);
  padding: var(--theme-markdown-block-padding);
  border: 1px solid
    color-mix(in srgb, var(--theme-card-border) 68%, transparent);
  background: color-mix(
    in srgb,
    var(--theme-surface-soft) 90%,
    var(--theme-surface)
  );
}

.grok-session-panel__muted {
  color: var(--theme-page-muted);
}

.grok-session-panel__count {
  background: color-mix(
    in srgb,
    rgb(var(--theme-info-rgb)) 12%,
    var(--theme-surface)
  );
  color: color-mix(
    in srgb,
    rgb(var(--theme-info-rgb)) 86%,
    var(--theme-page-text)
  );
  border: 1px solid
    color-mix(in srgb, rgb(var(--theme-info-rgb)) 24%, transparent);
}

.grok-session-panel__option {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  border-radius: calc(var(--theme-button-radius) + 2px);
  border: 1px solid
    color-mix(in srgb, var(--theme-card-border) 72%, transparent);
  background: color-mix(
    in srgb,
    var(--theme-surface-soft) 82%,
    var(--theme-surface)
  );
  padding: 0.75rem;
}

.grok-session-panel__checkbox {
  margin-top: 0.125rem;
  accent-color: var(--theme-accent);
}

.grok-session-panel__option-title {
  color: var(--theme-page-text);
}

.grok-session-panel__notice {
  border-radius: var(--theme-auth-feedback-radius);
  padding: 0.75rem 0.875rem;
  border: 1px solid
    color-mix(in srgb, rgb(var(--theme-success-rgb)) 20%, transparent);
  background: color-mix(
    in srgb,
    rgb(var(--theme-success-rgb)) 10%,
    var(--theme-surface)
  );
  color: color-mix(
    in srgb,
    rgb(var(--theme-success-rgb)) 84%,
    var(--theme-page-text)
  );
}

.grok-session-panel__result-row {
  border-radius: calc(var(--theme-button-radius) + 2px);
  padding: 0.75rem;
  border: 1px solid
    color-mix(in srgb, var(--theme-card-border) 72%, transparent);
  background: color-mix(
    in srgb,
    var(--theme-surface-soft) 84%,
    var(--theme-surface)
  );
}

.grok-session-panel__result-line {
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--theme-page-muted);
}

.grok-session-panel__result-state {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: 0.125rem 0.5rem;
  font-size: 0.75rem;
  font-weight: 700;
}

.grok-session-panel__result-state--success {
  background: color-mix(
    in srgb,
    rgb(var(--theme-success-rgb)) 12%,
    var(--theme-surface)
  );
  color: color-mix(
    in srgb,
    rgb(var(--theme-success-rgb)) 90%,
    var(--theme-page-text)
  );
}

.grok-session-panel__result-state--muted {
  background: color-mix(
    in srgb,
    var(--theme-page-border) 84%,
    var(--theme-surface)
  );
  color: var(--theme-page-muted);
}

.grok-session-panel__fingerprint {
  word-break: break-all;
  /* Fingerprints are technical identifiers — render them in a muted monospace
     tone so they don't compete with semantic status chips and primary text. */
  font-family: var(--theme-font-mono);
  color: var(--theme-page-muted);
}
</style>
