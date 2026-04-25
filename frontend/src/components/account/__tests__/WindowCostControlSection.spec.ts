import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import WindowCostControlSection from "../WindowCostControlSection.vue";

vi.mock("vue-i18n", async () => {
  const actual = await vi.importActual<typeof import("vue-i18n")>("vue-i18n");
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  };
});

function mountSection(enabled = false) {
  return mount(WindowCostControlSection, {
    props: {
      enabled,
      limit: null,
      stickyReserve: null,
    },
  });
}

describe("WindowCostControlSection", () => {
  it("emits enabled changes from the shared quota card", async () => {
    const wrapper = mountSection();

    await wrapper.find("button").trigger("click");

    expect(wrapper.text()).toContain(
      "admin.accounts.quotaControl.windowCost.label",
    );
    expect(wrapper.text()).toContain(
      "admin.accounts.quotaControl.windowCost.hint",
    );
    expect(wrapper.emitted("update:enabled")).toEqual([[true]]);
  });

  it("emits numeric limit updates", async () => {
    const wrapper = mountSection(true);
    const inputs = wrapper.findAll("input");

    await inputs[0].setValue("120");
    await inputs[1].setValue("20");

    expect(wrapper.text()).toContain(
      "admin.accounts.quotaControl.windowCost.limit",
    );
    expect(wrapper.emitted("update:limit")).toEqual([[120]]);
    expect(wrapper.emitted("update:stickyReserve")).toEqual([[20]]);
  });
});
