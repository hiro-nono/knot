package ai

import "knot-api/internal/domain"

const structuredInformationSchemaName = "structured_information"

var sourceTypeValues = []string{
	string(domain.SourceTypeIdentity),
	string(domain.SourceTypeFact),
	string(domain.SourceTypeSchedule),
	string(domain.SourceTypeCondition),
	string(domain.SourceTypeInteraction),
}

var sourceStatusValues = []string{
	string(domain.SourceStatusConfirmed),
	string(domain.SourceStatusUndecided),
	string(domain.SourceStatusUnknown),
}

var sourceInteractionTypeValues = []string{
	string(domain.SourceInteractionTypeRadio),
	string(domain.SourceInteractionTypeCheck),
	string(domain.SourceInteractionTypeText),
}

// structuredInformationSchema はAIに構造化出力させるJSON Schemaを組み立てる。
// タイトルとSource一覧(domain.Sourceの各フィールドに対応)を持つ。
func structuredInformationSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"title", "sources"},
		"properties": map[string]any{
			"title": map[string]any{
				"type": "string",
			},
			"sources": map[string]any{
				"type":  "array",
				"items": sourceSchema(),
			},
		},
	}
}

func sourceSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"type", "key", "value", "status", "question", "interaction_type", "options"},
		"properties": map[string]any{
			"type": map[string]any{
				"type": "string",
				"enum": sourceTypeValues,
			},
			"key": map[string]any{
				"type": "string",
			},
			"value": map[string]any{
				"type": "string",
			},
			"status": map[string]any{
				"type": "string",
				"enum": sourceStatusValues,
			},
			"question": map[string]any{
				"type": []string{"string", "null"},
			},
			"interaction_type": map[string]any{
				"type": []string{"string", "null"},
				"enum": interactionTypeEnumValues(),
			},
			"options": map[string]any{
				"type":  "array",
				"items": optionSchema(),
			},
		},
	}
}

func optionSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"value", "sort_order"},
		"properties": map[string]any{
			"value": map[string]any{
				"type": "string",
			},
			"sort_order": map[string]any{
				"type": "integer",
			},
		},
	}
}

func interactionTypeEnumValues() []any {
	values := make([]any, 0, len(sourceInteractionTypeValues)+1)
	for _, v := range sourceInteractionTypeValues {
		values = append(values, v)
	}
	return append(values, nil)
}
