package notification

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/prabalesh/zenko-backend/internal/db/sqlc"
)

type PushService interface {
	SendPush(ctx context.Context, userID, title, body, notifType string, data map[string]string) error
}

type pushService struct {
	db *sqlc.Queries
}

func NewPushService(db *sqlc.Queries) PushService {
	return &pushService{
		db: db,
	}
}

// SendPush sends an FCM notification if the user's preferences permit it.
func (s *pushService) SendPush(ctx context.Context, userID, title, body, notifType string, data map[string]string) error {
	uUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	uuidParam := pgtype.UUID{Bytes: uUUID, Valid: true}

	// 1. Check Preferences
	prefs, err := s.db.GetNotificationPreferences(ctx, uuidParam)
	if err != nil && err != pgx.ErrNoRows {
		return err
	}

	// Default behavior is to allow if no prefs row is found,
	// unless specifically suppressed via logic evaluation.
	if err == nil {
		if prefs.GlobalMute {
			return nil // Suppress notification silently
		}

		// Check specific type flags
		switch notifType {
		case "friend_request":
			if !prefs.FriendRequest {
				return nil
			}
		case "friend_accepted":
			if !prefs.FriendAccepted {
				return nil
			}
		case "challenge_in":
			if !prefs.ChallengeReceived {
				return nil
			}
		case "challenge_declined":
			if !prefs.ChallengeDeclined {
				return nil
			}
		case "reengagement":
			if !prefs.Reengagement {
				return nil
			}
		case "weekly_reset":
			if !prefs.WeeklyReset {
				return nil
			}
		}
	}

	// 2. Fetch Tokens
	tokens, err := s.db.GetUserFCMTokens(ctx, uuidParam)
	if err != nil || len(tokens) == 0 {
		return nil // No devices to notify
	}

	// 3. Send via FCM HTTP v1 API
	// In a real application, you would invoke the proper Google FCM SDK here.
	for _, token := range tokens {
		go func(t string) {
			// e.g., fcmClient.Send(ctx, message)
			// Mocked evaluation:
			log.Printf("Simulating FCM Push to token %s: [%s] %s\n", t, title, body)

			// If simulated FCM responds with Unregistered (404),
			// invalid := true
			// if invalid {
			//	 s.db.DeleteFCMToken(context.Background(), t)
			// }
		}(token)
	}

	return nil
}
