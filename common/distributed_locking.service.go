package common

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	DefaultExpiry  = 2 * time.Minute
	DefaultTimeout = 5 * time.Second
)

type IDistributedLockingService interface {
	Acquire(ctx context.Context, name string) (Lock, error)
	CustomeDurationAcquire(ctx context.Context, name string, expiresAt, timeout time.Duration) (Lock, error)
	Release(ctx context.Context, name string) error
	ReleaseMany(ctx context.Context, names ...string)
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

func (s *RedisLockService) Acquire(ctx context.Context, name string) (Lock, error) {
	return s.CustomeDurationAcquire(ctx, name, DefaultExpiry, DefaultTimeout)
}

func (s *RedisLockService) CustomeDurationAcquire(ctx context.Context, name string, expiresAt, timeout time.Duration) (Lock, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
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
		log.Printf("failed to acquire lock: %s", err.Error())
		ch <- Lock{}
		return
	}
	_, lockErr := s.client.SetEx(ctx, name, name, expiresAt).Result()
	if lockErr != nil {
		log.Printf("failed to set lock: %s", lockErr.Error())
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
			return errors.New("waiting for lock timed out")
		case <-time.After(1 * time.Second):
			continue
		}
	}
}

func (s *RedisLockService) tryAcquireLock(ctx context.Context, name string) error {
	value, err := s.client.Get(ctx, name).Result()
	if err == redis.Nil && value != name {
		return nil
	}
	return errors.New("failed to Acquire lock because it is already acquired")
}

func (s *RedisLockService) handleLockResult(lock Lock) (Lock, error) {
	if lock.Name == "" {
		return Lock{}, errors.New("lock is empty")
	}
	return lock, nil
}

func (s *RedisLockService) Release(ctx context.Context, name string) error {
	log.Printf("Releasing lock: %s", name)
	_, err := s.client.Del(ctx, name).Result()
	if err != nil {
		log.Printf("failed to Release lock: %s", err.Error())
		return errors.New("failed to Release lock")
	}
	return nil
}

func (s *RedisLockService) ReleaseMany(ctx context.Context, names ...string) {
	for _, name := range names {
		s.Release(ctx, name)
	}
}
