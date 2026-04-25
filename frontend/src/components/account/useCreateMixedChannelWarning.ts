import { computed, ref } from "vue";
import { adminAPI } from "@/api/admin";
import type {
  AccountPlatform,
  CheckMixedChannelResponse,
  CreateAccountRequest,
} from "@/types";
import {
  buildMixedChannelDetails,
  needsMixedChannelCheck,
  resolveMixedChannelWarningMessage,
  type Translate,
} from "@/components/account/accountModalShared";

interface UseCreateMixedChannelWarningOptions<TRequestContext> {
  getGroupIds: () => number[];
  getPlatform: () => AccountPlatform;
  isActiveRequest: (requestContext: TRequestContext) => boolean;
  resolveErrorMessage: (error: unknown) => string;
  showError: (message: string) => void;
  t: Translate;
  checkMixedChannelRisk?: typeof adminAPI.accounts.checkMixedChannelRisk;
}

interface OpenMixedChannelDialogOptions<TRequestContext> {
  message?: string;
  onConfirm: () => Promise<void>;
  requestContext: TRequestContext;
  response?: CheckMixedChannelResponse;
}

function isMixedChannelConflict(error: unknown): error is {
  response?: {
    data?: {
      message?: string;
    };
  };
} {
  const maybeError = error as
    | {
        response?: {
          status?: unknown;
          data?: {
            error?: unknown;
          };
        };
      }
    | null;
  return (
    maybeError?.response?.status === 409 &&
    maybeError.response.data?.error === "mixed_channel_warning"
  );
}

export function useCreateMixedChannelWarning<TRequestContext>(
  options: UseCreateMixedChannelWarningOptions<TRequestContext>,
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

  const currentPlatformNeedsCheck = () =>
    needsMixedChannelCheck(options.getPlatform());

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

  const openMixedChannelDialog = (
    dialog: OpenMixedChannelDialogOptions<TRequestContext>,
  ) => {
    mixedChannelWarningDetails.value = buildMixedChannelDetails(
      dialog.response,
    );
    mixedChannelWarningRawMessage.value =
      dialog.message ||
      dialog.response?.message ||
      options.t("admin.accounts.failedToCreate");
    mixedChannelWarningAction.value = async () => {
      if (!options.isActiveRequest(dialog.requestContext)) {
        return;
      }
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

  const withMixedChannelConfirmFlag = (
    payload: CreateAccountRequest,
  ): CreateAccountRequest => {
    if (needsMixedChannelCheck(payload.platform) && mixedChannelConfirmed.value) {
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
    requestContext: TRequestContext,
  ): Promise<boolean> => {
    if (!currentPlatformNeedsCheck()) {
      return true;
    }
    if (mixedChannelConfirmed.value) {
      return true;
    }

    try {
      const result = await checkMixedChannelRisk({
        platform: options.getPlatform(),
        group_ids: options.getGroupIds(),
      });
      if (!options.isActiveRequest(requestContext)) {
        return false;
      }
      if (!result.has_risk) {
        return true;
      }
      openMixedChannelDialog({
        response: result,
        requestContext,
        onConfirm,
      });
      return false;
    } catch (error: unknown) {
      if (!options.isActiveRequest(requestContext)) {
        return false;
      }
      options.showError(options.resolveErrorMessage(error));
      return false;
    }
  };

  const openMixedChannelConflictDialog = (
    error: unknown,
    requestContext: TRequestContext,
    onConfirm: () => Promise<void>,
  ) => {
    if (!isMixedChannelConflict(error) || !currentPlatformNeedsCheck()) {
      return false;
    }
    openMixedChannelDialog({
      message: error.response?.data?.message,
      requestContext,
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
