package domain

type SyncFromCRSInput struct {
	BaseURL            string
	Username           string
	Password           string
	SyncProxies        bool
	SelectedAccountIDs []string
}

type SyncFromCRSItemResult struct {
	CRSAccountID string `json:"crs_account_id"`
	Kind         string `json:"kind"`
	Name         string `json:"name"`
	Action       string `json:"action"`
	Error        string `json:"error,omitempty"`
}

type SyncFromCRSResult struct {
	Created int                     `json:"created"`
	Updated int                     `json:"updated"`
	Skipped int                     `json:"skipped"`
	Failed  int                     `json:"failed"`
	Items   []SyncFromCRSItemResult `json:"items"`
}

type PreviewFromCRSResult struct {
	NewAccounts      []CRSPreviewAccount `json:"new_accounts"`
	ExistingAccounts []CRSPreviewAccount `json:"existing_accounts"`
}

type CRSPreviewAccount struct {
	CRSAccountID string `json:"crs_account_id"`
	Kind         string `json:"kind"`
	Name         string `json:"name"`
	Platform     string `json:"platform"`
	Type         string `json:"type"`
}
