import { computed, reactive, ref } from "vue";
import type { Account } from "@/types";
import {
  createDefaultEditAccountForm,
  hydrateEditAccountForm,
  type EditAccountForm,
} from "@/components/account/accountModalShared";
import {
  formatDateTimeLocalInput,
  parseDateTimeLocalInput,
} from "@/utils/format";

type Translate = (key: string) => string;

export function useEditAccountFormState(t: Translate) {
  const form = reactive<EditAccountForm>(createDefaultEditAccountForm());
  const autoPauseOnExpired = ref(false);
  const mixedScheduling = ref(false);
  const allowOverages = ref(false);

  const statusOptions = computed<
    Array<{ value: EditAccountForm["status"]; label: string }>
  >(() => {
    const options: Array<{
      value: EditAccountForm["status"];
      label: string;
    }> = [
      { value: "active", label: t("common.active") },
      { value: "inactive", label: t("common.inactive") },
    ];
    if (form.status === "error") {
      options.push({ value: "error", label: t("admin.accounts.status.error") });
    }
    return options;
  });

  const expiresAtInput = computed({
    get: () => formatDateTimeLocalInput(form.expires_at),
    set: (value: string) => {
      form.expires_at = parseDateTimeLocalInput(value);
    },
  });

  const hydrateFormStateFromAccount = (account: Account) => {
    hydrateEditAccountForm(form, account);
    autoPauseOnExpired.value = account.auto_pause_on_expired === true;
    const extra = account.extra as Record<string, unknown> | undefined;
    mixedScheduling.value = extra?.mixed_scheduling === true;
    allowOverages.value = extra?.allow_overages === true;
  };

  return {
    allowOverages,
    autoPauseOnExpired,
    expiresAtInput,
    form,
    hydrateFormStateFromAccount,
    mixedScheduling,
    statusOptions,
  };
}
