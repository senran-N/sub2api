import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import SessionLimitControlSection from "../SessionLimitControlSection.vue";

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
  return mount(SessionLimitControlSection, {
    props: {
      enabled,
      maxSessions: null,
      idleTimeout: null,
    },
  });
}

describe("SessionLimitControlSection", () => {
  it("emits enabled changes from the shared quota card", async () => {
    const wrapper = mountSection();

    await wrapper.find("button").trigger("click");

    expect(wrapper.text()).toContain(
      "admin.accounts.quotaControl.sessionLimit.label",
    );
    expect(wrapper.text()).toContain(
      "admin.accounts.quotaControl.sessionLimit.hint",
    );
    expect(wrapper.emitted("update:enabled")).toEqual([[true]]);
  });

  it("emits session limit numeric updates", async () => {
    const wrapper = mountSection(true);
    const inputs = wrapper.findAll("input");

    await inputs[0].setValue("8");
    await inputs[1].setValue("45");

    expect(wrapper.text()).toContain(
      "admin.accounts.quotaControl.sessionLimit.maxSessions",
    );
    expect(wrapper.text()).toContain("common.minutes");
    expect(wrapper.emitted("update:maxSessions")).toEqual([[8]]);
    expect(wrapper.emitted("update:idleTimeout")).toEqual([[45]]);
  });
});
