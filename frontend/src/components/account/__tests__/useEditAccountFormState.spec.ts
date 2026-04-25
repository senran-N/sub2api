import { describe, expect, it } from "vitest";
import { useEditAccountFormState } from "../useEditAccountFormState";
import type { Account } from "@/types";

function buildAccount(overrides: Partial<Account> = {}): Account {
  return {
    id: 1,
    name: "Account",
    notes: null,
    platform: "antigravity",
    type: "upstream",
    credentials: {},
    extra: {},
    proxy_id: null,
    concurrency: 4,
    load_factor: null,
    priority: 8,
    rate_multiplier: null,
    status: "error",
    group_ids: [1, 2],
    expires_at: null,
    auto_pause_on_expired: false,
    created_at: "",
    updated_at: "",
    ...overrides,
  } as Account;
}

describe("useEditAccountFormState", () => {
  const t = (key: string) => key;

  it("hydrates edit form, auto-pause, and mixed scheduling state", () => {
    const state = useEditAccountFormState(t);

    state.hydrateFormStateFromAccount(
      buildAccount({
        extra: {
          mixed_scheduling: true,
          allow_overages: true,
        },
        auto_pause_on_expired: true,
      }),
    );

    expect(state.form.name).toBe("Account");
    expect(state.form.notes).toBe("");
    expect(state.form.rate_multiplier).toBe(1);
    expect(state.form.status).toBe("error");
    expect(state.autoPauseOnExpired.value).toBe(true);
    expect(state.mixedScheduling.value).toBe(true);
    expect(state.allowOverages.value).toBe(true);
    expect(state.statusOptions.value).toEqual([
      { value: "active", label: "common.active" },
      { value: "inactive", label: "common.inactive" },
      { value: "error", label: "admin.accounts.status.error" },
    ]);
  });

  it("syncs datetime-local input with expires_at", () => {
    const state = useEditAccountFormState(t);

    state.expiresAtInput.value = "2026-01-02T03:04";

    expect(state.form.expires_at).toBeGreaterThan(0);
    expect(state.expiresAtInput.value).toBe("2026-01-02T03:04");
  });
});
