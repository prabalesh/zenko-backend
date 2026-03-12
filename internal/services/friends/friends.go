package friends

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/prabalesh/zenko-backend/internal/db/sqlc"
	"github.com/prabalesh/zenko-backend/internal/pkg/dto"
	"github.com/prabalesh/zenko-backend/internal/pkg/errors"
)

type FriendsService interface {
	SendFriendRequest(ctx context.Context, senderID, receiverUsername string) error
	GetFriendsList(ctx context.Context, userID string, cursor *string, limit int) (*dto.FriendListResp, error)
	GetFriendRequests(ctx context.Context, userID string) (*dto.FriendRequestsResp, error)
	AcceptFriendRequest(ctx context.Context, userID, friendID string) error
	RemoveFriend(ctx context.Context, userID, friendID string) error
	BlockUser(ctx context.Context, userID, blockUserID string) error
	UnblockUser(ctx context.Context, userID, unblockUserID string) error
}

type friendsService struct {
	db *sqlc.Queries
}

func NewFriendsService(db *sqlc.Queries) FriendsService {
	return &friendsService{
		db: db,
	}
}

func (s *friendsService) SendFriendRequest(ctx context.Context, senderID, receiverUsername string) error {
	senderUUID, err := uuid.Parse(senderID)
	if err != nil {
		return errors.BadRequest("invalid sender id")
	}

	receiver, err := s.db.GetUserByUsername(ctx, receiverUsername)
	if err != nil {
		if err == pgx.ErrNoRows {
			return errors.BadRequest("receiver not found")
		}
		return errors.Internal("failed to fetch user")
	}

	if senderUUID == receiver.ID.Bytes {
		return errors.BadRequest("cannot send request to self")
	}

	sID := pgtype.UUID{Bytes: senderUUID, Valid: true}

	// 1. Check blocks or existing relations
	existing, err := s.db.GetFriendship(ctx, sqlc.GetFriendshipParams{
		SenderID:   sID,
		ReceiverID: receiver.ID,
	})
	if err != nil && err != pgx.ErrNoRows {
		return errors.Internal("failed to check friendship status")
	}
	if err == nil {
		if existing.Status == "blocked" {
			return errors.BadRequest("cannot interact with this user")
		}
		if existing.Status == "accepted" {
			return errors.BadRequest("already friends")
		}
		if existing.Status == "pending" {
			return errors.BadRequest("request already pending")
		}
	}

	// 2. Enforce 200 friend limit for sender
	friendCount, err := s.db.CountUserFriends(ctx, sID)
	if err != nil {
		return errors.Internal("failed to check friend count")
	}
	if friendCount >= 200 {
		return errors.BadRequest("friend limit (200) reached")
	}

	// 3. Insert pending request
	_, err = s.db.SendFriendRequest(ctx, sqlc.SendFriendRequestParams{
		SenderID:   sID,
		ReceiverID: receiver.ID,
	})
	if err != nil {
		return errors.Internal("failed to send friend request")
	}

	return nil
}

func (s *friendsService) AcceptFriendRequest(ctx context.Context, userID, friendID string) error {
	uUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.BadRequest("invalid user id")
	}
	fUUID, err := uuid.Parse(friendID)
	if err != nil {
		return errors.BadRequest("invalid friend id")
	}

	uID := pgtype.UUID{Bytes: uUUID, Valid: true}
	fID := pgtype.UUID{Bytes: fUUID, Valid: true}

	existing, err := s.db.GetFriendship(ctx, sqlc.GetFriendshipParams{
		SenderID:   uID,
		ReceiverID: fID,
	})
	if err != nil {
		return errors.BadRequest("request not found")
	}

	// Only receiver can accept an incoming request.
	if existing.ReceiverID.Bytes != uUUID {
		return errors.Unauthorized("not authorized to accept this request")
	}
	if existing.Status != "pending" {
		return errors.BadRequest("request is not pending")
	}

	// Check limits
	uCount, _ := s.db.CountUserFriends(ctx, uID)
	if uCount >= 200 {
		return errors.BadRequest("your friend limit is reached")
	}
	fCount, _ := s.db.CountUserFriends(ctx, fID)
	if fCount >= 200 {
		return errors.BadRequest("user's friend limit is reached")
	}

	err = s.db.UpdateFriendStatus(ctx, sqlc.UpdateFriendStatusParams{
		SenderID:   fID,
		ReceiverID: uID,
		Status:     "accepted",
	})
	if err != nil {
		return errors.Internal("failed to accept request")
	}

	return nil
}

