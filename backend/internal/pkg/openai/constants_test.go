package openai

import "testing"

func TestDefaultModels_ContainsCurrentOpenAICatalog(t *testing.T) {
	t.Parallel()

	byID := make(map[string]Model, len(DefaultModels))
	for _, model := range DefaultModels {
		byID[model.ID] = model
	}

	expected := map[string]string{
		"gpt-5.5":             "GPT-5.5",
		"gpt-5.4":             "GPT-5.4",
		"gpt-5.4-mini":        "GPT-5.4 Mini",
		"gpt-5.3-codex":       "GPT-5.3 Codex",
		"gpt-5.3-codex-spark": "GPT-5.3 Codex Spark",
		"gpt-5.2":             "GPT-5.2",
	}

	ids := DefaultModelIDs()
	idsByValue := make(map[string]struct{}, len(ids))
	for _, modelID := range ids {
		idsByValue[modelID] = struct{}{}
	}

	for modelID, displayName := range expected {
		model, ok := byID[modelID]
		if !ok {
			t.Fatalf("expected default OpenAI models to expose %s", modelID)
		}
		if model.DisplayName != displayName {
			t.Fatalf("expected %s display name %q, got %q", modelID, displayName, model.DisplayName)
		}
		if _, ok := idsByValue[modelID]; !ok {
			t.Fatalf("expected DefaultModelIDs to include %s", modelID)
		}
	}
}
