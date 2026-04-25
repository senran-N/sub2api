import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import WarmupSection from "../WarmupSection.vue";

vi.mock("vue-i18n", async () => {
  const actual = await vi.importActual<typeof import("vue-i18n")>("vue-i18n");
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  };
});

describe("WarmupSection", () => {
  it("renders warmup copy and emits enabled changes", async () => {
    const wrapper = mount(WarmupSection, {
      props: {
        enabled: false,
      },
    });

    await wrapper.find("button").trigger("click");

    expect(wrapper.text()).toContain("admin.accounts.interceptWarmupRequests");
    expect(wrapper.text()).toContain(
      "admin.accounts.interceptWarmupRequestsDesc",
    );
    expect(wrapper.emitted("update:enabled")).toEqual([[true]]);
  });
});
