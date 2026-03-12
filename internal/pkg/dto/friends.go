package dto

type SendFriendReqReq struct {
	Username string `json:"username" validate:"required"`
}

type FriendResp struct {
	ID           string  `json:"id"`
	Username     string  `json:"username"`
	AvatarURL    string  `json:"avatar_url"`
	Elo          int     `json:"elo"`
	WinRate      float64 `json:"win_rate"`
	OnlineStatus bool    `json:"online_status"`
	Status       string  `json:"status"` // accepted|pending|blocked
}

type FriendListResp struct {
	Friends    []FriendResp `json:"friends"`
	TotalCount int          `json:"total_count"`
}

type FriendRequestsResp struct {
	Incoming []FriendResp `json:"incoming"`
	Outgoing []FriendResp `json:"outgoing"`
}