func (s *friendsService) RemoveFriend(ctx context.Context, userID, friendID string) error {
	uUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.BadRequest("invalid ids")
	}
	fUUID, err := uuid.Parse(friendID)
	if err != nil {
		return errors.BadRequest("invalid ids")
	}

	err = s.db.DeleteFriendship(ctx, sqlc.DeleteFriendshipParams{
		SenderID:   pgtype.UUID{Bytes: uUUID, Valid: true},
		ReceiverID: pgtype.UUID{Bytes: fUUID, Valid: true},
	})
	if err != nil {
		return errors.Internal("failed to remove friend")
	}
	return nil
}

func (s *friendsService) BlockUser(ctx context.Context, userID, blockUserID string) error {
	uUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.BadRequest("invalid ids")
	}
	fUUID, err := uuid.Parse(blockUserID)
	if err != nil {
		return errors.BadRequest("invalid ids")
	}

	// Overwrite existing status with "blocked", marking uUUID as the sender of the block
	// Note: Schema doesn't dictate direction for 'blocked', but conventionally sender dictates the blocker
	// For simplicity, we just delete existing record and insert new blocked record
	s.db.DeleteFriendship(ctx, sqlc.DeleteFriendshipParams{
		SenderID:   pgtype.UUID{Bytes: uUUID, Valid: true},
		ReceiverID: pgtype.UUID{Bytes: fUUID, Valid: true},
	})

	_, err = s.db.SendFriendRequest(ctx, sqlc.SendFriendRequestParams{
		SenderID:   pgtype.UUID{Bytes: uUUID, Valid: true},
		ReceiverID: pgtype.UUID{Bytes: fUUID, Valid: true},
	})
	if err == nil {
		_ = s.db.UpdateFriendStatus(ctx, sqlc.UpdateFriendStatusParams{
			SenderID:   pgtype.UUID{Bytes: uUUID, Valid: true},
			ReceiverID: pgtype.UUID{Bytes: fUUID, Valid: true},
			Status:     "blocked",
		})
	}
	return nil
}

func (s *friendsService) UnblockUser(ctx context.Context, userID, unblockUserID string) error {
	return s.RemoveFriend(ctx, userID, unblockUserID) // Deleting the row removes the block implicitly
}

func (s *friendsService) GetFriendsList(ctx context.Context, userID string, cursor *string, limit int) (*dto.FriendListResp, error) {
	uUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.BadRequest("invalid user id")
	}
	uID := pgtype.UUID{Bytes: uUUID, Valid: true}

	cUUID := pgtype.UUID{Valid: false}
	if cursor != nil && *cursor != "" {
		if parsed, err := uuid.Parse(*cursor); err == nil {
			cUUID = pgtype.UUID{Bytes: parsed, Valid: true}
		}
	}

	rows, err := s.db.ListFriendsPaginated(ctx, sqlc.ListFriendsPaginatedParams{
		SenderID: uID,
		ID:       cUUID,
		Limit:    int32(limit),
	})
	if err != nil {
		return nil, errors.Internal("failed to list friends")
	}

	count, _ := s.db.CountUserFriends(ctx, uID)

	friends := make([]dto.FriendResp, 0, len(rows))
	for _, r := range rows {
		winRate := float64(0)
		if r.MatchesPlayed > 0 {
			winRate = float64(r.Wins) / float64(r.MatchesPlayed)
		}
		friends = append(friends, dto.FriendResp{
			ID:           uuid.UUID(r.ID.Bytes).String(),
			Username:     r.Username,
			AvatarURL:    r.AvatarUrl,
			Elo:          int(r.Elo),
			WinRate:      winRate,
			OnlineStatus: r.OnlineStatus,
			Status:       string(r.Status),
		})
	}

	return &dto.FriendListResp{
		Friends:    friends,
		TotalCount: int(count),
	}, nil
}

func (s *friendsService) GetFriendRequests(ctx context.Context, userID string) (*dto.FriendRequestsResp, error) {
	uUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.BadRequest("invalid user id")
	}
	uID := pgtype.UUID{Bytes: uUUID, Valid: true}

	rows, err := s.db.GetFriendRequestsList(ctx, uID)
	if err != nil {
		return nil, errors.Internal("failed to get friend requests")
	}

	incoming := make([]dto.FriendResp, 0)
	outgoing := make([]dto.FriendResp, 0)

	for _, r := range rows {
		winRate := float64(0)
		if r.MatchesPlayed > 0 {
			winRate = float64(r.Wins) / float64(r.MatchesPlayed)
		}
		f := dto.FriendResp{
			ID:           uuid.UUID(r.UserID.Bytes).String(),
			Username:     r.Username,
			AvatarURL:    r.AvatarUrl,
			Elo:          int(r.Elo),
			WinRate:      winRate,
			OnlineStatus: r.OnlineStatus,
			Status:       string(r.Status),
		}

		if r.ReceiverID.Bytes == uUUID {
			incoming = append(incoming, f)
		} else {
			outgoing = append(outgoing, f)
		}
	}

	return &dto.FriendRequestsResp{
		Incoming: incoming,
		Outgoing: outgoing,
	}, nil
}
