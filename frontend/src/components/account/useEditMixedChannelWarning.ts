import { computed, ref } from "vue";
import { adminAPI } from "@/api/admin";
import type {
  Account,
  CheckMixedChannelResponse,
  UpdateAccountRequest,
} from "@/types";
import { resolveRequestErrorMessage } from "@/utils/requestError";
import {
  buildMixedChannelDetails,
  needsMixedChannelCheck,
  resolveMixedChannelWarningMessage,
  type Translate,
} from "@/components/account/accountModalShared";

type MixedChannelAccount = Pick<Account, "id" | "platform">;

interface UseEditMixedChannelWarningOptions {
  getAccount: () => MixedChannelAccount | null | undefined;
  getGroupIds: () => number[];
  showError: (message: string) => void;
  t: Translate;
  checkMixedChannelRisk?: typeof adminAPI.accounts.checkMixedChannelRisk;
}

interface OpenMixedChannelDialogOptions {
  response?: CheckMixedChannelResponse;
  message?: string;
  onConfirm: () => Promise<void>;
}

function isMixedChannelConflict(error: unknown): error is { message?: string } {
  const maybeError = error as { status?: unknown; error?: unknown } | null;
  return (
    maybeError?.status === 409 &&
    maybeError.error === "mixed_channel_warning"
  );
}

export function useEditMixedChannelWarning(
  options: UseEditMixedChannelWarningOptions,
) {
  const checkMixedChannelRisk =
    options.checkMixedChannelRisk ?? adminAPI.accounts.checkMixedChannelRisk;
  const showMixedChannelWarning = ref(false);
  const mixedChannelWarningDetails = ref<ReturnType<
    typeof buildMixedChannelDetails
  > | null>(null);
  const mixedChannelWarningRawMessage = ref("");
  const mixedChannelWarningAction = ref<(() => Promise<void>) | null>(null);
  const mixedChannelConfirmed = ref(false);

  const mixedChannelWarningMessageText = computed(() =>
    resolveMixedChannelWarningMessage({
      details: mixedChannelWarningDetails.value,
      rawMessage: mixedChannelWarningRawMessage.value,
      t: options.t,
    }),
  );

  const currentAccountNeedsCheck = () => {
    const platform = options.getAccount()?.platform;
    return Boolean(platform && needsMixedChannelCheck(platform));
  };

  const resetMixedChannelDialog = () => {
    showMixedChannelWarning.value = false;
    mixedChannelWarningDetails.value = null;
    mixedChannelWarningRawMessage.value = "";
    mixedChannelWarningAction.value = null;
  };

  const resetMixedChannelState = () => {
    mixedChannelConfirmed.value = false;
    resetMixedChannelDialog();
  };

  const openMixedChannelDialog = (dialog: OpenMixedChannelDialogOptions) => {
    mixedChannelWarningDetails.value = buildMixedChannelDetails(
      dialog.response,
    );
    mixedChannelWarningRawMessage.value =
      dialog.message ||
      dialog.response?.message ||
      options.t("admin.accounts.failedToUpdate");
    mixedChannelWarningAction.value = async () => {
      mixedChannelConfirmed.value = true;
      await dialog.onConfirm();
    };
    showMixedChannelWarning.value = true;
  };

  const takeMixedChannelWarningAction = () => {
    const action = mixedChannelWarningAction.value;
    resetMixedChannelDialog();
    return action;
  };

  const withMixedChannelConfirmFlag = <T extends UpdateAccountRequest>(
    payload: T,
  ): T => {
    if (currentAccountNeedsCheck() && mixedChannelConfirmed.value) {
      return {
        ...payload,
        confirm_mixed_channel_risk: true,
      };
    }
    const cloned = { ...payload };
    delete cloned.confirm_mixed_channel_risk;
    return cloned;
  };

  const ensureMixedChannelConfirmed = async (
    onConfirm: () => Promise<void>,
  ): Promise<boolean> => {
    const account = options.getAccount();
    if (!account || !currentAccountNeedsCheck()) {
      return true;
    }
    if (mixedChannelConfirmed.value) {
      return true;
    }

    try {
      const result = await checkMixedChannelRisk({
        platform: account.platform,
        group_ids: options.getGroupIds(),
        account_id: account.id,
      });
      if (!result.has_risk) {
        return true;
      }
      openMixedChannelDialog({
        response: result,
        onConfirm,
      });
      return false;
    } catch (error: unknown) {
      options.showError(
        resolveRequestErrorMessage(
          error,
          options.t("admin.accounts.failedToUpdate"),
        ),
      );
      return false;
    }
  };

  const openMixedChannelConflictDialog = (
    error: unknown,
    onConfirm: () => Promise<void>,
  ) => {
    if (!isMixedChannelConflict(error) || !currentAccountNeedsCheck()) {
      return false;
    }
    openMixedChannelDialog({
      message: error.message,
      onConfirm,
    });
    return true;
  };

  return {
    ensureMixedChannelConfirmed,
    mixedChannelWarningMessageText,
    openMixedChannelConflictDialog,
    resetMixedChannelDialog,
    resetMixedChannelState,
    showMixedChannelWarning,
    takeMixedChannelWarningAction,
    withMixedChannelConfirmFlag,
  };
}
