import { describe, expect, it, vi } from "vitest";
import { ref } from "vue";
import { useEditMixedChannelWarning } from "../useEditMixedChannelWarning";
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
  const account = ref({ id: 7, platform });
  const groupIds = ref([1, 2]);
  const showError = vi.fn();
  const warning = useEditMixedChannelWarning({
    getAccount: () => account.value,
    getGroupIds: () => groupIds.value,
    showError,
    t,
    checkMixedChannelRisk,
  });

  return {
    account,
    checkMixedChannelRisk,
    groupIds,
    showError,
    warning,
  };
}

describe("useEditMixedChannelWarning", () => {
  it("checks mixed-channel risk for editable mixed-channel platforms", async () => {
    const { checkMixedChannelRisk, warning } = createWarning();
    const onConfirm = vi.fn();

    const canContinue = await warning.ensureMixedChannelConfirmed(onConfirm);

    expect(canContinue).toBe(true);
    expect(checkMixedChannelRisk).toHaveBeenCalledWith({
      platform: "antigravity",
      group_ids: [1, 2],
      account_id: 7,
    });
    expect(onConfirm).not.toHaveBeenCalled();
    expect(warning.showMixedChannelWarning.value).toBe(false);
  });

  it("opens a details-based warning and confirms the next payload", async () => {
    const onConfirm = vi.fn().mockResolvedValue(undefined);
    const { warning } = createWarning(
      "anthropic",
      vi.fn().mockResolvedValue({
        has_risk: true,
        message: "raw warning",
        details: {
          group_id: 1,
          group_name: "Mixed Group",
          current_platform: "anthropic",
          other_platform: "antigravity",
        },
      }),
    );

    const canContinue = await warning.ensureMixedChannelConfirmed(onConfirm);

    expect(canContinue).toBe(false);
    expect(warning.showMixedChannelWarning.value).toBe(true);
    expect(warning.mixedChannelWarningMessageText.value).toBe(
      "admin.accounts.mixedChannelWarning:Mixed Group",
    );
    expect(
      warning.withMixedChannelConfirmFlag({
        name: "Account",
        confirm_mixed_channel_risk: true,
      }),
    ).toEqual({ name: "Account" });

    const action = warning.takeMixedChannelWarningAction();
    expect(warning.showMixedChannelWarning.value).toBe(false);

    await action?.();

    expect(onConfirm).toHaveBeenCalledTimes(1);
    expect(warning.withMixedChannelConfirmFlag({ name: "Account" })).toEqual({
      name: "Account",
      confirm_mixed_channel_risk: true,
    });
  });

  it("opens a conflict warning for backend mixed-channel conflicts", () => {
    const { warning } = createWarning();
    const onConfirm = vi.fn().mockResolvedValue(undefined);

    const handled = warning.openMixedChannelConflictDialog(
      {
        status: 409,
        error: "mixed_channel_warning",
        message: "backend warning",
      },
      onConfirm,
    );

    expect(handled).toBe(true);
    expect(warning.showMixedChannelWarning.value).toBe(true);
    expect(warning.mixedChannelWarningMessageText.value).toBe(
      "backend warning",
    );
  });

  it("skips checks and strips stale confirmation for unrelated platforms", async () => {
    const { checkMixedChannelRisk, warning } = createWarning("openai");
    const onConfirm = vi.fn();

    const canContinue = await warning.ensureMixedChannelConfirmed(onConfirm);

    expect(canContinue).toBe(true);
    expect(checkMixedChannelRisk).not.toHaveBeenCalled();
    expect(
      warning.withMixedChannelConfirmFlag({
        name: "Account",
        confirm_mixed_channel_risk: true,
      }),
    ).toEqual({ name: "Account" });
  });

  it("reports check failures through the provided error presenter", async () => {
    const { showError, warning } = createWarning(
      "antigravity",
      vi.fn().mockRejectedValue({
        response: {
          data: {
            detail: "backend detail",
          },
        },
      }),
    );

    const canContinue = await warning.ensureMixedChannelConfirmed(vi.fn());

    expect(canContinue).toBe(false);
    expect(showError).toHaveBeenCalledWith("backend detail");
  });
});
