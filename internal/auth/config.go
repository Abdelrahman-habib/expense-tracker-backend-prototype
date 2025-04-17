package auth

import "time"

// Config holds all authentication related configuration
type Config struct {
	// JWT configuration
	JWT JWTConfig `mapstructure:"jwt"`

	// OAuth providers configuration
	OAuth OAuthConfig `mapstructure:"oauth"`

	// Cookie configuration
	Cookie CookieConfig `mapstructure:"cookie"`
}

// JWTConfig holds JWT specific configuration
type JWTConfig struct {
	AccessTokenSecret  string        `mapstructure:"access_token_secret"`
	RefreshTokenSecret string        `mapstructure:"refresh_token_secret"`
	AccessTokenTTL     time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL    time.Duration `mapstructure:"refresh_token_ttl"`
}

// OAuthConfig holds configuration for OAuth providers
type OAuthConfig struct {
	Google GoogleConfig `mapstructure:"google"`
	GitHub GitHubConfig `mapstructure:"github"`
}

// GoogleConfig holds Google OAuth specific configuration
type GoogleConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url"`
}

// GitHubConfig holds GitHub OAuth specific configuration
type GitHubConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url"`
}

// CookieConfig holds configuration for secure cookies
type CookieConfig struct {
	Domain   string `mapstructure:"domain"`
	Path     string `mapstructure:"path"`
	Secure   bool   `mapstructure:"secure"`
	SameSite string `mapstructure:"same_site"`
}
