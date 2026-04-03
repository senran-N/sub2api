package domain

// TempUnschedState represents a temporary unschedulable runtime state.
type TempUnschedState struct {
	UntilUnix       int64  `json:"until_unix"`
	TriggeredAtUnix int64  `json:"triggered_at_unix"`
	StatusCode      int    `json:"status_code"`
	MatchedKeyword  string `json:"matched_keyword"`
	RuleIndex       int    `json:"rule_index"`
	ErrorMessage    string `json:"error_message"`
}
