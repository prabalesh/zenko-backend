package dto

type GoogleCallbackResp struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	IsNewUser    bool     `json:"is_new_user"`
	User         UserResp `json:"user"`
}

type RefreshReq struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RefreshResp struct {
	AccessToken string `json:"access_token"`
}

type UserResp struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
	Elo       int    `json:"elo"`
	IsNewUser bool   `json:"is_new_user"`
}
