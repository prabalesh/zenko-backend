package dto

type NotifPrefsResp struct {
	FriendRequest     bool `json:"friend_request"`
	FriendAccepted    bool `json:"friend_accepted"`
	ChallengeReceived bool `json:"challenge_received"`
	ChallengeDeclined bool `json:"challenge_declined"`
	Reengagement      bool `json:"reengagement"`
	WeeklyReset       bool `json:"weekly_reset"`
	GlobalMute        bool `json:"global_mute"`
}

type UpdateNotifPrefsReq struct {
	FriendRequest     *bool `json:"friend_request"`
	FriendAccepted    *bool `json:"friend_accepted"`
	ChallengeReceived *bool `json:"challenge_received"`
	ChallengeDeclined *bool `json:"challenge_declined"`
	Reengagement      *bool `json:"reengagement"`
	WeeklyReset       *bool `json:"weekly_reset"`
	GlobalMute        *bool `json:"global_mute"`
}

type RegisterFCMReq struct {
	Token    string `json:"token" validate:"required"`
	Platform string `json:"platform" validate:"required,oneof=ios android"`
}

type NotificationResp struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	Data      any    `json:"data"`
	IsRead    bool   `json:"is_read"`
	CreatedAt string `json:"created_at"`
}
