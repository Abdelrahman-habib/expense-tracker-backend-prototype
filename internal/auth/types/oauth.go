package types

// GoogleOauthToken represents a Google OAuth token with its associated data
type GoogleOauthToken struct {
	ExternalAccountID string   `json:"external_account_id"`
	Token             string   `json:"token"`
	Scopes            []string `json:"scopes"`
}
