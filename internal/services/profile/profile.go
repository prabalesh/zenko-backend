package profile

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/microcosm-cc/bluemonday"
	"github.com/prabalesh/zenko-backend/internal/db/sqlc"
	"github.com/prabalesh/zenko-backend/internal/pkg/dto"
	"github.com/prabalesh/zenko-backend/internal/pkg/errors"
)

type ProfileService interface {
	SetUsername(ctx context.Context, userID, username string) error
	GetProfile(ctx context.Context, userID string) (*dto.ProfileResp, error)
	GetPublicProfile(ctx context.Context, username string) (*dto.ProfileResp, error)
	UpdateProfile(ctx context.Context, userID string, req *dto.UpdateProfileReq) error
	ChangeUsername(ctx context.Context, userID, newUsername string) error
	UpdateSocialLinks(ctx context.Context, userID string, req *dto.UpdateSocialLinksReq) error
}

type profileService struct {
	db        *sqlc.Queries
	sanitizer *bluemonday.Policy
}

func NewProfileService(db *sqlc.Queries) ProfileService {
	return &profileService{
		db:        db,
		sanitizer: bluemonday.StrictPolicy(), // Strip all HTML tags
	}
}

func (s *profileService) SetUsername(ctx context.Context, userID, newUsername string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.BadRequest("invalid user id")
	}

	_, err = s.db.GetUserByID(ctx, pgtype.UUID{Bytes: userUUID, Valid: true})
	if err != nil {
		if err == pgx.ErrNoRows {
			return errors.Unauthorized("user not found")
		}
		return errors.Internal("failed to fetch user")
	}

	// In a real flow, you might check if they already set their initial username
	// For this test, we just do a direct update
	_, err = s.db.UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:       pgtype.UUID{Bytes: userUUID, Valid: true},
		Username: pgtype.Text{String: newUsername, Valid: true},
	})
	if err != nil {
		// e.g. unique constraint violation on username
		return errors.BadRequest("failed to set username, it might be taken")
	}

	return nil
}

func (s *profileService) mapToProfileResp(user sqlc.User, links []sqlc.UserSocialLink) *dto.ProfileResp {
	resp := &dto.ProfileResp{
		ID:            uuid.UUID(user.ID.Bytes).String(),
		Username:      user.Username,
		AvatarURL:     user.AvatarUrl,
		Bio:           user.Bio.String,
		Country:       user.Country.String,
		Elo:           int(user.Elo),
		Wins:          int(user.Wins),
		Losses:        int(user.Losses),
		MatchesPlayed: int(user.MatchesPlayed),
		BestStreak:    int(user.BestStreak),
		CurrentStreak: int(user.CurrentStreak),
		XP:            int(user.Xp),
		CreatedAt:     user.CreatedAt.Time.Format(time.RFC3339),
	}

	if user.FavMode.Valid {
		resp.FavMode = string(user.FavMode.GameMode)
	}

	if resp.MatchesPlayed > 0 {
		resp.WinRate = float64(resp.Wins) / float64(resp.MatchesPlayed)
	}

	resp.SocialLinks = make([]dto.SocialLinkItem, 0, len(links))
	for _, l := range links {
		resp.SocialLinks = append(resp.SocialLinks, dto.SocialLinkItem{
			Platform: string(l.Platform),
			URL:      l.Url,
		})
	}

	return resp
}

func (s *profileService) GetProfile(ctx context.Context, userID string) (*dto.ProfileResp, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.BadRequest("invalid user id")
	}

	uuidParams := pgtype.UUID{Bytes: userUUID, Valid: true}

	user, err := s.db.GetUserByID(ctx, uuidParams)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.Unauthorized("user not found")
		}
		return nil, errors.Internal("failed to fetch user")
	}

	links, err := s.db.GetSocialLinksByUserID(ctx, uuidParams)
	if err != nil && err != pgx.ErrNoRows {
		return nil, errors.Internal("failed to fetch social links")
	}

	return s.mapToProfileResp(user, links), nil
}

func (s *profileService) GetPublicProfile(ctx context.Context, username string) (*dto.ProfileResp, error) {
	user, err := s.db.GetUserByUsername(ctx, username)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.BadRequest("user not found")
		}
		return nil, errors.Internal("failed to fetch user")
	}

	links, err := s.db.GetSocialLinksByUserID(ctx, user.ID)
	if err != nil && err != pgx.ErrNoRows {
		return nil, errors.Internal("failed to fetch social links")
	}

	return s.mapToProfileResp(user, links), nil
}

func (s *profileService) UpdateProfile(ctx context.Context, userID string, req *dto.UpdateProfileReq) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.BadRequest("invalid user id")
	}

	params := sqlc.UpdateUserParams{
		ID: pgtype.UUID{Bytes: userUUID, Valid: true},
	}

	if req.Bio != nil {
		cleanBio := s.sanitizer.Sanitize(*req.Bio)
		params.Bio = pgtype.Text{String: cleanBio, Valid: true}
	}
	if req.Country != nil {
		params.Country = pgtype.Text{String: *req.Country, Valid: true}
	}
	if req.Dob != nil {
		t, err := time.Parse("2006-01-02", *req.Dob)
		if err == nil {
			params.Dob = pgtype.Date{Time: t, Valid: true}
		}
	}

	_, err = s.db.UpdateUser(ctx, params)
	if err != nil {
		return errors.Internal("failed to update profile")
	}

	return nil
}

func (s *profileService) ChangeUsername(ctx context.Context, userID, newUsername string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.BadRequest("invalid user id")
	}
	uuidParams := pgtype.UUID{Bytes: userUUID, Valid: true}

	// 1. Fetch user to get old username
	user, err := s.db.GetUserByID(ctx, uuidParams)
	if err != nil {
		return errors.Internal("user not found")
	}

	// 2. Enforce logic: max 2 changes per 30 days
	count, err := s.db.CountUsernameChangesPast30Days(ctx, uuidParams)
	if err != nil {
		return errors.Internal("failed to verify username change config")
	}

	if count >= 2 {
		// Find when they can next change it (based on the earliest of the 2 recent changes)
		// For simplicity, just read the latest and add 30 days logic, or just return 429
		return errors.BadRequest(fmt.Sprintf("You have reached the limit of 2 username changes per 30 days."))
	}

	// 3. Update User
	_, err = s.db.UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:       uuidParams,
		Username: pgtype.Text{String: newUsername, Valid: true},
	})
	if err != nil {
		return errors.BadRequest("username update failed, might be already taken")
	}

	// 4. Log change
	err = s.db.InsertUsernameChange(ctx, sqlc.InsertUsernameChangeParams{
		UserID:      uuidParams,
		OldUsername: user.Username,
		NewUsername: newUsername,
	})
	if err != nil {
		return errors.Internal("failed to log username change") // should ideally be transacted
	}

	return nil
}

func (s *profileService) UpdateSocialLinks(ctx context.Context, userID string, req *dto.UpdateSocialLinksReq) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.BadRequest("invalid user id")
	}
	uuidParams := pgtype.UUID{Bytes: userUUID, Valid: true}

	// simple upsert loop (could be transacted for purity)
	for _, link := range req.Links {
		// Only upsert if URL is not empty, if empty we could delete explicitly though no requirement yet
		_, err := s.db.UpsertSocialLink(ctx, sqlc.UpsertSocialLinkParams{
			UserID:   uuidParams,
			Platform: sqlc.SocialPlatform(link.Platform),
			Url:      link.URL,
		})
		if err != nil {
			return errors.Internal("failed to update social links")
		}
	}
	return nil
}
