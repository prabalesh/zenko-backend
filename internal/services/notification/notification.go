package notification

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/prabalesh/zenko-backend/internal/db/sqlc"
	"github.com/prabalesh/zenko-backend/internal/pkg/dto"
	"github.com/prabalesh/zenko-backend/internal/pkg/errors"
)

type NotificationService interface {
	GetPreferences(ctx context.Context, userID string) (*dto.NotifPrefsResp, error)
	UpdatePreferences(ctx context.Context, userID string, req *dto.UpdateNotifPrefsReq) error
	RegisterFCMToken(ctx context.Context, userID string, req *dto.RegisterFCMReq) error
	GetNotifications(ctx context.Context, userID string, cursor *string, limit int) ([]dto.NotificationResp, error)
	MarkAsRead(ctx context.Context, userID, notifID string) error
	MarkAllAsRead(ctx context.Context, userID string) error
}

type notificationService struct {
	db *sqlc.Queries
}

func NewNotificationService(db *sqlc.Queries) NotificationService {
	return &notificationService{
		db: db,
	}
}

func (s *notificationService) GetPreferences(ctx context.Context, userID string) (*dto.NotifPrefsResp, error) {
	uUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.BadRequest("invalid user id")
	}

	prefs, err := s.db.GetNotificationPreferences(ctx, pgtype.UUID{Bytes: uUUID, Valid: true})
	if err != nil {
		if err == pgx.ErrNoRows {
			// Return default preferences
			return &dto.NotifPrefsResp{
				FriendRequest:     true,
				FriendAccepted:    true,
				ChallengeReceived: true,
				ChallengeDeclined: true,
				Reengagement:      true,
				WeeklyReset:       true,
				GlobalMute:        false,
			}, nil
		}
		return nil, errors.Internal("failed to fetch notification preferences")
	}

	return &dto.NotifPrefsResp{
		FriendRequest:     prefs.FriendRequest,
		FriendAccepted:    prefs.FriendAccepted,
		ChallengeReceived: prefs.ChallengeReceived,
		ChallengeDeclined: prefs.ChallengeDeclined,
		Reengagement:      prefs.Reengagement,
		WeeklyReset:       prefs.WeeklyReset,
		GlobalMute:        prefs.GlobalMute,
	}, nil
}

func (s *notificationService) UpdatePreferences(ctx context.Context, userID string, req *dto.UpdateNotifPrefsReq) error {
	uUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.BadRequest("invalid user id")
	}

	params := sqlc.UpsertNotificationPreferencesParams{
		UserID: pgtype.UUID{Bytes: uUUID, Valid: true},
	}

	if req.FriendRequest != nil {
		params.FriendRequest = pgtype.Bool{Bool: *req.FriendRequest, Valid: true}
	}
	if req.FriendAccepted != nil {
		params.FriendAccepted = pgtype.Bool{Bool: *req.FriendAccepted, Valid: true}
	}
	if req.ChallengeReceived != nil {
		params.ChallengeReceived = pgtype.Bool{Bool: *req.ChallengeReceived, Valid: true}
	}
	if req.ChallengeDeclined != nil {
		params.ChallengeDeclined = pgtype.Bool{Bool: *req.ChallengeDeclined, Valid: true}
	}
	if req.Reengagement != nil {
		params.Reengagement = pgtype.Bool{Bool: *req.Reengagement, Valid: true}
	}
	if req.WeeklyReset != nil {
		params.WeeklyReset = pgtype.Bool{Bool: *req.WeeklyReset, Valid: true}
	}
	if req.GlobalMute != nil {
		params.GlobalMute = pgtype.Bool{Bool: *req.GlobalMute, Valid: true}
	}

	_, err = s.db.UpsertNotificationPreferences(ctx, params)
	if err != nil {
		return errors.Internal("failed to update notification preferences")
	}

	return nil
}

func (s *notificationService) RegisterFCMToken(ctx context.Context, userID string, req *dto.RegisterFCMReq) error {
	uUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.BadRequest("invalid user id")
	}

	err = s.db.RegisterFCMToken(ctx, sqlc.RegisterFCMTokenParams{
		UserID:   pgtype.UUID{Bytes: uUUID, Valid: true},
		FcmToken: req.Token,
		Platform: sqlc.PlatformType(req.Platform),
	})
	if err != nil {
		return errors.Internal("failed to register FCM token")
	}

	return nil
}

func (s *notificationService) GetNotifications(ctx context.Context, userID string, cursor *string, limit int) ([]dto.NotificationResp, error) {
	uUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.BadRequest("invalid user id")
	}

	cUUID := pgtype.UUID{Valid: false}
	if cursor != nil && *cursor != "" {
		if parsed, err := uuid.Parse(*cursor); err == nil {
			cUUID = pgtype.UUID{Bytes: parsed, Valid: true}
		}
	}

	rows, err := s.db.ListNotificationsPaginated(ctx, sqlc.ListNotificationsPaginatedParams{
		UserID: pgtype.UUID{Bytes: uUUID, Valid: true},
		ID:     cUUID,
		Limit:  int32(limit),
	})
	if err != nil {
		return nil, errors.Internal("failed to list notifications")
	}

	res := make([]dto.NotificationResp, 0, len(rows))
	for _, r := range rows {
		res = append(res, dto.NotificationResp{
			ID:        uuid.UUID(r.ID.Bytes).String(),
			Type:      string(r.Type),
			Title:     r.Title,
			Body:      r.Body,
			Data:      r.Data,
			IsRead:    r.IsRead,
			CreatedAt: r.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return res, nil
}

func (s *notificationService) MarkAsRead(ctx context.Context, userID, notifID string) error {
	uUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.BadRequest("invalid user id")
	}
	nUUID, err := uuid.Parse(notifID)
	if err != nil {
		return errors.BadRequest("invalid notification id")
	}

	err = s.db.MarkNotificationRead(ctx, sqlc.MarkNotificationReadParams{
		ID:     pgtype.UUID{Bytes: nUUID, Valid: true},
		UserID: pgtype.UUID{Bytes: uUUID, Valid: true},
	})
	if err != nil {
		return errors.Internal("failed to mark notification as read")
	}

	return nil
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, userID string) error {
	uUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.BadRequest("invalid user id")
	}

	err = s.db.MarkAllNotificationsRead(ctx, pgtype.UUID{Bytes: uUUID, Valid: true})
	if err != nil {
		return errors.Internal("failed to mark all notifications as read")
	}

	return nil
}
