import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import PoolModeSection from "../PoolModeSection.vue";

vi.mock("vue-i18n", async () => {
  const actual = await vi.importActual<typeof import("vue-i18n")>("vue-i18n");
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, values?: Record<string, unknown>) =>
        values ? `${key}:${JSON.stringify(values)}` : key,
    }),
  };
});

describe("PoolModeSection", () => {
  it("emits enabled changes without owning parent state", async () => {
    const wrapper = mount(PoolModeSection, {
      props: {
        enabled: false,
        retryCount: 1,
      },
    });

    await wrapper.find("button").trigger("click");

    expect(wrapper.text()).toContain("admin.accounts.poolMode");
    expect(wrapper.text()).toContain("admin.accounts.poolModeHint");
    expect(wrapper.emitted("update:enabled")).toEqual([[true]]);
    expect(wrapper.find("input").exists()).toBe(false);
  });

  it("renders retry count controls and emits numeric updates when enabled", async () => {
    const wrapper = mount(PoolModeSection, {
      props: {
        enabled: true,
        retryCount: 2,
      },
    });

    await wrapper.find("input").setValue("4");

    expect(wrapper.text()).toContain("admin.accounts.poolModeInfo");
    expect(wrapper.text()).toContain("admin.accounts.poolModeRetryCount");
    expect(wrapper.emitted("update:retryCount")).toEqual([[4]]);
  });
});
