package redis

import (
	"context"
	"github.com/certified-juniors/AtomHack/internal/domain"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type sessionRedisRepository struct {
	client *redis.Client
}

func NewSessionRedisRepository(client *redis.Client) domain.SessionRepository {
	return &sessionRedisRepository{client}
}

func (s *sessionRedisRepository) Add(session domain.Session) error {
	if session.Token == "" {
		return domain.ErrInvalidToken
	}

	duration := session.ExpiresAt.Sub(time.Now())
	err := s.client.Set(context.TODO(), session.Token, session.UserID, duration).Err()
	if err != nil {
		return err
	}

	return nil
}

func (s *sessionRedisRepository) AddCodeByID(id int, code string) error {
	if code == "" || id <= 0 {
		return domain.ErrBadRequest
	}

	err := s.client.Set(context.TODO(), strconv.Itoa(id), code, 24*time.Hour).Err()
	if err != nil {
		return err
	}

	return nil
}

func (s *sessionRedisRepository) DeleteByToken(token string) error {
	if token == "" {
		return domain.ErrInvalidToken
	}

	err := s.client.Del(context.Background(), token).Err()
	if err != nil {
		return err
	}

	return nil
}

func (s *sessionRedisRepository) GetUserID(token string) (string, error) {
	if token == "" {
		return "", domain.ErrInvalidToken
	}

	id, err := s.client.Get(context.Background(), token).Result()
	if err != nil {
		return "", err
	}

	return id, nil
}

func (s *sessionRedisRepository) GetCodeByID(id string) (string, error) {
	if id == "" {
		return "", domain.ErrInvalidToken
	}

	id, err := s.client.Get(context.Background(), id).Result()
	if err != nil {
		return "", err
	}

	return id, nil
}
