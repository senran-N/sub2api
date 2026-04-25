import { describe, expect, it, vi } from "vitest";
import { ref } from "vue";
import { useCreateMixedChannelWarning } from "../useCreateMixedChannelWarning";
import type { AccountPlatform } from "@/types";

const t = (key: string, values?: Record<string, unknown>) => {
  if (!values) {
    return key;
  }
  return `${key}:${values.groupName}`;
};

function createWarning(
  platform: AccountPlatform = "antigravity",
  checkMixedChannelRisk = vi.fn().mockResolvedValue({ has_risk: false }),
) {
  const active = ref(true);
  const selectedPlatform = ref<AccountPlatform>(platform);
  const groupIds = ref([1, 2]);
  const showError = vi.fn();
  const warning = useCreateMixedChannelWarning<number>({
    getGroupIds: () => groupIds.value,
    getPlatform: () => selectedPlatform.value,
    isActiveRequest: () => active.value,
    resolveErrorMessage: (error) =>
      (error as { message?: string }).message || "failed",
    showError,
    t,
    checkMixedChannelRisk,
  });

  return {
    active,
    checkMixedChannelRisk,
    groupIds,
    selectedPlatform,
    showError,
    warning,
  };
}

describe("useCreateMixedChannelWarning", () => {
  it("checks risk for mixed-channel create platforms", async () => {
    const { checkMixedChannelRisk, warning } = createWarning();
    const onConfirm = vi.fn();

    const canContinue = await warning.ensureMixedChannelConfirmed(onConfirm, 1);

    expect(canContinue).toBe(true);
    expect(checkMixedChannelRisk).toHaveBeenCalledWith({
      platform: "antigravity",
      group_ids: [1, 2],
    });
    expect(onConfirm).not.toHaveBeenCalled();
    expect(warning.showMixedChannelWarning.value).toBe(false);
  });

  it("opens a warning and marks subsequent create payloads confirmed", async () => {
    const onConfirm = vi.fn().mockResolvedValue(undefined);
    const { warning } = createWarning(
      "anthropic",
      vi.fn().mockResolvedValue({
        has_risk: true,
        details: {
          group_id: 1,
          group_name: "Mixed Group",
          current_platform: "anthropic",
          other_platform: "antigravity",
        },
      }),
    );

    const canContinue = await warning.ensureMixedChannelConfirmed(onConfirm, 1);

    expect(canContinue).toBe(false);
    expect(warning.showMixedChannelWarning.value).toBe(true);
    expect(warning.mixedChannelWarningMessageText.value).toBe(
      "admin.accounts.mixedChannelWarning:Mixed Group",
    );

    const action = warning.takeMixedChannelWarningAction();
    await action?.();

    expect(onConfirm).toHaveBeenCalledTimes(1);
    expect(
      warning.withMixedChannelConfirmFlag({
        name: "Account",
        platform: "anthropic",
        type: "oauth",
        credentials: {},
      }),
    ).toMatchObject({
      confirm_mixed_channel_risk: true,
    });
  });

  it("opens backend conflict warnings with the backend message", () => {
    const { warning } = createWarning();
    const onConfirm = vi.fn().mockResolvedValue(undefined);

    const handled = warning.openMixedChannelConflictDialog(
      {
        response: {
          status: 409,
          data: {
            error: "mixed_channel_warning",
            message: "backend warning",
          },
        },
      },
      1,
      onConfirm,
    );

    expect(handled).toBe(true);
    expect(warning.showMixedChannelWarning.value).toBe(true);
    expect(warning.mixedChannelWarningMessageText.value).toBe(
      "backend warning",
    );
  });

  it("skips unrelated platforms and strips stale confirmation flags", async () => {
    const { checkMixedChannelRisk, warning } = createWarning("openai");

    const canContinue = await warning.ensureMixedChannelConfirmed(vi.fn(), 1);

    expect(canContinue).toBe(true);
    expect(checkMixedChannelRisk).not.toHaveBeenCalled();
    expect(
      warning.withMixedChannelConfirmFlag({
        name: "Account",
        platform: "openai",
        type: "oauth",
        credentials: {},
        confirm_mixed_channel_risk: true,
      }),
    ).toEqual({
      name: "Account",
      platform: "openai",
      type: "oauth",
      credentials: {},
    });
  });
});
