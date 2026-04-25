import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import AutoPauseOnExpiredSection from "../AutoPauseOnExpiredSection.vue";

vi.mock("vue-i18n", async () => {
  const actual = await vi.importActual<typeof import("vue-i18n")>("vue-i18n");
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  };
});

describe("AutoPauseOnExpiredSection", () => {
  it("renders expiration copy and emits enabled changes", async () => {
    const wrapper = mount(AutoPauseOnExpiredSection, {
      props: {
        enabled: true,
      },
    });

    await wrapper.find("button").trigger("click");

    expect(wrapper.text()).toContain("admin.accounts.autoPauseOnExpired");
    expect(wrapper.text()).toContain("admin.accounts.autoPauseOnExpiredDesc");
    expect(wrapper.emitted("update:enabled")).toEqual([[false]]);
  });
});
