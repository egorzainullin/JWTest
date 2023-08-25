package token

type AccessAndRefreshTokens struct {
	AccessToken  string `json:"refresh_token"`
	RefreshToken string `json:"access_token"`
}
