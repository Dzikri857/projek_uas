package model

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}
