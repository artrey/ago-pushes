package tokens

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

type UserPushToken struct {
	UserID    int64
	PushToken string
	Created   int64
}

func (s *Service) GetToken(ctx context.Context, userID int64) (*UserPushToken, error) {
	token := &UserPushToken{
		UserID: userID,
	}

	err := s.pool.QueryRow(ctx, `
		SELECT pushToken, CAST(EXTRACT(EPOCH FROM created) AS INTEGER) FROM tokens WHERE userId = $1
	`, userID).Scan(&token.PushToken, &token.Created)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return token, nil
}

func (s *Service) Register(ctx context.Context, userID int64, pushToken string) error {
	_, err := s.pool.Exec(
		ctx,
		`INSERT INTO tokens(userId, pushToken) VALUES($1, $2)`,
		userID, pushToken,
	)
	return err
}
