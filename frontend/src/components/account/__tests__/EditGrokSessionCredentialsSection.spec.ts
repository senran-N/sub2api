import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import EditGrokSessionCredentialsSection from "../EditGrokSessionCredentialsSection.vue";

vi.mock("vue-i18n", async () => {
  const actual = await vi.importActual<typeof import("vue-i18n")>("vue-i18n");
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  };
});

describe("EditGrokSessionCredentialsSection", () => {
  it("emits session token updates and shows leave-empty copy", async () => {
    const wrapper = mount(EditGrokSessionCredentialsSection, {
      props: {
        sessionToken: "",
      },
    });

    await wrapper.find("input").setValue("grok-session-token");

    expect(wrapper.text()).toContain("admin.accounts.grok.sessionToken");
    expect(wrapper.text()).toContain("admin.accounts.leaveEmptyToKeep");
    expect(wrapper.emitted("update:sessionToken")).toEqual([
      ["grok-session-token"],
    ]);
  });
});
