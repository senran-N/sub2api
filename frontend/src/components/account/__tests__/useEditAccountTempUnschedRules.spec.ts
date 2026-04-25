import { describe, expect, it } from "vitest";
import { useEditAccountTempUnschedRules } from "../useEditAccountTempUnschedRules";

describe("useEditAccountTempUnschedRules", () => {
  it("hydrates enabled rules from credentials", () => {
    const rules = useEditAccountTempUnschedRules();

    rules.hydrateTempUnschedRulesFromCredentials({
      temp_unschedulable_enabled: true,
      temp_unschedulable_rules: [
        {
          error_code: 429,
          keywords: ["rate", "limited"],
          duration_minutes: 5,
          description: "rate limit",
        },
      ],
    });

    expect(rules.tempUnschedEnabled.value).toBe(true);
    expect(rules.tempUnschedRules.value).toEqual([
      {
        error_code: 429,
        keywords: "rate, limited",
        duration_minutes: 5,
        description: "rate limit",
      },
    ]);
  });

  it("adds, updates, moves, and removes rules", () => {
    const rules = useEditAccountTempUnschedRules();

    rules.addTempUnschedRule({
      error_code: 429,
      keywords: "first",
      duration_minutes: 1,
      description: "",
    });
    rules.addTempUnschedRule({
      error_code: 500,
      keywords: "second",
      duration_minutes: 2,
      description: "",
    });
    rules.updateTempUnschedRule(0, "keywords", "updated");
    rules.updateTempUnschedRule(0, "error_code", 503);
    rules.moveTempUnschedRule(0, 1);

    expect(rules.tempUnschedRules.value).toEqual([
      {
        error_code: 500,
        keywords: "second",
        duration_minutes: 2,
        description: "",
      },
      {
        error_code: 503,
        keywords: "updated",
        duration_minutes: 1,
        description: "",
      },
    ]);

    rules.removeTempUnschedRule(0);
    expect(rules.tempUnschedRules.value).toEqual([
      {
        error_code: 503,
        keywords: "updated",
        duration_minutes: 1,
        description: "",
      },
    ]);
  });

  it("resets rule state", () => {
    const rules = useEditAccountTempUnschedRules();

    rules.hydrateTempUnschedRulesFromCredentials({
      temp_unschedulable_enabled: true,
      temp_unschedulable_rules: [{ keywords: "rate" }],
    });
    rules.resetTempUnschedRules();

    expect(rules.tempUnschedEnabled.value).toBe(false);
    expect(rules.tempUnschedRules.value).toEqual([]);
  });
});
