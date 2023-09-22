package common

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	DefaultDuration = 5 * time.Second
)

type IDistributedLockingService interface {
	Acquire(ctx context.Context, name string, expiresAt time.Duration) (Lock, error)
	Release(ctx context.Context, name string) error
}

type Lock struct {
	Name      string
	ExpiresAt time.Duration
}

type RedisLockService struct {
	client *redis.Client
}

func CreateNewRedisLockService(client *redis.Client) IDistributedLockingService {
	return &RedisLockService{
		client,
	}
}

func (s *RedisLockService) Acquire(ctx context.Context, name string, expiresAt time.Duration) (Lock, error) {
	ctx, cancel := context.WithTimeout(ctx, expiresAt)
	defer cancel()
	ch := make(chan Lock)
	go s.acquire(ctx, ch, name, expiresAt)
	select {
	case <-ctx.Done():
		return Lock{}, errors.New("failed to Acquire lock")
	case lock := <-ch:
		return s.handleLockResult(lock)
	}
}

func (s *RedisLockService) acquire(ctx context.Context, ch chan Lock, name string, expiresAt time.Duration) {
	defer ctx.Done()
	err := s.waitForLock(ctx, name)
	if err != nil {
		log.Printf("failed to Acquire lock: %s", err.Error())
		ch <- Lock{}
		return
	}
	cmd := s.client.SetEx(ctx, name, name, expiresAt)
	lockErr := cmd.Err()
	if lockErr != nil {
		log.Printf("failed to Acquire lock: %s", lockErr.Error())
		ch <- Lock{}
	}
	ch <- Lock{Name: name, ExpiresAt: expiresAt}
}

func (s *RedisLockService) waitForLock(ctx context.Context, name string) error {
	for {
		err := s.tryAcquireLock(ctx, name)
		if err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return errors.New("failed to Acquire lock")
		case <-time.After(1 * time.Second):
			continue
		}
	}
}

func (s *RedisLockService) tryAcquireLock(ctx context.Context, name string) error {
	cmd := s.client.Get(ctx, name)
	err := cmd.Err()
	if err == redis.Nil {
		return nil
	}
	return err
}

func (s *RedisLockService) handleLockResult(lock Lock) (Lock, error) {
	if lock.Name == "" {
		return Lock{}, errors.New("failed to Acquire lock")
	}
	return lock, nil
}

func (s *RedisLockService) Release(ctx context.Context, name string) error {
	return nil
}
