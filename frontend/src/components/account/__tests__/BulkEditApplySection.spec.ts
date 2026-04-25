import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import BulkEditApplySection from "../BulkEditApplySection.vue";

vi.mock("vue-i18n", async () => {
  const actual = await vi.importActual<typeof import("vue-i18n")>("vue-i18n");
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  };
});

describe("BulkEditApplySection", () => {
  it("renders translated copy, scopes the body, and emits enabled updates", async () => {
    const wrapper = mount(BulkEditApplySection, {
      props: {
        enabled: false,
        hintKey: "admin.accounts.bulkEdit.hint",
        id: "bulk-test",
        labelKey: "admin.accounts.bulkEdit.label",
      },
      slots: {
        default: "<button>inner</button>",
      },
    });

    await wrapper.find("input").setValue(true);

    expect(wrapper.text()).toContain("admin.accounts.bulkEdit.label");
    expect(wrapper.text()).toContain("admin.accounts.bulkEdit.hint");
    expect(wrapper.find("#bulk-test-body").classes()).toContain("opacity-50");
    expect(wrapper.emitted("update:enabled")).toEqual([[true]]);
  });
});
