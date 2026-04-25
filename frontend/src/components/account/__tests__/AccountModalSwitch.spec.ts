import { describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";
import AccountModalSwitch from "../AccountModalSwitch.vue";

describe("AccountModalSwitch", () => {
  it("emits toggled boolean values", async () => {
    const wrapper = mount(AccountModalSwitch, {
      props: {
        modelValue: false,
        ariaLabel: "Toggle setting",
      },
    });

    await wrapper.find("button").trigger("click");

    expect(wrapper.emitted("update:modelValue")).toEqual([[true]]);
    expect(wrapper.find("button").attributes("aria-label")).toBe(
      "Toggle setting",
    );
    expect(wrapper.find("button").attributes("aria-pressed")).toBe("false");
  });

  it("renders enabled state", () => {
    const wrapper = mount(AccountModalSwitch, {
      props: {
        modelValue: true,
      },
    });

    expect(wrapper.find("button").classes()).toContain(
      "account-modal-switch--enabled",
    );
  });
});
