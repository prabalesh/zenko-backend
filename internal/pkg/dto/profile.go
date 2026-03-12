package dto

type SetUsernameReq struct {
	Username string `json:"username" validate:"required,min=3,max=30,alphanum"`
}

type UpdateProfileReq struct {
	Bio     *string `json:"bio" validate:"omitempty,max=150"`
	Country *string `json:"country" validate:"omitempty,len=2"`
	Dob     *string `json:"dob" validate:"omitempty,datetime=2006-01-02"`
}

type UpdateSocialLinksReq struct {
	Links []SocialLinkItem `json:"links" validate:"dive"`
}

type SocialLinkItem struct {
	Platform string `json:"platform" validate:"required,oneof=instagram twitter github linkedin youtube"`
	URL      string `json:"url" validate:"required,url,max=255"`
}

type ProfileResp struct {
	ID            string           `json:"id"`
	Username      string           `json:"username"`
	AvatarURL     string           `json:"avatar_url"`
	Bio           string           `json:"bio"`
	Country       string           `json:"country"`
	Elo           int              `json:"elo"`
	Wins          int              `json:"wins"`
	Losses        int              `json:"losses"`
	WinRate       float64          `json:"win_rate"`
	MatchesPlayed int              `json:"matches_played"`
	BestStreak    int              `json:"best_streak"`
	CurrentStreak int              `json:"current_streak"`
	XP            int              `json:"xp"`
	FavMode       string           `json:"fav_mode"`
	SocialLinks   []SocialLinkItem `json:"social_links"`
	CreatedAt     string           `json:"created_at"`
	// dob and is_bot are omitted
}

type ChangeUsernameReq struct {
	Username string `json:"username" validate:"required,min=3,max=30,alphanum"`
}
