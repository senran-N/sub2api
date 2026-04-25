import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import QuotaLimitSection from "../QuotaLimitSection.vue";

vi.mock("vue-i18n", async () => {
  const actual = await vi.importActual<typeof import("vue-i18n")>("vue-i18n");
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  };
});

function mountSection() {
  return mount(QuotaLimitSection, {
    props: {
      dailyLimit: 20,
      dailyResetHour: null,
      dailyResetMode: null,
      notifyDailyEnabled: false,
      notifyDailyThreshold: null,
      notifyDailyThresholdType: null,
      notifyTotalEnabled: false,
      notifyTotalThreshold: null,
      notifyTotalThresholdType: null,
      notifyWeeklyEnabled: true,
      notifyWeeklyThreshold: 75,
      notifyWeeklyThresholdType: "percentage",
      resetTimezone: null,
      totalLimit: 100,
      weeklyLimit: null,
      weeklyResetDay: null,
      weeklyResetHour: null,
      weeklyResetMode: null,
    },
    global: {
      stubs: {
        QuotaLimitCard: {
          props: [
            "totalLimit",
            "dailyLimit",
            "weeklyLimit",
            "dailyResetMode",
            "dailyResetHour",
            "weeklyResetMode",
            "weeklyResetDay",
            "weeklyResetHour",
            "resetTimezone",
          ],
          emits: [
            "update:totalLimit",
            "update:dailyLimit",
            "update:weeklyLimit",
            "update:dailyResetMode",
            "update:dailyResetHour",
            "update:weeklyResetMode",
            "update:weeklyResetDay",
            "update:weeklyResetHour",
            "update:resetTimezone",
          ],
          template: `
            <section data-testid="quota-card">
              <button type="button" data-testid="total" @click="$emit('update:totalLimit', 200)">total</button>
              <button type="button" data-testid="daily" @click="$emit('update:dailyLimit', 40)">daily</button>
              <button type="button" data-testid="weekly" @click="$emit('update:weeklyLimit', 300)">weekly</button>
              <button type="button" data-testid="daily-mode" @click="$emit('update:dailyResetMode', 'fixed')">daily-mode</button>
              <button type="button" data-testid="daily-hour" @click="$emit('update:dailyResetHour', 8)">daily-hour</button>
              <button type="button" data-testid="weekly-mode" @click="$emit('update:weeklyResetMode', 'rolling')">weekly-mode</button>
              <button type="button" data-testid="weekly-day" @click="$emit('update:weeklyResetDay', 2)">weekly-day</button>
              <button type="button" data-testid="weekly-hour" @click="$emit('update:weeklyResetHour', 9)">weekly-hour</button>
              <button type="button" data-testid="timezone" @click="$emit('update:resetTimezone', 'UTC')">timezone</button>
            </section>
          `,
        },
        QuotaNotifyToggle: {
          props: ["enabled", "threshold", "thresholdType"],
          emits: ["update:enabled", "update:threshold", "update:thresholdType"],
          template: `
            <section class="notify-toggle">
              <button type="button" class="notify-enabled" @click="$emit('update:enabled', true)">enabled</button>
              <button type="button" class="notify-threshold" @click="$emit('update:threshold', 60)">threshold</button>
              <button type="button" class="notify-type" @click="$emit('update:thresholdType', 'fixed')">type</button>
            </section>
          `,
        },
      },
    },
  });
}

describe("QuotaLimitSection", () => {
  it("renders quota copy and forwards quota card updates", async () => {
    const wrapper = mountSection();

    await wrapper.find('[data-testid="total"]').trigger("click");
    await wrapper.find('[data-testid="daily"]').trigger("click");
    await wrapper.find('[data-testid="weekly"]').trigger("click");
    await wrapper.find('[data-testid="daily-mode"]').trigger("click");
    await wrapper.find('[data-testid="daily-hour"]').trigger("click");
    await wrapper.find('[data-testid="weekly-mode"]').trigger("click");
    await wrapper.find('[data-testid="weekly-day"]').trigger("click");
    await wrapper.find('[data-testid="weekly-hour"]').trigger("click");
    await wrapper.find('[data-testid="timezone"]').trigger("click");

    expect(wrapper.text()).toContain("admin.accounts.quotaLimit");
    expect(wrapper.text()).toContain("admin.accounts.quotaNotify.title");
    expect(wrapper.emitted("update:totalLimit")).toEqual([[200]]);
    expect(wrapper.emitted("update:dailyLimit")).toEqual([[40]]);
    expect(wrapper.emitted("update:weeklyLimit")).toEqual([[300]]);
    expect(wrapper.emitted("update:dailyResetMode")).toEqual([["fixed"]]);
    expect(wrapper.emitted("update:dailyResetHour")).toEqual([[8]]);
    expect(wrapper.emitted("update:weeklyResetMode")).toEqual([["rolling"]]);
    expect(wrapper.emitted("update:weeklyResetDay")).toEqual([[2]]);
    expect(wrapper.emitted("update:weeklyResetHour")).toEqual([[9]]);
    expect(wrapper.emitted("update:resetTimezone")).toEqual([["UTC"]]);
  });

  it("maps daily, weekly, and total notification updates by position", async () => {
    const wrapper = mountSection();
    const toggles = wrapper.findAll(".notify-toggle");

    await toggles[0].find(".notify-enabled").trigger("click");
    await toggles[0].find(".notify-threshold").trigger("click");
    await toggles[0].find(".notify-type").trigger("click");
    await toggles[1].find(".notify-enabled").trigger("click");
    await toggles[1].find(".notify-threshold").trigger("click");
    await toggles[1].find(".notify-type").trigger("click");
    await toggles[2].find(".notify-enabled").trigger("click");
    await toggles[2].find(".notify-threshold").trigger("click");
    await toggles[2].find(".notify-type").trigger("click");

    expect(wrapper.emitted("update:notifyDailyEnabled")).toEqual([[true]]);
    expect(wrapper.emitted("update:notifyDailyThreshold")).toEqual([[60]]);
    expect(wrapper.emitted("update:notifyDailyThresholdType")).toEqual([
      ["fixed"],
    ]);
    expect(wrapper.emitted("update:notifyWeeklyEnabled")).toEqual([[true]]);
    expect(wrapper.emitted("update:notifyWeeklyThreshold")).toEqual([[60]]);
    expect(wrapper.emitted("update:notifyWeeklyThresholdType")).toEqual([
      ["fixed"],
    ]);
    expect(wrapper.emitted("update:notifyTotalEnabled")).toEqual([[true]]);
    expect(wrapper.emitted("update:notifyTotalThreshold")).toEqual([[60]]);
    expect(wrapper.emitted("update:notifyTotalThresholdType")).toEqual([
      ["fixed"],
    ]);
  });
});
