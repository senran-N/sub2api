import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import CreateAccountSchedulingSection from "../CreateAccountSchedulingSection.vue";

vi.mock("vue-i18n", async () => {
  const actual = await vi.importActual<typeof import("vue-i18n")>("vue-i18n");
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  };
});

function mountSection(platform: "openai" | "antigravity" = "openai") {
  return mount(CreateAccountSchedulingSection, {
    props: {
      proxyId: null,
      proxies: [],
      concurrency: 1,
      loadFactor: null,
      priority: 1,
      rateMultiplier: 1,
      expiresAt: "",
      platform,
      mixedScheduling: false,
      allowOverages: false,
      groupIds: [],
      groups: [],
      simpleMode: false,
    },
    global: {
      stubs: {
        ProxySelector: {
          emits: ["update:modelValue"],
          template:
            '<button type="button" class="proxy-stub" @click="$emit(\'update:modelValue\', 3)">proxy</button>',
        },
        GroupSelector: {
          emits: ["update:modelValue"],
          template:
            '<button type="button" class="group-stub" @click="$emit(\'update:modelValue\', [2])">groups</button>',
        },
      },
    },
  });
}

describe("CreateAccountSchedulingSection", () => {
  it("emits scheduling field updates", async () => {
    const wrapper = mountSection();
    const inputs = wrapper.findAll("input");

    await wrapper.find(".proxy-stub").trigger("click");
    await inputs[0].setValue("0");
    await inputs[1].setValue("4");
    await inputs[2].setValue("9");
    await inputs[3].setValue("1.25");
    await inputs[4].setValue("2026-04-25T12:30");
    await wrapper.find(".group-stub").trigger("click");

    expect(wrapper.text()).toContain("admin.accounts.proxy");
    expect(wrapper.emitted("update:proxyId")).toEqual([[3]]);
    expect(wrapper.emitted("update:concurrency")).toEqual([[1]]);
    expect(wrapper.emitted("update:loadFactor")).toEqual([[4]]);
    expect(wrapper.emitted("update:priority")).toEqual([[9]]);
    expect(wrapper.emitted("update:rateMultiplier")).toEqual([[1.25]]);
    expect(wrapper.emitted("update:expiresAt")).toEqual([
      ["2026-04-25T12:30"],
    ]);
    expect(wrapper.emitted("update:groupIds")).toEqual([[[2]]]);
  });

  it("emits Antigravity option changes", async () => {
    const wrapper = mountSection("antigravity");
    const checkboxes = wrapper.findAll('input[type="checkbox"]');

    await checkboxes[0].setValue(true);
    await checkboxes[1].setValue(true);

    expect(wrapper.text()).toContain("admin.accounts.mixedScheduling");
    expect(wrapper.text()).toContain("admin.accounts.allowOverages");
    expect(wrapper.emitted("update:mixedScheduling")).toEqual([[true]]);
    expect(wrapper.emitted("update:allowOverages")).toEqual([[true]]);
  });
});
