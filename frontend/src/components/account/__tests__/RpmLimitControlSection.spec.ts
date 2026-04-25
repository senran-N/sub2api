import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import RpmLimitControlSection from "../RpmLimitControlSection.vue";

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
  return mount(RpmLimitControlSection, {
    props: {
      enabled,
      baseRpm: null,
      strategy: "tiered",
      stickyBuffer: null,
      userMsgQueueMode: "",
      userMsgQueueModeOptions: [
        { value: "", label: "Off" },
        { value: "throttle", label: "Throttle" },
      ],
    },
  });
}

describe("RpmLimitControlSection", () => {
  it("keeps user message queue visible when RPM limit is disabled", async () => {
    const wrapper = mountSection(false);

    await wrapper.find("button").trigger("click");
    await wrapper
      .findAll("button")
      .find((button) => button.text() === "Throttle")!
      .trigger("click");

    expect(wrapper.text()).toContain(
      "admin.accounts.quotaControl.rpmLimit.label",
    );
    expect(wrapper.text()).toContain(
      "admin.accounts.quotaControl.rpmLimit.userMsgQueue",
    );
    expect(wrapper.text()).not.toContain(
      "admin.accounts.quotaControl.rpmLimit.baseRpm",
    );
    expect(wrapper.emitted("update:enabled")).toEqual([[true]]);
    expect(wrapper.emitted("update:userMsgQueueMode")).toEqual([["throttle"]]);
  });

  it("emits RPM limit field updates when enabled", async () => {
    const wrapper = mountSection(true);
    const inputs = wrapper.findAll("input");

    await inputs[0].setValue("120");
    await inputs[1].setValue("12");
    await wrapper
      .findAll("button")
      .find((button) =>
        button.text().includes(
          "admin.accounts.quotaControl.rpmLimit.strategyStickyExempt",
        ),
      )!
      .trigger("click");

    expect(wrapper.text()).toContain(
      "admin.accounts.quotaControl.rpmLimit.baseRpm",
    );
    expect(wrapper.emitted("update:baseRpm")).toEqual([[120]]);
    expect(wrapper.emitted("update:stickyBuffer")).toEqual([[12]]);
    expect(wrapper.emitted("update:strategy")).toEqual([["sticky_exempt"]]);
  });
});
