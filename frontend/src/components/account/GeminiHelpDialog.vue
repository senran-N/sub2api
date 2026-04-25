<template>
  <BaseDialog
    :show="show"
    :title="t('admin.accounts.gemini.helpDialog.title')"
    width="wide"
    @close="emit('close')"
  >
    <div class="gemini-help-dialog space-y-6">
      <div>
        <h3 class="gemini-help-dialog__title">
          {{ t("admin.accounts.gemini.setupGuide.title") }}
        </h3>
        <div class="space-y-4">
          <div>
            <p class="gemini-help-dialog__subtitle">
              {{ t("admin.accounts.gemini.setupGuide.checklistTitle") }}
            </p>
            <ul class="gemini-help-dialog__list">
              <li>
                {{ t("admin.accounts.gemini.setupGuide.checklistItems.usIp") }}
              </li>
              <li>
                {{ t("admin.accounts.gemini.setupGuide.checklistItems.age") }}
              </li>
            </ul>
          </div>
          <div>
            <p class="gemini-help-dialog__subtitle">
              {{ t("admin.accounts.gemini.setupGuide.activationTitle") }}
            </p>
            <ul class="gemini-help-dialog__list">
              <li>
                {{
                  t(
                    "admin.accounts.gemini.setupGuide.activationItems.geminiWeb",
                  )
                }}
              </li>
              <li>
                {{
                  t(
                    "admin.accounts.gemini.setupGuide.activationItems.gcpProject",
                  )
                }}
              </li>
            </ul>
            <div class="mt-2 flex flex-wrap gap-2">
              <a
                :href="helpLinks.countryCheck"
                target="_blank"
                rel="noreferrer"
                class="gemini-help-dialog__link text-sm"
              >
                {{ t("admin.accounts.gemini.setupGuide.links.countryCheck") }}
              </a>
              <span class="gemini-help-dialog__muted">·</span>
              <a
                :href="helpLinks.countryChange"
                target="_blank"
                rel="noreferrer"
                class="gemini-help-dialog__link text-sm"
              >
                修改归属地
              </a>
              <span class="gemini-help-dialog__muted">·</span>
              <a
                :href="helpLinks.geminiWebActivation"
                target="_blank"
                rel="noreferrer"
                class="gemini-help-dialog__link text-sm"
              >
                {{
                  t(
                    "admin.accounts.gemini.setupGuide.links.geminiWebActivation",
                  )
                }}
              </a>
              <span class="gemini-help-dialog__muted">·</span>
              <a
                :href="helpLinks.gcpProject"
                target="_blank"
                rel="noreferrer"
                class="gemini-help-dialog__link text-sm"
              >
                {{ t("admin.accounts.gemini.setupGuide.links.gcpProject") }}
              </a>
            </div>
          </div>
        </div>
      </div>

      <div class="gemini-help-dialog__section">
        <h3 class="gemini-help-dialog__choice-title mb-3 text-sm font-semibold">
          {{ t("admin.accounts.gemini.quotaPolicy.title") }}
        </h3>
        <p class="gemini-help-dialog__notice gemini-help-dialog__notice--amber mb-4 text-xs">
          {{ t("admin.accounts.gemini.quotaPolicy.note") }}
        </p>
        <div class="overflow-x-auto">
          <table class="w-full text-xs">
            <thead class="gemini-help-dialog__table-head">
              <tr>
                <th class="gemini-help-dialog__table-heading">
                  {{ t("admin.accounts.gemini.quotaPolicy.columns.channel") }}
                </th>
                <th class="gemini-help-dialog__table-heading">
                  {{ t("admin.accounts.gemini.quotaPolicy.columns.account") }}
                </th>
                <th class="gemini-help-dialog__table-heading">
                  {{ t("admin.accounts.gemini.quotaPolicy.columns.limits") }}
                </th>
              </tr>
            </thead>
            <tbody class="gemini-help-dialog__table-body">
              <tr>
                <td class="gemini-help-dialog__table-primary">
                  {{
                    t(
                      "admin.accounts.gemini.quotaPolicy.rows.googleOne.channel",
                    )
                  }}
                </td>
                <td class="gemini-help-dialog__table-secondary">Free</td>
                <td class="gemini-help-dialog__table-secondary">
                  {{
                    t(
                      "admin.accounts.gemini.quotaPolicy.rows.googleOne.limitsFree",
                    )
                  }}
                </td>
              </tr>
              <tr>
                <td class="gemini-help-dialog__table-primary"></td>
                <td class="gemini-help-dialog__table-secondary">Pro</td>
                <td class="gemini-help-dialog__table-secondary">
                  {{
                    t(
                      "admin.accounts.gemini.quotaPolicy.rows.googleOne.limitsPro",
                    )
                  }}
                </td>
              </tr>
              <tr>
                <td class="gemini-help-dialog__table-primary"></td>
                <td class="gemini-help-dialog__table-secondary">Ultra</td>
                <td class="gemini-help-dialog__table-secondary">
                  {{
                    t(
                      "admin.accounts.gemini.quotaPolicy.rows.googleOne.limitsUltra",
                    )
                  }}
                </td>
              </tr>
              <tr>
                <td class="gemini-help-dialog__table-primary">
                  {{ t("admin.accounts.gemini.quotaPolicy.rows.gcp.channel") }}
                </td>
                <td class="gemini-help-dialog__table-secondary">Standard</td>
                <td class="gemini-help-dialog__table-secondary">
                  {{
                    t(
                      "admin.accounts.gemini.quotaPolicy.rows.gcp.limitsStandard",
                    )
                  }}
                </td>
              </tr>
              <tr>
                <td class="gemini-help-dialog__table-primary"></td>
                <td class="gemini-help-dialog__table-secondary">Enterprise</td>
                <td class="gemini-help-dialog__table-secondary">
                  {{
                    t(
                      "admin.accounts.gemini.quotaPolicy.rows.gcp.limitsEnterprise",
                    )
                  }}
                </td>
              </tr>
              <tr>
                <td class="gemini-help-dialog__table-primary">
                  {{
                    t("admin.accounts.gemini.quotaPolicy.rows.aiStudio.channel")
                  }}
                </td>
                <td class="gemini-help-dialog__table-secondary">Free</td>
                <td class="gemini-help-dialog__table-secondary">
                  {{
                    t(
                      "admin.accounts.gemini.quotaPolicy.rows.aiStudio.limitsFree",
                    )
                  }}
                </td>
              </tr>
              <tr>
                <td class="gemini-help-dialog__table-primary"></td>
                <td class="gemini-help-dialog__table-secondary">Paid</td>
                <td class="gemini-help-dialog__table-secondary">
                  {{
                    t(
                      "admin.accounts.gemini.quotaPolicy.rows.aiStudio.limitsPaid",
                    )
                  }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <div class="mt-4 flex flex-wrap gap-3">
          <a
            :href="quotaDocs.codeAssist"
            target="_blank"
            rel="noreferrer"
            class="gemini-help-dialog__link text-sm"
          >
            {{ t("admin.accounts.gemini.quotaPolicy.docs.codeAssist") }}
          </a>
          <a
            :href="quotaDocs.aiStudio"
            target="_blank"
            rel="noreferrer"
            class="gemini-help-dialog__link text-sm"
          >
            {{ t("admin.accounts.gemini.quotaPolicy.docs.aiStudio") }}
          </a>
          <a
            :href="quotaDocs.vertex"
            target="_blank"
            rel="noreferrer"
            class="gemini-help-dialog__link text-sm"
          >
            {{ t("admin.accounts.gemini.quotaPolicy.docs.vertex") }}
          </a>
        </div>
      </div>

      <div class="gemini-help-dialog__section">
        <h3 class="gemini-help-dialog__choice-title mb-3 text-sm font-semibold">
          {{ t("admin.accounts.gemini.helpDialog.apiKeySection") }}
        </h3>
        <div class="flex flex-wrap gap-3">
          <a
            :href="helpLinks.apiKey"
            target="_blank"
            rel="noreferrer"
            class="gemini-help-dialog__link text-sm"
          >
            {{ t("admin.accounts.gemini.accountType.apiKeyLink") }}
          </a>
          <a
            :href="helpLinks.aiStudioPricing"
            target="_blank"
            rel="noreferrer"
            class="gemini-help-dialog__link text-sm"
          >
            {{ t("admin.accounts.gemini.accountType.quotaLink") }}
          </a>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end">
        <button
          type="button"
          class="btn btn-primary"
          @click="emit('close')"
        >
          {{ t("common.close") }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import BaseDialog from "@/components/common/BaseDialog.vue";

defineProps<{
  show: boolean;
  helpLinks: {
    apiKey: string;
    aiStudioPricing: string;
    countryChange: string;
    countryCheck: string;
    gcpProject: string;
    geminiWebActivation: string;
  };
  quotaDocs: {
    aiStudio: string;
    codeAssist: string;
    vertex: string;
  };
}>();

const emit = defineEmits<{
  close: [];
}>();

const { t } = useI18n();
</script>

<style scoped>
.gemini-help-dialog__title,
.gemini-help-dialog__subtitle,
.gemini-help-dialog__choice-title {
  color: var(--theme-page-text);
}

.gemini-help-dialog__title {
  margin-bottom: 0.75rem;
  font-weight: 600;
}

.gemini-help-dialog__subtitle {
  margin-bottom: 0.5rem;
  font-size: 0.875rem;
  font-weight: 600;
}

.gemini-help-dialog__list {
  list-style: disc;
  margin-left: 1.25rem;
  color: var(--theme-page-muted);
  font-size: 0.875rem;
}

.gemini-help-dialog__muted {
  color: var(--theme-page-muted);
}

.gemini-help-dialog__link {
  color: var(--theme-accent);
}

.gemini-help-dialog__link:hover {
  color: var(--theme-accent-hover);
}

.gemini-help-dialog__section {
  border-top: 1px solid
    color-mix(in srgb, var(--theme-page-border) 76%, transparent);
  padding-top: 1.5rem;
}

.gemini-help-dialog__notice {
  display: block;
  border-radius: var(--theme-auth-feedback-radius);
  padding: 0.5rem 0.75rem;
}

.gemini-help-dialog__notice--amber {
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

.gemini-help-dialog__table-head {
  background: color-mix(
    in srgb,
    var(--theme-surface-soft) 90%,
    var(--theme-surface)
  );
}

.gemini-help-dialog__table-heading {
  color: var(--theme-page-muted);
  padding: 0.5rem;
  text-align: left;
}

.gemini-help-dialog__table-body {
  border-top: 1px solid
    color-mix(in srgb, var(--theme-page-border) 76%, transparent);
}

.gemini-help-dialog__table-primary,
.gemini-help-dialog__table-secondary {
  padding: 0.5rem;
}

.gemini-help-dialog__table-primary {
  color: var(--theme-page-text);
}

.gemini-help-dialog__table-secondary {
  color: var(--theme-page-muted);
}
</style>
